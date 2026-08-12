// Package ollama implementa ports.ModelProvider consultando la API
// HTTP local que expone Ollama (http://localhost:11434) cuando está
// corriendo. Es el único paquete del proyecto que sabe que "el backend
// de modelos" es concretamente Ollama.
package ollama

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pokydev/argos/internal/core/domain"
)

// defaultBaseURL es el host donde Ollama escucha por defecto en Windows
// (y en general) cuando se instala con su configuración estándar.
const defaultBaseURL = "http://localhost:11434"

// Client es la implementación de ports.ModelProvider para Ollama.
type Client struct {
	baseURL string
	http    *http.Client
}

// New crea un cliente de Ollama apuntando al host local por defecto.
func New() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// tagsResponse mapea la respuesta de GET /api/tags (modelos descargados).
type tagsResponse struct {
	Models []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"models"`
}

// psResponse mapea la respuesta de GET /api/ps (modelos cargados en
// memoria en este momento).
type psResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ListModels implementa ports.ModelProvider.
func (c *Client) ListModels() ([]domain.Model, error) {
	tags, err := c.fetchTags()
	if err != nil {
		return nil, err
	}

	// Si /api/ps falla, no es un error crítico: seguimos con la lista de
	// modelos instalados, simplemente sin poder marcar cuáles están
	// cargados en memoria.
	loaded, _ := c.fetchLoaded()

	models := make([]domain.Model, 0, len(tags.Models))
	for _, m := range tags.Models {
		models = append(models, domain.Model{
			Name:   m.Name,
			Size:   m.Size,
			Loaded: loaded[m.Name],
		})
	}
	return models, nil
}

func (c *Client) fetchTags() (tagsResponse, error) {
	var out tagsResponse

	resp, err := c.http.Get(c.baseURL + "/api/tags")
	if err != nil {
		return out, fmt.Errorf("no se pudo conectar con Ollama en %s (¿está corriendo? probá 'ollama serve'): %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return out, fmt.Errorf("Ollama respondió con estado inesperado: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, fmt.Errorf("respuesta de Ollama inválida: %w", err)
	}

	return out, nil
}

func (c *Client) fetchLoaded() (map[string]bool, error) {
	resp, err := c.http.Get(c.baseURL + "/api/ps")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("estado inesperado: %d", resp.StatusCode)
	}

	var out psResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	loaded := make(map[string]bool, len(out.Models))
	for _, m := range out.Models {
		loaded[m.Name] = true
	}
	return loaded, nil
}