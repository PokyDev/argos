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
