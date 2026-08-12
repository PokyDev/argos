package ports

import "github.com/pokydev/argos/internal/core/domain"

// ModelProvider define el contrato que el core necesita para detectar
// modelos de IA instalados localmente, sin conocer el backend concreto
// (Ollama hoy, otro backend mañana — ver RNF-07).
type ModelProvider interface {
	// ListModels devuelve los modelos detectados localmente, indicando
	// cuáles están actualmente cargados en memoria. Si el backend no
	// está disponible (ej. Ollama no está corriendo), devuelve un error
	// describible al usuario.
	ListModels() ([]domain.Model, error)
}

// ModelRunner es el puerto para enviar un prompt a un modelo local y
// recibir su respuesta completa. Es la pieza que conecta el chat
// interactivo (RF-04) con el modelo activo de la sesión.
type ModelRunner interface {
	// Generate envía prompt al modelo indicado y devuelve su respuesta
	// completa (sin streaming, por ahora — ver notas de Fase 2).
	Generate(model, prompt string) (string, error)
}
