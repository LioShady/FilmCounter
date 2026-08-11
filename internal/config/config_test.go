package config

import (
	"strings"
	"testing"
)

func TestNew_Success(t *testing.T) {
	t.Setenv("DB_NAME", "mydb")
	t.Setenv("DB_PASSWORD", "secretpass")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("POISK_KINO_API_KEY", "secret-key")

	cfg, err := New()
	if err != nil {
		t.Fatalf("New returned an error: %s", err.Error())
	}

	if cfg.DBNAME != "mydb" {
		t.Fatalf("DBNAME returned %q, want %q", cfg.DBNAME, "mydb")
	}
	if cfg.DBPASS != "secretpass" {
		t.Fatalf("DBPASS returned %q, want %q", cfg.DBPASS, "secretpass")
	}
	if cfg.DBHOST != "localhost" {
		t.Fatalf("DBHOST returned %q, want %q", cfg.DBHOST, "localhost")
	}
	if cfg.DBPORT != "5432" {
		t.Fatalf("DBPORT returned %q, want %q", cfg.DBPORT, "5432")
	}
	if cfg.DBUSER != "postgres" {
		t.Fatalf("DBUSER returned %q, want %q", cfg.DBUSER, "postgres")
	}
	if cfg.PoiskKinoApiKey != "secret-key" {
		t.Fatalf("PoiskKinoApiKey returned %q, want %q", cfg.PoiskKinoApiKey, "secret-key")
	}
}

func TestNew_MissingEnv(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		wantError string
	}{
		{
			name: "Missing DB_NAME",
			envVars: map[string]string{
				"DB_PASSWORD":        "secretpass",
				"DB_HOST":            "localhost",
				"DB_PORT":            "5432",
				"DB_USER":            "postgres",
				"POISK_KINO_API_KEY": "secret-key",
			},
			wantError: "DB_NAME",
		},
		{
			name: "Missing DB_PASSWORD",
			envVars: map[string]string{
				"DB_NAME":            "mydb",
				"DB_HOST":            "localhost",
				"DB_PORT":            "5432",
				"DB_USER":            "postgres",
				"POISK_KINO_API_KEY": "secret-key",
			},
			wantError: "DB_PASSWORD",
		},
		{
			name: "Missing DB_HOST",
			envVars: map[string]string{
				"DB_NAME":            "mydb",
				"DB_PASSWORD":        "secretpass",
				"DB_PORT":            "5432",
				"DB_USER":            "postgres",
				"POISK_KINO_API_KEY": "secret-key",
			},
			wantError: "DB_HOST",
		},
		{
			name: "Missing DB_PORT",
			envVars: map[string]string{
				"DB_NAME":            "mydb",
				"DB_PASSWORD":        "secretpass",
				"DB_HOST":            "localhost",
				"DB_USER":            "postgres",
				"POISK_KINO_API_KEY": "secret-key",
			},
			wantError: "DB_PORT",
		},
		{
			name: "Missing DB_USER",
			envVars: map[string]string{
				"DB_NAME":            "mydb",
				"DB_PASSWORD":        "secretpass",
				"DB_HOST":            "localhost",
				"DB_PORT":            "5432",
				"POISK_KINO_API_KEY": "secret-key",
			},
			wantError: "DB_USER",
		},
		{
			name: "Missing POISK_KINO_API_KEY",
			envVars: map[string]string{
				"DB_NAME":     "mydb",
				"DB_PASSWORD": "secretpass",
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "postgres",
			},
			wantError: "POISK_KINO_API_KEY",
		},
		{
			name: "Empty DB_NAME should fail",
			envVars: map[string]string{
				"DB_NAME":            "",
				"DB_PASSWORD":        "secretpass",
				"DB_HOST":            "localhost",
				"DB_PORT":            "5432",
				"DB_USER":            "postgres",
				"POISK_KINO_API_KEY": "secret-key",
			},
			wantError: "DB_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			cfg, err := New()
			if err == nil {
				t.Fatalf("New() error == nil, want error")
			}
			if cfg != nil {
				t.Fatalf("New() returned cfg != nil, want nil")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("New() returned err %q, want it to contain %q", err.Error(), tt.wantError)
			}
		})
	}
}
