package usecase

import (
	"fmt"
	"strings"

	"github.com/pokydev/argos/internal/core/domain"
	"github.com/pokydev/argos/internal/core/ports"
)

// summarizeSession completa Session.Title y Session.Summary antes de
// persistir (RF-08: "título y resumen se generan con el modelo activo
// al persistir la sesión", project_description.md §6).
//
// Si no hay modelo activo, o la llamada al modelo falla, o la respuesta
// no viene en el formato esperado, cae a un fallback derivado del
// primer mensaje del usuario. Esto es necesario en la práctica: los
// modelos chicos (0.5B–3B) con los que corre Argos en el hardware
// CPU-bound de destino no siempre respetan instrucciones de formato de
// forma confiable (ver phase_two.md §7) — un resumen "mejor esfuerzo"
// es preferible a que /history falle por completo.
func summarizeSession(runner ports.ModelRunner, session *domain.Session) {
	fallbackTitle, fallbackSummary := fallbackTitleAndSummary(session)

	if session.ActiveModel == "" {
		session.Title, session.Summary = fallbackTitle, fallbackSummary
		return
	}

	raw, err := runner.Generate(session.ActiveModel, summarizationPrompt(session))
	if err != nil {
		session.Title, session.Summary = fallbackTitle, fallbackSummary
		return
	}

	title, summary, ok := parseTitleAndSummary(raw)
	if !ok {
		title, summary = fallbackTitle, fallbackSummary
	}
	session.Title, session.Summary = title, summary
}

func summarizationPrompt(session *domain.Session) string {
	var b strings.Builder
	b.WriteString("Resumí la siguiente conversación en exactamente dos líneas, sin texto adicional ni explicaciones:\n")
	b.WriteString("Titulo: <título corto, menos de 8 palabras>\n")
	b.WriteString("Resumen: <una oración breve>\n\n")
	b.WriteString("Conversación:\n")
	for _, m := range session.Messages {
		b.WriteString(fmt.Sprintf("%s: %s\n", m.Role, truncate(m.Content, 300)))
	}
	return b.String()
}

// parseTitleAndSummary interpreta la respuesta del modelo buscando
// líneas "Titulo:"/"Título:" y "Resumen:" en cualquier orden, sin
// asumir mayúsculas exactas ni que no haya texto extra alrededor
// (tolerante a que modelos chicos no sigan el formato al pie de la letra).
func parseTitleAndSummary(raw string) (title, summary string, ok bool) {
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "titulo:"):
			title = strings.TrimSpace(line[len("titulo:"):])
		case strings.HasPrefix(lower, "título:"):
			title = strings.TrimSpace(line[len("título:"):])
		case strings.HasPrefix(lower, "resumen:"):
			summary = strings.TrimSpace(line[len("resumen:"):])
		}
	}
	return title, summary, title != "" && summary != ""
}

// fallbackTitleAndSummary usa el primer mensaje del usuario cuando no
// se pudo generar título/resumen con el modelo.
func fallbackTitleAndSummary(session *domain.Session) (title, summary string) {
	for _, m := range session.Messages {
		if m.Role == "user" {
			return truncate(m.Content, 40), truncate(m.Content, 100)
		}
	}
	return "Conversación sin título", ""
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
