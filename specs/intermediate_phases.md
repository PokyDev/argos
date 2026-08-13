# Fases Intermedias — Registro de cambios entre fases

> Documento de trabajo bajo metodología SDD. Complementa a `project_description.md`
> (fuente de verdad) y a los `phase_N.md` (cierres de fase del cronograma de 20 días).
> Este archivo agrupa **todas** las fases intermedias (2.5, 3.5, etc. si llegan a
> existir) — cada una como una sección propia, no un archivo por fase intermedia.

---

## Fase Intermedia 2.5 — Mejoras de visualización en terminal

**Estado: completada.** Motivada por RF-14 y RNF-12 (ver nota 12/08/2026 en
`project_description.md` §6). Objetivo: input fijo separado del historial,
con scroll independiente, sin tocar `core`.

### Checklist

- [x] Área de input fija en la parte inferior, separada del historial por un separador visual.
- [x] Input fijo mientras el historial hace scroll independiente.
- [x] Eco del input en el historial tras Enter (persiste aunque el campo se limpie).
- [x] Saltos de línea entre cada bloque input/output.

### Estructura de carpetas (cambios)

```
argos/
├── internal/
│   └── adapters/
│       └── terminal/
│           ├── terminal.go   [reescrito]
│           └── model.go      [nuevo]
└── go.mod                    [+ bubbletea, bubbles, lipgloss]
```

Todo el cambio vive en `internal/adapters/terminal/`. `core` no se tocó —
`ports.SessionIO` y `cmd/argos/main.go` quedan idénticos a Fase 2. Esto valida
en la práctica la promesa de la arquitectura hexagonal (ver `phase_one.md` §1):
reemplazar por completo el mecanismo de I/O no requirió cambiar el core.

### El problema de diseño y cómo se resolvió

`ports.SessionIO.ReadLine()` es una llamada **bloqueante**: `SessionService`
se queda esperando ahí. Un `tea.Program` de bubbletea, en cambio, es dueño de
su propio loop de eventos (`Update`/`View`) y no expone nada parecido a un
`ReadLine()` síncrono.

**Puente implementado en `terminal.go`:** el `tea.Program` corre en una
goroutine separada. Dos canales lo conectan con el resto del sistema:
- **Entrada:** el modelo bubbletea empuja cada línea confirmada (Enter) a un
  `chan string` con buffer (16). `ReadLine()` bloquea leyendo de ahí.
- **Salida:** `WriteLine()` usa `tea.Program.Send()` — el mecanismo soportado
  por bubbletea para inyectar mensajes desde otra goroutine sin condiciones
  de carrera.
- **Cierre:** Ctrl+C dentro de la TUI, o `Close()` desde `main.go`, terminan
  el `tea.Program`. Eso cierra un `doneCh` que `ReadLine()` también escucha,
  devolviendo `io.EOF` — mismo patrón EOF de Fase 1, sin inventar un error nuevo.
  (Nota: Ctrl+C reemplaza al viejo "EOF de stdin" como forma de cerrar sesión;
  ya no aplica el escenario de pipe/redirect de Fase 1.)

### Archivos

**`internal/adapters/terminal/model.go`** — implementa `tea.Model`.
Composición: `viewport.Model` (historial) + `textinput.Model` (input fijo),
unidos con `lipgloss` en `View()`. Responsabilidades no obvias:
- El envío por `inputCh` en `KeyEnter` es no bloqueante (`select`/`default`)
  a propósito: si el core sigue procesando el turno anterior (ej. inferencia
  en curso), un Enter no debe congelar el redibujado de la TUI.
- Las teclas se enrutan explícitamente: `PgUp`/`PgDown`/rueda del mouse van
  al `viewport`; todo lo demás va solo al `textinput`. No se reenvían todas
  las teclas al `viewport` porque `bubbles/viewport` bindea `j`/`k` para
  scroll por defecto — escribir esas letras en el prompt dispararía scroll
  accidental si se dejara pasar sin filtrar.

**`internal/adapters/terminal/terminal.go`** — implementa `ports.SessionIO`.
Es la traducción literal del puente descripto arriba a los tres métodos del
contrato (`ReadLine`/`WriteLine`/`Close`).

### Bugs encontrados y solución

**Bug 1 — outputs no responsivos al resize.** `bubbles/viewport` no hace
word-wrap: `SetContent()` guarda el string tal cual y lo recorta (clipping)
al `Width` actual, sin reflowear al cambiar el ancho. En la primera versión,
`WindowSizeMsg` actualizaba `viewport.Width`/`Height` pero nunca volvía a
llamar `SetContent` con el contenido re-wrappeado. **Fix:** función
`refreshViewportContent()` que re-renderiza *todo* el historial con
`lipgloss.NewStyle().Width(...)` (que sí wrappea) cada vez que cambia el
ancho o se agrega contenido — no solo al agregar.

**Bug 2 — duplicado visual del input tras resize rápido.** Causa no
confirmada con certeza (no reproducible en el entorno de desarrollo de
Claude, sin TTY real). Hipótesis fundamentada: artefacto de redibujado
donde un frame viejo no se limpia del todo antes de dibujar el nuevo;
plausible en Windows porque no hay `SIGWINCH` — bubbletea hace *polling*
del tamaño de consola, lo que puede capturar tamaños intermedios durante un
resize rápido. **Mitigación aplicada:** devolver `tea.ClearScreen` en cada
`WindowSizeMsg`, forzando un clear+redraw completo en vez de confiar en el
diffing incremental del renderer. **Validado por Javier como aceptable**
para esta fase; el costo (un clear completo por resize, no por cada
render) se acepta como trade-off consciente, no como deuda técnica
oculta — queda anotado para revisar si se encuentra una causa raíz más
precisa o una solución más granular a futuro.

### Notas para Fase 3

- El diseño de dos regiones no introduce ningún acoplamiento nuevo con
  `domain.Session`/`domain.Message` (pendientes de Fase 3) — el historial en
  `model.go` es puramente visual (`[]string`), no reemplaza al historial real
  de conversación que se va a modelar en el core.
- Housekeeping resuelto en esta fase: `argos.exe` dejó de trackearse en git
  (estaba pendiente desde `phase_two.md` §8).
