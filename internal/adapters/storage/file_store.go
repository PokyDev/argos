// Package storage implementa ports.HistoryStore guardando cada sesión
// como un archivo JSON individual en disco. Es el único paquete del
// proyecto que sabe que la persistencia de conversaciones es,
// concretamente, "un JSON por sesión en .argos/sessions/".
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pokydev/argos/internal/core/domain"
)

// FileStore es la implementación de ports.HistoryStore para Fase 3.
type FileStore struct {
	sessionsDir string
}

// New prepara el store dentro de <root>/.argos/sessions/, creando el
// directorio si hace falta. root es la raíz del proyecto sobre la que
// corre argos — hoy siempre el directorio de trabajo actual; resolver
// --path (RF-12) queda para el checklist de /init. Nunca se usa $HOME,
// para respetar RNF-08 (acceso limitado a rutas explícitas).
func New(root string) (*FileStore, error) {
	dir := filepath.Join(root, ".argos", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("no se pudo preparar el directorio de historial %s: %w", dir, err)
	}
	return &FileStore{sessionsDir: dir}, nil
}

func (s *FileStore) path(id string) string {
	return filepath.Join(s.sessionsDir, id+".json")
}

// Save implementa ports.HistoryStore.
func (s *FileStore) Save(session domain.Session) error {
	if session.ID == "" {
		return fmt.Errorf("no se puede guardar una sesión sin ID")
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("no se pudo serializar la sesión: %w", err)
	}

	if err := os.WriteFile(s.path(session.ID), data, 0o644); err != nil {
		return fmt.Errorf("no se pudo escribir el archivo de la sesión: %w", err)
	}
	return nil
}

// List implementa ports.HistoryStore. Sesiones más recientes primero.
func (s *FileStore) List() ([]domain.SessionMeta, error) {
	entries, err := os.ReadDir(s.sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el directorio de historial: %w", err)
	}

	metas := make([]domain.SessionMeta, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.sessionsDir, e.Name()))
		if err != nil {
			// Un archivo individual ilegible no debe tumbar todo el
			// listado (RNF-06) — se ignora esa entrada y se sigue.
			continue
		}

		var sess domain.Session
		if err := json.Unmarshal(data, &sess); err != nil {
			continue
		}

		metas = append(metas, domain.SessionMeta{
			ID:        sess.ID,
			Title:     sess.Title,
			Summary:   sess.Summary,
			CreatedAt: sess.CreatedAt,
		})
	}

	sort.Slice(metas, func(i, j int) bool {
		return metas[i].CreatedAt.After(metas[j].CreatedAt)
	})

	return metas, nil
}

// Load implementa ports.HistoryStore.
func (s *FileStore) Load(id string) (domain.Session, error) {
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		return domain.Session{}, fmt.Errorf("no se encontró la sesión '%s': %w", id, err)
	}

	var sess domain.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return domain.Session{}, fmt.Errorf("el archivo de la sesión '%s' está corrupto: %w", id, err)
	}

	return sess, nil
}
