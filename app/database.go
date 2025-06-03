package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func GetConnection() *sql.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")

	log.Printf("Mencoba koneksi ke database: %v", dbName)

	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbHost, dbPort, dbName)

	dbUrl := os.Getenv("DB_URL")

	if dbUrl != "" {
	    log.Printf("Mencoba koneksi ke database DB_URL")
		connStr = dbUrl
	}


	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("Error membuka koneksi database: %v", err)
	}

	// Set konfigurasi koneksi
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(100)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(60 * time.Minute)

	// Cek koneksi dengan timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		log.Printf("Gagal terhubung ke database dalam 10 detik: %v", err)
		log.Printf("Mencoba koneksi ulang...")

		// Coba lagi dengan timeout yang lebih lama
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			db.Close()
			log.Fatalf("Koneksi database gagal setelah percobaan ulang: %v", err)
		}
	}

	log.Printf("Berhasil terhubung ke database %s", dbName)
	log.Printf("Max Open Connections: %d", db.Stats().MaxOpenConnections)
	log.Printf("Open Connections: %d", db.Stats().OpenConnections)
	log.Printf("In Use Connections: %d", db.Stats().InUse)
	log.Printf("Idle Connections: %d", db.Stats().Idle)

	return db
}
