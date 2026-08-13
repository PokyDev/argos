# Project Description — Argos CLI

> Documento maestro bajo metodología **SDD (Specification-Driven Development)**.
> Este archivo debe ser adjuntado al inicio de cada sesión de trabajo para mantener el contexto completo del proyecto.

---

## 1. Nombre

**Argos CLI**

---

## 2. Descripción

Argos es una **CLI interactiva desarrollada en Go** (versión `go1.26.4`, target `windows/amd64`), inspirada en el funcionamiento de herramientas como **Claude Code** o **Antigravity CLI**, pero con personalización y funcionalidades propias.

El objetivo central es contar con un agente conversacional local que:

- Detecte automáticamente los **modelos de IA instalados localmente** (por ejemplo, modelos servidos por **Ollama**).
- Permita **interacción por chat y por voz**.
- Esté orientado principalmente a la **auditoría de código, revisión de código existente y debugging** — **no** está pensado para generar código complejo desde cero.
- Sea de **uso personal**.

**Visión a futuro (fuera del alcance de la primera versión):** integración con un dispositivo móvil que permita manipular el equipo de escritorio de forma remota a través del agente Argos.

### Principios de diseño
- Local-first: prioriza modelos locales (Ollama u otros) sobre servicios en la nube.
- Enfoque en lectura/análisis de código, no en generación masiva.
- Extensible: arquitectura modular para agregar comandos y capacidades (voz, móvil) sin romper el core.

---

## 3. Requisitos Funcionales

| ID | Requisito |
|----|-----------|
| RF-01 | El sistema debe iniciarse mediante el comando `argos` en terminal, lanzando una sesión interactiva (REPL). |
| RF-02 | El sistema debe detectar automáticamente los modelos de IA instalados localmente (ej. modelos disponibles vía Ollama), listando nombre, tamaño y estado (disponible/cargado). |
| RF-03 | El usuario debe poder seleccionar/cambiar el modelo activo dentro de la sesión (`/model <nombre>`). |
| RF-04 | El sistema debe soportar interacción por **texto/chat** dentro de la terminal. |
| RF-05 | El sistema debe soportar interacción por **voz** (entrada por micrófono y opcionalmente salida por voz/TTS). |
| RF-06 | El agente debe poder **leer y analizar archivos de código** del proyecto actual (auditoría de código). |
| RF-07 | El agente debe poder **detectar y explicar errores** (debugging) en código existente, sin generar soluciones de alta complejidad. |
| RF-08 | El sistema debe mantener y **persistir** el historial de conversación por proyecto (directorio `.argos/`), permitiendo listar conversaciones previas mediante `/history` (título, fecha, resumen) y retomar (cargar) una conversación seleccionada como sesión activa. |
| RF-09 | El sistema debe permitir construir contexto de proyecto para el agente mediante `/init`: tras seleccionar el modelo activo desde una lista interactiva de modelos detectados (RF-02), el sistema recorre los archivos del proyecto y genera un archivo persistente `ARGOS.md` en la raíz del proyecto con un resumen de arquitectura/propósito, que se carga automáticamente como contexto activo en cada sesión mientras exista. |
| RF-10 | El sistema debe contar con comandos internos tipo "slash command", entre ellos (propuesta inicial): |
| | `/help` — lista de comandos disponibles |
| | `/model` — listar/cambiar modelo de IA activo |
| | `/voice on\|off` — activar/desactivar modo voz |
| | `/scan <ruta>` — auditar una carpeta o archivo de código |
| | `/init` — generar/regenerar `ARGOS.md` (contexto persistente del proyecto) |
| | `/clear` — limpiar contexto o historial de sesión |
| | `/history` — ver/retomar conversaciones previas |
| | `/exit` o `/quit` — salir de la sesión interactiva |
| | `/config` — ver o editar configuración local de Argos |
| | `/models list` — refrescar/listar modelos detectados localmente |
| RF-11 | El sistema debe guardar configuración local del usuario (modelo por defecto, preferencias de voz, rutas frecuentes) en un archivo de configuración persistente. |
| RF-12 | El sistema debe permitir ejecutar Argos apuntando a un directorio de proyecto específico (`argos --path <ruta>`). |
| RF-13 | El sistema debe generar reportes/resúmenes de auditoría de código exportables (ej. Markdown o texto plano). |
| RF-14 | El sistema debe presentar la interfaz de terminal con un **área de input fija** en la parte inferior de la pantalla, separada visualmente del área de output (historial de conversación) mediante un separador, de modo que el historial pueda desplazarse (scroll) de forma independiente al campo de escritura. |

