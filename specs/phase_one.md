# Fase 1 — Fundamentos y estructura base (Días 1–4)

> Documento de trabajo bajo metodología SDD. Complementa a `project_description.md` (fuente de verdad del proyecto). Se actualiza a medida que avanza la Fase 1.

---

## 1. Decisión de arquitectura

**Arquitectura Hexagonal (Ports & Adapters)**, sin las capas adicionales de Clean, con vocabulario ligero de dominio tomado de DDD donde aporte legibilidad (sin agregados ni value objects formales).

**Justificación:** el problema central de Argos no es modelar un dominio de negocio complejo, sino aislar el core conversacional (parseo de comandos, orquestación de sesión) de un conjunto de integraciones externas que van a cambiar con el tiempo: Ollama hoy, otro backend de modelos mañana; terminal hoy, posible socket/API para el módulo móvil después; sin voz hoy, con STT/TTS en la Fase 4. Hexagonal resuelve esto con el mínimo de ceremonia y es idiomático en Go (interfaces pequeñas + inyección de dependencias explícita en `main.go`, sin frameworks de DI).

### Estructura de carpetas

```
argos/
├── cmd/
│   └── argos/
│       └── main.go                # entrypoint, wiring de adaptadores, flags (--path)
├── internal/
│   ├── core/                      # núcleo, no depende de nada externo
│   │   ├── domain/                # entidades ligeras: Session, Message, ModelInfo, Command
│   │   ├── ports/                 # interfaces: ModelProvider, VoiceInput, VoiceOutput, ConfigStore, CodeScanner
│   │   └── usecase/               # orquestación: SessionService, ScanService, CommandDispatcher
│   ├── adapters/                  # implementaciones concretas de los ports
│   │   ├── ollama/                # implementa ModelProvider
│   │   ├── voice/                 # implementa VoiceInput/VoiceOutput (STT/TTS) — Fase 4
│   │   ├── config/                # implementa ConfigStore (archivo local persistente, RF-11)
│   │   ├── scanner/                # implementa CodeScanner (lectura/auditoría de archivos, RF-06/RF-13)
│   │   └── terminal/              # REPL, render de colores/formato, lectura de stdin (RF-04, RNF-05)
│   └── platform/                  # cross-cutting: logging, manejo de errores, utilidades internas
├── configs/                        # config por defecto / ejemplo
├── docs/                           # project_description.md + documentos de fase (SDD)
└── go.mod
```

**Regla de dependencia:** `core` no importa nada de `adapters`. Los `adapters` implementan las interfaces definidas en `core/ports`. `cmd/argos/main.go` es el único punto que conoce e inyecta las implementaciones concretas.

---

## 2. Checklist Fase 1 (Días 1–4)

- [x] Definir arquitectura general del proyecto (paquetes, estructura de carpetas Go).
- [x] Configurar entorno de desarrollo (Go 1.26.4, módulos, dependencias base).
- [ ] Implementar el comando `argos` y el loop interactivo (REPL) básico.
- [ ] Implementar sistema base de comandos slash (`/help`, `/exit`, `/clear`).

---

## 3. Notas / próximos pasos

- **Entorno de desarrollo completado:** módulo Go inicializado, estructura de carpetas creada, `cmd/argos/main.go` mínimo (imprime "Argos CLI iniciado") compilado y validado para `windows/amd64`. Repositorio `argos` creado en GitHub (público, rama por defecto `main`), con `README.md` (estructura + checklist global) y `.gitignore` ya commiteados y pusheados.
- **Siguiente paso (tercer ítem del checklist):** implementar el comando `argos` y el loop interactivo (REPL) básico. Esto implica:
  - Reemplazar el `main.go` mínimo por uno que arranque un loop de lectura de stdin (`bufio.Scanner` o similar) y lo mantenga corriendo hasta una señal de salida.
  - Definir dónde vive esta lógica: al ser interacción por terminal, corresponde a `internal/adapters/terminal/` (no al `core` directamente — el core solo debe conocer una interfaz de entrada/salida, no `os.Stdin` concreto).
  - Dejar el loop preparado para recibir texto libre por ahora (el parseo de comandos slash es el cuarto ítem, se aborda después).
- Los `ports` de `core/ports` deben definirse antes de escribir cualquier adaptador — el orden correcto es: interfaz primero, implementación después.
- Este documento se actualiza en cada sesión conforme se completen ítems del checklist; no reemplaza a `project_description.md`, lo complementa con el detalle de ejecución de esta fase.