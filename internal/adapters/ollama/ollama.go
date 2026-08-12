// Package ollama implementa ports.ModelProvider consultando la API
// HTTP local que expone Ollama (http://localhost:11434) cuando está
// corriendo. Es el único paquete del proyecto que sabe que "el backend
// de modelos" es concretamente Ollama.
package ollama

import (
	"bytes"
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
	http    *http.Client // timeout corto, para /api/tags y /api/ps
	genHTTP *http.Client // timeout largo, para /api/generate, posible demora
}

// New crea un cliente de Ollama apuntando al host local por defecto.
func New() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
		genHTTP: &http.Client{Timeout: 5 * time.Minute},
	}
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type generateErrorResponse struct {
	Error string `json:"error"`
}

func (c *Client) Generate(model, prompt string) (string, error) {
	if model == "" {
		return "", fmt.Errorf("no hay un modelo activo (usá /model <nombre>)")
	}

	body, err := json.Marshal(generateRequest{Model: model, Prompt: prompt, Stream: false})
	if err != nil {
		return "", fmt.Errorf("no se pudo preparar la solicitud: %w", err)
	}

	resp, err := c.genHTTP.Post(c.baseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("no se pudo conectar con Ollama en %s (¿está corriendo? probá /ollama init): %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr generateErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Error != "" {
			return "", fmt.Errorf("Ollama respondió con error: %s", apiErr.Error)
		}
		return "", fmt.Errorf("Ollama respondió con estado inesperado: %d", resp.StatusCode)
	}

	var out generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("respuesta de Ollama inválida: %w", err)
	}

	return out.Response, nil
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
