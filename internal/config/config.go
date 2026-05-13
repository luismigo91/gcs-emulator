package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/backend"
)

type Config struct {
	Port            int
	DefaultProject  string
	StorageMode     backend.BackendMode
	StoragePath     string
	FlushInterval   time.Duration
	ServiceOverrides map[string]ServiceConfig
}

type ServiceConfig struct {
	StorageMode backend.BackendMode
	StoragePath string
}

func Load() *Config {
	cfg := &Config{
		Port:           9090,
		DefaultProject: "test-project",
		StorageMode:    backend.ModeMemory,
		StoragePath:    "./data",
		FlushInterval:  5 * time.Second,
		ServiceOverrides: make(map[string]ServiceConfig),
	}

	if port := os.Getenv("GCP_EMULATOR_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if project := os.Getenv("GCP_EMULATOR_DEFAULT_PROJECT"); project != "" {
		cfg.DefaultProject = project
	}

	if mode := os.Getenv("GCP_EMULATOR_STORAGE_MODE"); mode != "" {
		cfg.StorageMode = backend.BackendMode(strings.ToLower(mode))
	}

	if path := os.Getenv("GCP_EMULATOR_STORAGE_PATH"); path != "" {
		cfg.StoragePath = path
	}

	if interval := os.Getenv("GCP_EMULATOR_FLUSH_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			cfg.FlushInterval = d
		}
	}

	cfg.loadServiceOverrides()

	return cfg
}

func (c *Config) loadServiceOverrides() {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "GCP_EMULATOR_SERVICE_") {
			parts := strings.SplitN(strings.TrimPrefix(env, "GCP_EMULATOR_SERVICE_"), "=", 2)
			if len(parts) != 2 {
				continue
			}

			serviceName := strings.ToLower(parts[0])
			value := parts[1]

			if strings.HasPrefix(value, "mode:") {
				mode := strings.TrimPrefix(value, "mode:")
				if mode != "" {
					override := c.ServiceOverrides[serviceName]
					override.StorageMode = backend.BackendMode(mode)
					c.ServiceOverrides[serviceName] = override
				}
			}

			if strings.HasPrefix(value, "path:") {
				path := strings.TrimPrefix(value, "path:")
				if path != "" {
					override := c.ServiceOverrides[serviceName]
					override.StoragePath = path
					c.ServiceOverrides[serviceName] = override
				}
			}
		}
	}
}

func (c *Config) GetServiceConfig(serviceName string) (backend.BackendMode, string) {
	if override, exists := c.ServiceOverrides[serviceName]; exists {
		mode := override.StorageMode
		if mode == "" {
			mode = c.StorageMode
		}
		path := override.StoragePath
		if path == "" {
			path = c.StoragePath
		}
		return mode, path
	}
	return c.StorageMode, c.StoragePath
}
