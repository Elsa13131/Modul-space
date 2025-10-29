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
    mux.Handle("/static/css/", http.StripPrefix("/static/css/", cssHandler{}))

    // Récupère le port attribué par Render
    port := os.Getenv("PORT")
    if port == "" {
        port = "10000" // pour test local
    }
    addr := ":" + port

    log.Printf("Serving site (HTML & CSS only) at http://localhost%s", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatal(err)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    p := r.URL.Path
    if p == "/" || p == "" || p == "/index.html" {
        http.ServeFile(w, r, "index.html")
        return
    }
    http.NotFound(w, r)
}

type cssHandler struct{}

func (cssHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if strings.Contains(r.URL.Path, "..") {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    if !strings.HasSuffix(r.URL.Path, ".css") {
        http.Error(w, "Forbidden: only .css files are served", http.StatusForbidden)
        return
    }
    name := filepath.Base(r.URL.Path)
    path := filepath.Join("static", "css", name)
    http.ServeFile(w, r, path)
}
