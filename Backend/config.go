package main

import (
	"os"
)

var (
	BaseURL   = os.Getenv("SUPABASE_URL")
	APIKey    = os.Getenv("SUPABASE_ANON_KEY")
	JWTSecret = os.Getenv("JWT_SECRET")
)