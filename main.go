package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AbhishekMashetty/url-shortner/handlers"
	"github.com/AbhishekMashetty/url-shortner/internal"
	"github.com/AbhishekMashetty/url-shortner/store"
)

func main() {
	connStr := "postgres://abhishekmasetty@localhost:5432/url_shortner?sslmode=disable"
	store, err := store.NewPostgresStore(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	handler := handlers.NewURLHandler(store)
	router := internal.SetupRouter(handler)

	fmt.Println("🚀 Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// var urlStore = make(map[string]string)

// const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// func main() {
// 	r := mux.NewRouter()

// 	r.HandleFunc("/", homeHandler).Methods("GET")
// 	r.HandleFunc("/shorten", shortenURLHandler).Methods("POST")
// 	r.HandleFunc("/{shortCode}", redirectHandler).Methods("GET")

// 	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		http.Error(w, "404 page not found: "+r.URL.Path, http.StatusNotFound)
// 	})

// 	fmt.Println("🚀 Server is running on http://localhost:8080")
// 	err := http.ListenAndServe(":8080", r)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func homeHandler(w http.ResponseWriter, r *http.Request) {
// 	html := `
// 		<!DOCTYPE html>
// 		<html>
// 		<head>
// 			<title>URL Shortener</title>
// 		</head>
// 		<body>
// 			<h1>🚀 URL Shortener</h1>
// 			<form action="/shorten" method="POST">
// 				<label for="url">Enter a URL to shorten:</label><br><br>
// 				<input type="text" id="url" name="url" style="width: 300px;" required><br><br>
// 				<input type="submit" value="Shorten">
// 			</form>
// 		</body>
// 		</html>
// 	`
// 	w.Header().Set("Content-Type", "text/html")
// 	fmt.Fprint(w, html)
// }

// func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
// 	err := r.ParseForm()
// 	if err != nil {
// 		http.Error(w, "Invalid form data", http.StatusBadRequest)
// 		return
// 	}

// 	originalURL := r.FormValue("url")
// 	if originalURL == "" {
// 		http.Error(w, "No URL Passed", http.StatusBadRequest)
// 		return
// 	}

// 	shortCode := generateShortCode()

// 	urlStore[shortCode] = originalURL

// 	shortURL := fmt.Sprintf("http://localhost:8080/%s", shortCode)
// 	fmt.Fprintf(w, "Shortened URL: %s\n", shortURL)
// }

// func generateShortCode() string {
// 	rand.Seed(time.Now().UnixNano())
// 	length := 6
// 	code := make([]byte, length)
// 	for i := range code {
// 		code[i] = charset[rand.Intn(len(charset))]
// 	}
// 	return string(code)
// }

// func redirectHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("🔍 redirectHandler was called")
// 	vars := mux.Vars(r)
// 	shortCode := vars["shortCode"]
// 	originalURL, ok := urlStore[shortCode]
// 	log.Println("originalURL:", originalURL, "ok:", ok)
// 	if !ok {
// 		http.Error(w, "Short URL not found", http.StatusFound)
// 	}

// 	http.Redirect(w, r, originalURL, http.StatusFound)
// }
