package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("static/img"))))
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Serveur Modul-space démarré sur http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if p == "/" || p == "" || p == "/index.html" {
		http.ServeFile(w, r, "templates/index.html")
		return
	}
	if strings.Contains(p, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if strings.HasSuffix(p, ".html") {
		name := filepath.Clean(strings.TrimPrefix(p, "/"))
		filePath := filepath.Join("templates", name)
		http.ServeFile(w, r, filePath)
		return
	}
	http.ServeFile(w, r, "templates/index.html")
}
