package main

import (
    "log"
    "net/http"
    "path/filepath"
    "strings"
)

func main() {
    mux := http.NewServeMux()

    // Serve index at / and /index.html
    mux.HandleFunc("/", indexHandler)

    // Serve only CSS under /static/css/
    mux.Handle("/static/css/", http.StripPrefix("/static/css/", cssHandler{}))

    addr := ":8080"
    log.Printf("Serving site (HTML & CSS only) at http://localhost%s", addr)
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatal(err)
    }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    // only allow root or explicit index.html
    p := r.URL.Path
    if p == "/" || p == "" || p == "/index.html" {
        http.ServeFile(w, r, "index.html")
        return
    }

    // prevent access to any other file types via root handler
    // (CSS requests are handled separately)
    http.NotFound(w, r)
}

type cssHandler struct{}

func (cssHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Protect against path traversal
    if strings.Contains(r.URL.Path, "..") {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Only allow .css files
    if !strings.HasSuffix(r.URL.Path, ".css") {
        http.Error(w, "Forbidden: only .css files are served", http.StatusForbidden)
        return
    }

    // Serve file from static/css/<name>.css
    name := filepath.Base(r.URL.Path)
    path := filepath.Join("static", "css", name)
    http.ServeFile(w, r, path)
}
