// Package config se encarga de leer archivos de configuración JSON y convertirlos a structs Go.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ServiceConfig describe un servicio a monitorear, tal como viene en el JSON de configuración.
type ServiceConfig struct {
	Type           string `json:"type"` // "http" | "docker" | "tcp" | "ping"
	Name           string `json:"name"`
	Target         string `json:"target"` // url / nombre de contenedor / host:puerto / host
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// Config es la configuración completa del programa.
type Config struct {
	PollInterval string          `json:"poll_interval"` // ej. "10s"
	SerialPort   string          `json:"serial_port"`   // ej. "/dev/ttyACM0" (Linux/Mac) o "COM3" (Windows)
	BaudRate     int             `json:"baud_rate"`
	Services     []ServiceConfig `json:"services"`
}

// PollIntervalDuration convierte el string de intervalo (ej. "10s") a time.Duration.
func (c *Config) PollIntervalDuration() (time.Duration, error) {
	return time.ParseDuration(c.PollInterval)
}

// Load lee y parsea el archivo de configuración JSON desde disco.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config inválida: %w", err)
	}
	if cfg.BaudRate == 0 {
		cfg.BaudRate = 115200
	}
	return &cfg, nil
}
