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
		return &VirusTotalScanner{apiKey: apiKey, client: &http.Client{Timeout: 12 * time.Second}}
	}
	return &MockLinkScanner{}
}

func (m *MockLinkScanner) ScanURL(ctx context.Context, rawURL string) (ScanResult, error) {
	if strings.TrimSpace(rawURL) == "" {
		return ScanResult{Status: domain.ScanSafe, Verdict: "sin link externo"}, nil
	}
	lower := strings.ToLower(rawURL)
	if strings.Contains(lower, "malware") || strings.Contains(lower, "phish") || strings.Contains(lower, "scam") || strings.Contains(lower, "steamcommunity") || strings.Contains(lower, "loginform") {
		return ScanResult{Status: domain.ScanMalicious, Verdict: "alerta: el link contiene indicadores compatibles con phishing/malware"}, nil
	}
	if strings.Contains(lower, "short") || strings.Contains(lower, "bit.ly") || strings.Contains(lower, "tinyurl") {
		return ScanResult{Status: domain.ScanSuspicious, Verdict: "alerta: link acortado o de confianza limitada"}, nil
	}
	return ScanResult{Status: domain.ScanSafe, Verdict: "sin detecciones de riesgo"}, nil
}

type VirusTotalScanner struct {
	apiKey string
	client *http.Client
}

func (v *VirusTotalScanner) ScanURL(ctx context.Context, rawURL string) (ScanResult, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ScanResult{Status: domain.ScanSafe, Verdict: "sin link externo"}, nil
	}
	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return ScanResult{Status: domain.ScanSuspicious, Verdict: "alerta: URL inválida o no normalizada"}, nil
	}

	result, found, err := v.getURLReport(ctx, rawURL)
	if err != nil {
		return ScanResult{}, err
	}
	if found {
		return result, nil
	}

	if err := v.submitURL(ctx, rawURL); err != nil {
		return ScanResult{Status: domain.ScanPending, Verdict: "VirusTotal no tenía análisis previo y no se pudo enviar la URL"}, nil
	}
	return ScanResult{Status: domain.ScanPending, Verdict: "URL enviada a VirusTotal; análisis pendiente"}, nil
}

func (v *VirusTotalScanner) getURLReport(ctx context.Context, rawURL string) (ScanResult, bool, error) {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(rawURL))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.virustotal.com/api/v3/urls/"+encoded, nil)
	if err != nil {
		return ScanResult{}, false, err
	}
	req.Header.Set("x-apikey", v.apiKey)

	resp, err := v.client.Do(req)
	if err != nil {
		return ScanResult{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ScanResult{}, false, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ScanResult{Status: domain.ScanPending, Verdict: "análisis externo no disponible"}, true, nil
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
		return ScanResult{}, true, err
	}
	if payload.Data.Attributes.Stats.Malicious > 0 {
		return ScanResult{Status: domain.ScanMalicious, Verdict: "alerta: VirusTotal detectó motores maliciosos"}, true, nil
	}
	if payload.Data.Attributes.Stats.Suspicious > 0 {
		return ScanResult{Status: domain.ScanSuspicious, Verdict: "alerta: VirusTotal detectó motores sospechosos"}, true, nil
	}
	return ScanResult{Status: domain.ScanSafe, Verdict: "VirusTotal sin detecciones"}, true, nil
}

func (v *VirusTotalScanner) submitURL(ctx context.Context, rawURL string) error {
	form := url.Values{}
	form.Set("url", rawURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://www.virustotal.com/api/v3/urls", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("x-apikey", v.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
