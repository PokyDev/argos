// Package scanner implementa ports.ProjectScanner leyendo y escribiendo
// directamente sobre el filesystem local. Es el único adapter del
// proyecto con acceso al árbol de archivos del proyecto activo (RNF-08:
// acceso limitado a las rutas explícitamente indicadas por el usuario —
// acá, la raíz de la sesión).
package scanner

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/pokydev/argos/internal/core/domain"
)

// excludedDirs son las carpetas que /init (y, a futuro, /scan) ignoran
// siempre. Lista fija en esta fase — parsear .gitignore queda fuera de
// alcance para no sobre-diseñar (ver project_description.md §6).
var excludedDirs = map[string]bool{
	".git":         true,
	".argos":       true,
	"node_modules": true,
	"vendor":       true,
}

const (
	// maxFileSize evita que un archivo enorme (ej. un dump de datos o un
	// lockfile gigante) infle desmedidamente el prompt enviado al modelo.
	maxFileSize = 256 * 1024 // 256 KB

	// maxFiles topea la cantidad de archivos que Walk incluye como
	// contexto. Con el hardware CPU-bound de destino (ver phase_two.md
	// §7) y modelos chicos, escanear un proyecto grande archivo por
	// archivo puede tardar minutos — se prioriza que /init termine en un
	// tiempo razonable sobre cobertura exhaustiva. Es un detalle interno
	// del adapter: Walk devuelve un bool `truncated` para avisar al core
	// sin exponerle el número concreto (core no importa adapters).
	maxFiles = 60

	// sniffBytes es cuántos bytes se inspeccionan para decidir si un
	// archivo es binario (heurística de byte nulo), sin depender de una
	// lista fija de extensiones que siempre queda incompleta.
	sniffBytes = 512
)

// Scanner es la implementación de ports.ProjectScanner para Fase 3.
type Scanner struct{}

// New crea un Scanner. No tiene estado propio: cada llamada a Walk/
// WriteContext recibe la raíz explícitamente.
func New() *Scanner {
	return &Scanner{}
}

// Walk implementa ports.ProjectScanner.
func (s *Scanner) Walk(root string) ([]domain.ScannedFile, bool, error) {
	var files []domain.ScannedFile
	var truncated bool

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Un archivo o carpeta puntual ilegible (permisos, etc.) no
			// debe abortar todo el recorrido (RNF-06) — se lo salta.
			return nil
		}

		if d.IsDir() {
			if path != root && excludedDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if len(files) >= maxFiles {
			truncated = true
			return nil
		}

		info, err := d.Info()
		if err != nil || info.Size() == 0 || info.Size() > maxFileSize {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if looksBinary(content) {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}

		files = append(files, domain.ScannedFile{
			Path:    filepath.ToSlash(rel),
			Content: string(content),
		})
		return nil
	})
	if err != nil {
		return nil, false, fmt.Errorf("no se pudo recorrer el proyecto en %s: %w", root, err)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, truncated, nil
}

// looksBinary detecta binarios por contenido (byte nulo en los primeros
// sniffBytes) en vez de por extensión: cubre casos no previstos en una
// lista fija (ej. un .dat, un .bin custom) sin mantenimiento adicional.
func looksBinary(content []byte) bool {
	n := len(content)
	if n > sniffBytes {
		n = sniffBytes
	}
	return bytes.IndexByte(content[:n], 0) != -1
}

// WriteContext implementa ports.ProjectScanner. Escribe ARGOS.md en la
// raíz del proyecto (no en .argos/) a propósito, para que sea
// versionable en git y editable por humanos (ver project_description.md
// §6, punto 2) — por eso tampoco va en .gitignore.
func (s *Scanner) WriteContext(root, content string) error {
	path := filepath.Join(root, "ARGOS.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("no se pudo escribir %s: %w", path, err)
	}
	return nil
}
