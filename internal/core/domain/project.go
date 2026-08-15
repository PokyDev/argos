package domain

// ScannedFile representa un archivo de texto del proyecto detectado por
// el escáner (ports.ProjectScanner), listo para pasarse como contexto a
// un modelo. Path es relativo a la raíz del proyecto (más legible en
// prompts y en el propio ARGOS.md que /init genera).
type ScannedFile struct {
	Path    string
	Content string
}
