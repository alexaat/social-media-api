package main

import (
	"github.com/google/uuid"
)

func generateSessionId() string {
	return uuid.New().String()
}
