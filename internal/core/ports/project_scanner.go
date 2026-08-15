package ports

import "github.com/pokydev/argos/internal/core/domain"

// ProjectScanner define el contrato que el core necesita para recorrer
// el árbol de archivos del proyecto activo y persistir el contexto
// generado, sin conocer el mecanismo concreto de acceso a disco. Lo usa
// /init (RF-09) hoy; /scan (RF-06/RF-07), en un checklist futuro de
// Fase 3, va a reutilizar Walk para no duplicar lectura de filesystem
// (ver project_description.md §6). Centralizar esto en un único adapter
// mantiene acotado RNF-08 (acceso limitado a rutas explícitas).
type ProjectScanner interface {
	// Walk recorre root recursivamente y devuelve los archivos de texto
	// legibles como candidatos a contexto, excluyendo directorios no
	// relevantes (.git, .argos, node_modules, vendor) y archivos
	// binarios. La implementación puede aplicar límites de tamaño/
	// cantidad para no inflar desmedidamente el contexto que después se
	// le pasa a un modelo local (ver adapters/scanner); truncated
	// indica si se alcanzó el límite de cantidad y quedaron archivos
	// sin incluir, para que el core pueda avisarlo sin conocer el
	// número concreto del límite (ese detalle es del adapter).
	Walk(root string) (files []domain.ScannedFile, truncated bool, err error)

	// WriteContext escribe content como ARGOS.md en la raíz del
	// proyecto (root), sobrescribiendo por completo si ya existía (RF-09:
	// /init regenera desde cero, sin merge incremental).
	WriteContext(root, content string) error
}
