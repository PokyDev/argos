package ports

import "github.com/pokydev/argos/internal/core/domain"

// HistoryStore define el contrato que el core necesita para persistir y
// recuperar conversaciones (RF-08), sin conocer el mecanismo de
// almacenamiento concreto. La implementación de Fase 3 guarda cada
// sesión como un JSON en .argos/sessions/ (internal/adapters/storage/),
// pero el core solo ve esta interfaz.
type HistoryStore interface {
	// Save persiste la sesión completa, incluidos sus mensajes.
	// Sobrescribe el archivo existente si ya había una sesión con el
	// mismo ID (misma sesión guardada más de una vez).
	Save(session domain.Session) error

	// List devuelve el listado liviano de sesiones guardadas (sin los
	// mensajes completos), pensado para mostrarse en /history.
	List() ([]domain.SessionMeta, error)

	// Load recupera una sesión completa (con sus mensajes) por ID.
	Load(id string) (domain.Session, error)
}
