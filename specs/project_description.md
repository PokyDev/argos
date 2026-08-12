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
| RF-08 | El sistema debe mantener un **historial de conversación** dentro de la sesión activa. |
| RF-09 | El sistema debe permitir **adjuntar/referenciar archivos o carpetas** del proyecto para dar contexto al agente. |
| RF-10 | El sistema debe contar con comandos internos tipo "slash command", entre ellos (propuesta inicial): |
| | `/help` — lista de comandos disponibles |
| | `/model` — listar/cambiar modelo de IA activo |
| | `/voice on\|off` — activar/desactivar modo voz |
| | `/scan <ruta>` — auditar una carpeta o archivo de código |
| | `/context` — mostrar archivos/contexto cargado actualmente |
| | `/clear` — limpiar contexto o historial de sesión |
| | `/history` — ver historial de la conversación actual |
| | `/exit` o `/quit` — salir de la sesión interactiva |
| | `/config` — ver o editar configuración local de Argos |
| | `/models list` — refrescar/listar modelos detectados localmente |
| RF-11 | El sistema debe guardar configuración local del usuario (modelo por defecto, preferencias de voz, rutas frecuentes) en un archivo de configuración persistente. |
| RF-12 | El sistema debe permitir ejecutar Argos apuntando a un directorio de proyecto específico (`argos --path <ruta>`). |
| RF-13 | El sistema debe generar reportes/resúmenes de auditoría de código exportables (ej. Markdown o texto plano). |

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

### Fase 3 — Chat interactivo y contexto de proyecto (Días 9–12)
- [ ] Implementar historial de conversación (`/history`).
- [ ] Implementar carga de contexto de archivos/carpetas (`/context`, `--path`).
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