---

## 4. Requisitos No Funcionales

| ID | Requisito |
|----|-----------|
| RNF-01 | **Rendimiento:** la detección de modelos locales y el arranque de la CLI no debe tardar más de unos pocos segundos en condiciones normales. |
| RNF-02 | **Portabilidad del target:** compilación y ejecución garantizada para `windows/amd64` con Go `1.26.4`; código escrito de forma que facilite portar a Linux/macOS a futuro. |
| RNF-03 | **Modularidad:** arquitectura en paquetes Go desacoplados (core CLI, agente/IA, voz, detección de modelos, configuración) para facilitar mantenimiento y extensión futura (ej. módulo móvil). |
| RNF-04 | **Privacidad y control local:** priorizar procesamiento local (modelos vía Ollama) sin envío obligatorio de datos a servicios externos, dado que es una herramienta de uso personal. |
| RNF-05 | **Usabilidad de terminal:** interfaz interactiva clara, con colores/formato legible y mensajes de error entendibles. |
| RNF-06 | **Confiabilidad:** manejo adecuado de errores (modelo no disponible, Ollama no corriendo, micrófono no detectado, etc.) sin cerrar abruptamente la sesión. |
| RNF-07 | **Extensibilidad:** el diseño debe permitir añadir nuevos comandos y nuevos backends de modelos (no solo Ollama) sin reescribir el núcleo. |
| RNF-08 | **Seguridad básica:** el acceso a archivos del sistema debe limitarse a las rutas explícitamente indicadas por el usuario. |
| RNF-09 | **Mantenibilidad:** código Go idiomático, con pruebas unitarias mínimas para los módulos core (detección de modelos, parsing de comandos). |
| RNF-10 | **Configurabilidad:** parámetros clave (modelo por defecto, idioma de voz, formato de reportes) deben ser configurables sin recompilar el binario. |
| RNF-11 | **Preparación para futuro módulo móvil:** el diseño del agente y su capa de comunicación deben evitar acoplamientos que impidan exponer una API/socket local reutilizable desde una app móvil más adelante. |
| RNF-12 | **Legibilidad de sesión:** debe existir separación visual (saltos de línea) entre cada bloque de input y su output correspondiente para evitar que el contenido se vea aglomerado; todo input enviado por el usuario debe quedar registrado de forma persistente en el historial visible de la terminal, incluso después de que el campo de input fijo lo "limpie" para el siguiente mensaje. |

---

## 5. Cronograma por Fases (20 días)

### Fase 1 — Fundamentos y estructura base (Días 1–4)
- [x] Definir arquitectura general del proyecto (paquetes, estructura de carpetas Go).
- [x] Configurar entorno de desarrollo (Go 1.26.4, módulos, dependencias base).
- [x] Implementar el comando `argos` y el loop interactivo (REPL) básico.
- [x] Implementar sistema base de comandos slash (`/help`, `/exit`, `/clear`).

### Fase 2 — Integración con modelos locales (Días 5–8)
- [x] Implementar detección automática de modelos instalados vía Ollama.
- [x] Implementar comando `/model` y `/models list`.
- [x] Establecer comunicación básica CLI ↔ modelo local (envío/recepción de prompts).
- [x] Manejo de errores cuando Ollama no está corriendo o no hay modelos disponibles.

### Fase Intermedia 2.5 — Mejoras de visualización en terminal

> Motivación: escribir directamente debajo de los outputs resulta incómodo y el contenido se ve amontonado. Se busca una experiencia de terminal más clara, con un campo de input fijo y separación visual entre turnos.

