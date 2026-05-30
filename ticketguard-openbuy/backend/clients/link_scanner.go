package clients

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ticketguard/backend/domain"
)

type ScanResult struct {
	Status  domain.ScanStatus `json:"status"`
	Verdict string            `json:"verdict"`
}

type LinkScanner interface {
	ScanURL(ctx context.Context, rawURL string) (ScanResult, error)
}

type MockLinkScanner struct{}

func NewLinkScanner(apiKey string) LinkScanner {
	if strings.TrimSpace(apiKey) != "" {
		return &VirusTotalScanner{apiKey: apiKey, client: &http.Client{Timeout: 10 * time.Second}}
	}
	return &MockLinkScanner{}
}

func (m *MockLinkScanner) ScanURL(ctx context.Context, rawURL string) (ScanResult, error) {
	if strings.TrimSpace(rawURL) == "" {
		return ScanResult{Status: domain.ScanSafe, Verdict: "sin link externo"}, nil
	}
	lower := strings.ToLower(rawURL)
	if strings.Contains(lower, "malware") || strings.Contains(lower, "phish") || strings.Contains(lower, "scam") {
		return ScanResult{Status: domain.ScanMalicious, Verdict: "mock: el link contiene indicadores de phishing/malware"}, nil
	}
	if strings.Contains(lower, "short") || strings.Contains(lower, "bit.ly") {
		return ScanResult{Status: domain.ScanSuspicious, Verdict: "mock: link acortado o de confianza limitada"}, nil
	}
	return ScanResult{Status: domain.ScanSafe, Verdict: "mock: no se detectaron indicadores riesgosos"}, nil
}

type VirusTotalScanner struct {
	apiKey string
	client *http.Client
}

func (v *VirusTotalScanner) ScanURL(ctx context.Context, rawURL string) (ScanResult, error) {
	if strings.TrimSpace(rawURL) == "" {
		return ScanResult{Status: domain.ScanSafe, Verdict: "sin link externo"}, nil
	}
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return ScanResult{Status: domain.ScanSuspicious, Verdict: "URL inválida o no normalizada"}, nil
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte(rawURL))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.virustotal.com/api/v3/urls/"+encoded, nil)
	if err != nil {
		return ScanResult{}, err
	}
	req.Header.Set("x-apikey", v.apiKey)

	resp, err := v.client.Do(req)
	if err != nil {
		return ScanResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ScanResult{Status: domain.ScanPending, Verdict: "VirusTotal no posee análisis previo para esta URL"}, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ScanResult{Status: domain.ScanPending, Verdict: "análisis externo no disponible"}, nil
	}

	var payload struct {
		Data struct {
			Attributes struct {
				Stats struct {
					Malicious  int `json:"malicious"`
					Suspicious int `json:"suspicious"`
				} `json:"last_analysis_stats"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ScanResult{}, err
	}
	if payload.Data.Attributes.Stats.Malicious > 0 {
		return ScanResult{Status: domain.ScanMalicious, Verdict: "VirusTotal detectó motores maliciosos"}, nil
	}
	if payload.Data.Attributes.Stats.Suspicious > 0 {
		return ScanResult{Status: domain.ScanSuspicious, Verdict: "VirusTotal detectó motores sospechosos"}, nil
	}
	return ScanResult{Status: domain.ScanSafe, Verdict: "VirusTotal sin detecciones"}, nil
}
