package main

import (
	"os"
	"strings"
)

func IsProduction() bool {
	return os.Getenv("ENV") == "production"
}

func GetEnvList(name string) []string {
	raw := os.Getenv(name)
	if raw == "" {
		return []string{}
	}

	parts := strings.Split(raw, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
