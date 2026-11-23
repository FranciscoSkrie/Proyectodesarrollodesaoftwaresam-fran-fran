# Flujo GitFlow propuesto

- `main`: versiones estables que se pueden presentar.
- `develop`: integración del trabajo terminado.
- `feature/<nombre>`: una rama por funcionalidad o corrección.
- `release/<version>`: estabilización y pruebas antes de entregar.
- `hotfix/<tema>`: correcciones urgentes sobre `main`.

## Flujo de trabajo

1. Crear `develop` desde `main` si todavía no existe.
2. Crear cada rama `feature/*` desde `develop`.
3. Hacer commits pequeños y descriptivos.
4. Abrir Pull Request hacia `develop` y revisar los cambios.
5. Crear `release/v7` desde `develop`, ejecutar pruebas y corregir.
6. Integrar la release en `main` y `develop`, y etiquetar `v7.0.0`.

Este documento describe el flujo esperado; la evidencia real de uso se verifica en el historial y las ramas del repositorio remoto.
