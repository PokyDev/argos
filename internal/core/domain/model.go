package domain

// Model representa un modelo de IA detectado en un backend local
// (Ollama hoy, otros backends en el futuro — ver RNF-07).
type Model struct {
	Name   string // nombre del modelo tal como lo reporta el backend (ej. "llama3:8b")
	Size   int64  // tamaño en bytes
	Loaded bool   // true si el modelo está actualmente cargado en memoria
}
