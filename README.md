# Argos CLI

CLI interactiva desarrollada en Go (`go1.26.4`, target `windows/amd64`), orientada a **auditoría de código, revisión y debugging** — no a generación masiva de código. Detecta modelos de IA locales (vía Ollama) y prioriza un enfoque *local-first*.

## Estructura del proyecto

```
argos/
├── cmd/
│   └── argos/
│       └── main.go
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   ├── ports/
│   │   └── usecase/
│   ├── adapters/
│   │   ├── ollama/
│   │   ├── voice/
│   │   ├── config/
│   │   ├── scanner/
│   │   └── terminal/
│   └── platform/
├── configs/
├── docs/
└── go.mod
```

Arquitectura: **Hexagonal (Ports & Adapters)**. El `core` no depende de `adapters`; cada integración externa (modelo, voz, config, terminal) es un adaptador intercambiable detrás de una interfaz definida en `core/ports`.

## Checklist global (20 días)

### Fase 1 — Fundamentos y estructura base (Días 1–4)
- [x] Definir arquitectura general del proyecto
- [x] Configurar entorno de desarrollo (Go 1.26.4, módulos, dependencias base)
- [x] Implementar el comando `argos` y el loop interactivo (REPL) básico
- [x] Implementar sistema base de comandos slash (`/help`, `/exit`, `/clear`)

### Fase 2 — Integración con modelos locales (Días 5–8)
- [x] Detección automática de modelos vía Ollama
- [x] Comandos `/model` y `/models list`
- [x] Comunicación básica CLI ↔ modelo local
- [x] Manejo de errores cuando Ollama no está corriendo

### Fase Intermedia 2.5 — Mejoras de visualización en terminal

> Motivación: escribir directamente debajo de los outputs resulta incómodo y el contenido se ve amontonado. Se busca una experiencia de terminal más clara, con un campo de input fijo y separación visual entre turnos.

- [ ] Definir un área de input fija en la parte inferior de la terminal, separada del área de output mediante un separador visual (línea horizontal u otro delimitador).
- [ ] Garantizar que el campo de input permanezca fijo (no se desplace) mientras el historial de outputs hace scroll hacia arriba.
- [ ] Tras enviar un input, registrarlo (echo) en el historial/scroll de la terminal para que quede visible junto a su output correspondiente.
- [ ] Insertar saltos de línea entre cada bloque de input y su output correspondiente, para evitar que el contenido se vea pegado.

### Fase 3 — Chat interactivo y contexto de proyecto (Días 9–12)
- [ ] Historial de conversación (`/history`)
- [ ] Carga de contexto de archivos/carpetas (`/context`, `--path`)
- [ ] Comando `/scan` para auditoría de código
- [ ] Generación de reportes de auditoría exportables

### Fase 4 — Interacción por voz (Días 13–16)
- [ ] Selección de librerías STT
- [ ] Comando `/voice on|off`
- [ ] Entrada por voz funcional
- [ ] (Opcional) Salida por voz (TTS)

### Fase 5 — Configuración, pulido y pruebas (Días 17–20)
- [ ] Sistema de configuración persistente (`/config`)
- [ ] Pruebas unitarias para módulos core
- [ ] Pulido de UX de terminal
- [ ] Documentación (README interno) y validación de build final `windows/amd64`

---
*Documento fuente de verdad del proyecto: `docs/project_description.md`. Cambios de alcance o arquitectura deben reflejarse ahí primero.*