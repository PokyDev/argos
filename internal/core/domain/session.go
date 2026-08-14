package domain

import (
	"crypto/rand"
	"fmt"
	"time"
)

// Message representa un turno individual dentro de una conversación
// (RF-08). Role es "user" o "assistant".
type Message struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	At      time.Time `json:"at"`
}

// Session agrupa el estado completo de una conversación: identidad,
// modelo activo y el historial de mensajes.
//
// A partir de Fase 3, ActiveModel vive acá (dejó de ser campo suelto de
// CommandDispatcher, ver phase_two.md §8 y project_description.md §6).
// La dueña de la instancia activa es SessionService; CommandDispatcher
// solo recibe un puntero a ella y no mantiene estado de sesión propio.
type Session struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	CreatedAt   time.Time `json:"created_at"`
	ActiveModel string    `json:"active_model"`
	Messages    []Message `json:"messages"`
}

// NewSession crea una sesión vacía con un ID único basado en timestamp,
// lista para usarse como sesión activa desde el arranque de argos.
func NewSession() *Session {
	return &Session{
		ID:        newSessionID(),
		CreatedAt: time.Now(),
	}
}

// newSessionID genera un identificador legible y con baja probabilidad
// de colisión sin depender de una librería de UUID externa (RNF-04:
// se prefiere mantener el árbol de dependencias mínimo y 100% local).
func newSessionID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%x", time.Now().Format("20060102-150405"), b)
}

// SessionMeta es la vista liviana de una Session usada por /history para
// listar conversaciones guardadas sin tener que cargar todos los
// mensajes de cada una (RF-08).
type SessionMeta struct {
	ID        string
	Title     string
	Summary   string
	CreatedAt time.Time
}

// ListOption es una entrada genérica para selección interactiva sobre
// ports.SessionIO.SelectFromList. La usa /history hoy (elegir
// conversación); /init la reutilizará en un checklist posterior de
// Fase 3 para elegir modelo activo (mismo componente de selección, ver
// project_description.md §6).
type ListOption struct {
	Title       string
	Description string
}
