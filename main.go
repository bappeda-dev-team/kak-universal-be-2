package main

import (
	"ekak_kabupaten_madiun/helper"
	"ekak_kabupaten_madiun/middleware"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func NewServer(authMiddleware *middleware.AuthMiddleware) *http.Server {
	host := os.Getenv("host")
	port := os.Getenv("port")
	addr := fmt.Sprintf("%s:%s", host, port)

	cors := helper.NewCORSMiddleware()

	if addr == ":" {
		addr = "localhost:8080"
	}

	return &http.Server{
		Addr:    addr,
		Handler: cors.Handler(authMiddleware),
	}
}

func main() {
	flag.Parse()
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	jwksURL := os.Getenv("KEYCLOAK_JWKS_URL") // e.g. https://keycloak.local/realms/myrealm/protocol/openid-connect/certs
	if jwksURL == "" {
		log.Fatal("Error KEYCLOAK JWKS NOT FOUND")
	}
	erru := middleware.InitJWKS(jwksURL)
	if erru != nil {
		log.Fatalf("Failed to initialize JWKS: %v", erru)
	}

	// Initialize dan jalankan server
	server := InitializeServer()
	log.Printf("Server berjalan di %s", server.Addr)
	err = server.ListenAndServe()
	helper.PanicIfError(err)
}
