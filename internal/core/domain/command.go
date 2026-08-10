package domain

// Command representa una instrucción interna tipo slash-command
// parseada desde la entrada del usuario (ej. "/model llama3").
type Command struct {
	Name string   // nombre sin la barra, en minusculas (ej. "model")
	Args []string // argumentos  separados por espacios (ej. "llama3")
}