- [x] Definir un área de input fija en la parte inferior de la terminal, separada del área de output mediante un separador visual (línea horizontal u otro delimitador).
- [x] Garantizar que el campo de input permanezca fijo (no se desplace) mientras el historial de outputs hace scroll hacia arriba.
- [x] Tras enviar un input, registrarlo (echo) en el historial/scroll de la terminal para que quede visible junto a su output correspondiente.
- [x] Insertar saltos de línea entre cada bloque de input y su output correspondiente, para evitar que el contenido se vea pegado.

### Fase 3 — Chat interactivo y contexto de proyecto (Días 9–12)
- [ ] Implementar historial de conversación persistente (`/history`).
- [ ] Implementar `/init` (selección interactiva de modelo + generación de `ARGOS.md`, `--path`).
- [ ] Implementar comando `/scan` para auditoría de código.
- [ ] Implementar generación de reportes de auditoría exportables.

### Fase 4 — Interacción por voz (Días 13–16)
- [ ] Investigar/seleccionar librerías Go (o bindings) para entrada de voz (STT).
- [ ] Implementar comando `/voice on|off`.
- [ ] Implementar entrada por voz funcional dentro de la sesión.
- [ ] (Opcional según tiempo) Implementar salida por voz (TTS).

### Fase 5 — Configuración, pulido y pruebas (Días 17–20)
- [ ] Implementar sistema de configuración persistente (`/config`).
- [ ] Escribir pruebas unitarias para módulos core.
- [ ] Pulir UX de terminal (colores, formato, manejo de errores).
- [ ] Documentar uso de Argos CLI (README interno) y validar build final `windows/amd64`.

---

## 6. Notas de contexto para futuras sesiones

- Este documento es la fuente de verdad del proyecto bajo metodología SDD.
- Cualquier cambio de alcance, nuevo requisito o ajuste de cronograma debe reflejarse aquí antes de continuar el desarrollo.
- La integración móvil **no** forma parte del cronograma actual de 20 días; queda registrada como visión futura (ver sección 2 y RNF-11).

- (12/08/2026): La Fase 2.5 (mejoras de visualización de terminal) introduce RF-14 y RNF-12. **Decisión tomada:** se implementará con `bubbletea` + `lipgloss` (+ `bubbles/textinput` solo para el campo de input), en lugar de `tview` o ANSI manual. Justificación: da control explícito sobre el layout de dos regiones (historial con scroll + input fijo) sin heredar un framework de widgets completo (tview), y ANSI manual queda descartado por su volatilidad y complejidad de mantenimiento fuera de casos muy puntuales. Vive en `internal/adapters/terminal/`; no afecta a `core` (sigue implementando `ports.SessionIO`). 100% local, cumple RNF-04. [FINISH]

