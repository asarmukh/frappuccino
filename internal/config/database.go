package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Подключение к базе данных с тайм-аутом
func ConnectDB() *sql.DB {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		GetEnv("DB_HOST", "localhost"),
		GetEnv("DB_USER", "latte"),
		GetEnv("DB_PASSWORD", "latte"),
		GetEnv("DB_NAME", "frappuccino"),
		GetEnv("DB_PORT", "5432"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Database connection error:", err)
	}

	// Ожидание доступности базы данных с тайм-аутом
	waitForDB(db)
	fmt.Println("✅ Connected to PostgreSQL")
	return db
}

// Ожидание доступности базы данных
func waitForDB(db *sql.DB) {
	timeout := time.After(30 * time.Second)
	tick := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			log.Fatal("❌ Тайм-аут подключения к БД")
		case <-tick:
			if err := db.Ping(); err == nil {
				return
			}
			fmt.Println("⏳ Ожидание подключения к БД...")
		}
	}
}
