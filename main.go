package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	memoryStore = make(map[string]string)
	mu          sync.RWMutex
	db          *sql.DB
	useDB       bool
)

const urlChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Генерация сокращённого URL
func generateShortURL() (string, error) {
	length := 7
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(urlChars))))
		if err != nil {
			return "", err
		}
		result[i] = urlChars[num.Int64()]
	}
	return string(result), nil
}

// Обработчик POST запросов для создания сокращённого URL
func createShortURLHandler(w http.ResponseWriter, r *http.Request) {
	var originalURL string
	if err := json.NewDecoder(r.Body).Decode(&originalURL); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Генерация сокращённого URL
	shortURL, err := generateShortURL()
	if err != nil {
		http.Error(w, "Failed to generate short URL", http.StatusInternalServerError)
		return
	}

	if useDB {
		// Сохранение в базе данных
		_, err = db.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", shortURL, originalURL)
		if err != nil {
			http.Error(w, "Failed to save URL in database", http.StatusInternalServerError)
			return
		}
	} else {
		// Сохранение в памяти
		mu.Lock()
		memoryStore[shortURL] = originalURL
		mu.Unlock()
	}

	// Возвращаем сокращённый URL
	w.WriteHeader(http.StatusOK)
	log.Printf("Saving URL: %s -> %s", shortURL, originalURL)
	fmt.Fprintf(w, "http://localhost:8080/s/%s", shortURL)
}

// Обработчик GET запросов для получения оригинального URL по сокращённому
func getOriginalURLHandler(w http.ResponseWriter, r *http.Request) {

	var originalURL string
	shortURL := r.URL.Path[len("/s/"):]

	if useDB {
		// Получение из базы данных
		err := db.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortURL).Scan(&originalURL)
		if err == sql.ErrNoRows {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Failed to retrieve URL from database", http.StatusInternalServerError)
			return
		}
	} else {
		// Получение из памяти
		mu.RLock()
		originalURL, exists := memoryStore[shortURL]
		mu.RUnlock()
		if !exists && originalURL == "" {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

	}
	originalURL = memoryStore[shortURL]

	// Возвращаем оригинальный URL
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", originalURL)
}

// Инициализация подключения к базе данных
func initDB() {
	var err error
	connStr := "user=postgres dbname=urls sslmode=disable password=password"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Создание таблицы, если её нет
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			short_url VARCHAR(7) PRIMARY KEY,
			original_url TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

// Основная функция
func main() {
	// Определение использования базы данных на основе флага запуска
	useDB = len(os.Args) > 1 && os.Args[1] == "-d"
	if useDB {
		initDB()
		defer db.Close()
	}

	http.HandleFunc("/create", createShortURLHandler)
	http.HandleFunc("/s/", getOriginalURLHandler)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