- (13/08/2026): Diseño de Fase 3 (chat interactivo y contexto de proyecto). Los checklists del cronograma quedaban subespecificados, así que se fija el diseño acá antes de escribir código. Nota reescrita en el mismo día tras corregir una confusión: **se mantiene `/history`** (persistencia de conversaciones) y **se descarta `/context`**, reemplazado por `/init` (contexto de proyecto vía `ARGOS.md`, inspirado en `/init` de Claude Code). Actualiza RF-08 y RF-09 (ver sección 3) y resuelve el pendiente arquitectónico anotado desde `phase_two.md` §8.

  1. **Historial persistente (RF-08).** Al arrancar `argos` sobre un directorio (el actual, o el de `--path`, RF-12) se crea `.argos/` en esa misma ruta si no existe — nunca en `$HOME`, para no romper RNF-08 (acceso limitado a rutas explícitas). Cada conversación se guarda como un JSON individual en `.argos/sessions/`. `/history` lista las sesiones guardadas (título, fecha, resumen); título y resumen se generan con el modelo activo al persistir la sesión. La selección (flechas o click) usa un modo de lista sobre el mismo `tea.Program` de Fase 2.5 (candidato: `bubbles/list`), sin introducir un framework de UI nuevo. Seleccionar una conversación la carga como sesión activa (mensajes + modelo usado en su momento).
  2. **Contexto de proyecto vía `/init` (RF-09).** Flujo: (a) lista interactiva de modelos detectados vía Ollama (RF-02) — mismo componente de selección que en el punto 1, aplicado acá a elegir modelo en vez de conversación; (b) el modelo elegido queda como `Session.ActiveModel`; (c) recorrido del árbol de archivos del proyecto excluyendo `.git/`, `.argos/`, `node_modules/`, `vendor/` y binarios (lista fija en esta fase; parsear `.gitignore` queda fuera de alcance para no sobre-diseñar); (d) por cada archivo se le pide al modelo un resumen de propósito/responsabilidad, y esos resúmenes se sintetizan en un único **`ARGOS.md` en la raíz del proyecto** (no dentro de `.argos/`) — a propósito, para que sea versionable en git y editable por humanos, igual que `CLAUDE.md`. Por eso **no** va en `.gitignore` (a diferencia de `.argos/` y `argos.exe`).
  3. **Auto-carga de `ARGOS.md`.** Si existe al arrancar `argos` sobre un proyecto, se carga automáticamente como contexto activo de la sesión — es el punto central de tener un archivo de contexto persistente; no hay comando manual para cargarlo aparte (no existe `/context`). Costo aceptado: agranda el prompt en cada turno; dado el hardware CPU-bound de Javier y los modelos chicos (0.5B–3B) con los que trabaja, esto puede pesar en la práctica. Queda anotado como candidato a volverse configurable (activar/desactivar auto-carga) en Fase 5 vía `/config` (RNF-10) si hace falta.
  4. **`/init` regenera desde cero, no actualiza incrementalmente.** Sin diff/merge con un `ARGOS.md` existente en esta fase — se sobreescribe completo. Limitación de alcance intencional, no deuda técnica.
  5. **`/scan <ruta>` (RF-06, RF-07, RF-13)** no cambia: prompt de auditoría especializado (lenguaje detectado, responsabilidad del archivo, brechas de seguridad, mejoras, refactors posibles), resultado impreso en la sesión y exportado como Markdown a `.argos/reports/`.
  6. **Arquitectura — entidades y puertos.**
     - `domain.Session` / `domain.Message`: entidades pendientes desde Fase 2 (`phase_two.md` §8). `Session` agrupa `ID`, `Title`, `CreatedAt`, `Messages []Message`, `ActiveModel`. El contenido de `ARGOS.md` se carga en memoria al iniciar sesión (no es un campo que el usuario setee a mano, ya no existe `/context` para eso).
     - **Decisión sobre el pendiente de `phase_two.md` §8:** `activeModel` deja de vivir en `CommandDispatcher` y pasa a `Session`, dueña ahora `SessionService`. `CommandDispatcher` deja de tener estado propio de sesión; opera sobre la `Session` activa que recibe.
     - Nuevo puerto `ports.HistoryStore` (`Save(Session) error`, `List() ([]SessionMeta, error)`, `Load(id) (Session, error)`) — implementado en un adapter de archivos JSON (paquete a definir al implementar, ej. `internal/adapters/storage/`).
     - El recorrido de archivos con exclusión para `/init` y `/scan` se centraliza en `internal/adapters/scanner/` (carpeta ya existente y vacía desde Fase 1) — evita duplicar lectura de filesystem y mantiene acotado el acceso (RNF-08) a un solo adapter.
     - La orquestación de `/init` (múltiples llamadas a `Generate`, una por archivo, más síntesis final) amerita su propio archivo en vez de vivir en `command_dispatcher.go`, ej. `internal/core/usecase/project_init.go`.
     - `/clear` deja de ser stub (resuelve el `TODO(fase 3)` de `command_dispatcher.go`): limpia `Session.Messages` de la sesión activa. No toca `ARGOS.md` (eso solo lo regenera `/init`) ni el historial ya persistido en `.argos/sessions/`.
  7. **Housekeeping:** `.argos/` sigue yendo a `.gitignore` (contiene `sessions/` y `reports/`); `ARGOS.md` explícitamente **no** va en `.gitignore`.