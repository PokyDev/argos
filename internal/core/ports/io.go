package ports

// SessionIO define el contrato de entrada/salida que el core necesita
// para interactuar con el usuario, sin conocer el mecanismo concreto
// (terminal hoy, voz o socket remoto en el futuro).
type SessionIO interface {
	// ReadLine bloquea hasta recibir una línea de entrada del usuario.
	// Devuelve io.EOF cuando no hay más entrada disponible (ej. Ctrl+D).
	ReadLine() (string, error)

	// WriteLine imprime una línea de salida hacia el usuario.
	WriteLine(msg string)

	// Close libera recursos del adaptador (si aplica).
	Close() error
}
