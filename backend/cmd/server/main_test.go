package main

import (
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type stubConfig struct {
	ports.ConfigProvider
	password string
}

func (s stubConfig) GetPassword() string { return s.password }

type stubUsers struct {
	ports.UserRepository
	count int
}

func (s stubUsers) CountUsers() (int, error) { return s.count, nil }

func TestRefuseWeakAdminSeed(t *testing.T) {
	cases := []struct {
		name     string
		appEnv   string
		password string
		users    int
		wantErr  bool
	}{
		{"production weak password fresh install", "production", "changeme", 0, true},
		{"production default admin fresh install", "production", "admin", 0, true},
		{"production strong password fresh install", "production", "a-long-unique-passphrase", 0, false},
		{"production weak password existing install", "production", "changeme", 3, false},
		{"development weak password", "development", "admin", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tc.appEnv)
			err := refuseWeakAdminSeed(stubConfig{password: tc.password}, stubUsers{count: tc.users})
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "ADMIN_PASSWORD") {
					t.Errorf("error %q should mention ADMIN_PASSWORD", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
