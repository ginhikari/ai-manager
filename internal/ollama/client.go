package ollama

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ModelInfo struct {
	Name         string    `json:"name"`
	Size         uint64    `json:"size"`
	Digest       string    `json:"digest"`
	ModifiedAt   time.Time `json:"modified_at"`
	Quantization string    `json:"quantization"`
	Loaded       bool      `json:"loaded"`
	ProcessID    string    `json:"process_id,omitempty"`
}

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(host string, port int) *Client {
	return &Client{
		BaseURL: fmt.Sprintf("http://%s:%d", host, port),
		HTTP:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) IsServerRunning() bool {
	resp, err := c.HTTP.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (c *Client) ListModels() ([]ModelInfo, error) {
	resp, err := c.HTTP.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name       string `json:"name"`
			Size       uint64 `json:"size"`
			Digest     string `json:"digest"`
			ModifiedAt string `json:"modified_at"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var models []ModelInfo
	for _, m := range result.Models {
		models = append(models, ModelInfo{
			Name:         m.Name,
			Size:         m.Size,
			Digest:       m.Digest,
			ModifiedAt:   parseTime(m.ModifiedAt),
			Quantization: extractQuantization(m.Name),
		})
	}
	return models, nil
}

func (c *Client) LoadModel(name string) error {
	req, err := http.NewRequest("POST", c.BaseURL+"/api/load", strings.NewReader(`{"name":"`+name+`"}`))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to load model: %s", string(body))
	}
	return nil
}

func (c *Client) UnloadModel(name string) error {
	req, err := http.NewRequest("POST", c.BaseURL+"/api/generate", strings.NewReader(`{"model":"`+name+`","prompt":" unload","stream":false}`))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) DiscoverModels(modelDir string) ([]ModelInfo, error) {
	ollamaModels, err := c.ListModels()
	if err != nil {
		ollamaModels = []ModelInfo{}
	}

	manifestDir := filepath.Join(modelDir, "manifests")
	var diskModels []ModelInfo

	entries, err := os.ReadDir(manifestDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				manifestPath := filepath.Join(manifestDir, entry.Name(), "latest")
				data, err := os.ReadFile(manifestPath)
				if err != nil {
					continue
				}
				var manifest struct {
					RepoDigests []string `json:"RepoDigests"`
				}
				if err := json.Unmarshal(data, &manifest); err != nil {
					continue
				}
				for _, digest := range manifest.RepoDigests {
					parts := strings.Split(digest, "/")
					name := parts[len(parts)-1]
					diskModels = append(diskModels, ModelInfo{
						Name:         name,
						Quantization: extractQuantization(name),
						Loaded:       false,
					})
				}
			}
		}
	}

	seen := make(map[string]bool)
	var result []ModelInfo
	for _, m := range ollamaModels {
		if !seen[m.Name] {
			result = append(result, m)
			seen[m.Name] = true
		}
	}
	for _, m := range diskModels {
		if !seen[m.Name] {
			result = append(result, m)
			seen[m.Name] = true
		}
	}

	return result, nil
}

func (c *Client) RunOllamaCommand(args ...string) (string, error) {
	cmd := exec.Command("ollama", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", s)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}

func extractQuantization(name string) string {
	parts := strings.Split(name, ":")
	if len(parts) > 1 {
		quant := parts[len(parts)-1]
		if strings.HasPrefix(quant, "Q") {
			return quant
		}
	}
	return "unknown"
}
