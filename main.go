package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "path/filepath"
    "strings"
    "syscall"
    "time"
)

// Minimal static file server prepared for cloud deployment.
func main() {
    mux := http.NewServeMux()

    // Serve index.html at root (SPA friendly)
    mux.HandleFunc("/", indexHandler)

    // Serve everything under ./static at /static/
    fs := http.FileServer(http.Dir("static"))
    mux.Handle("/static/", http.StripPrefix("/static/", fs))

    // Serve everything under ./img at /img/ so images in project root are accessible
    imgFs := http.FileServer(http.Dir("img"))
    mux.Handle("/img/", http.StripPrefix("/img/", imgFs))

    // Serve fonts from ./fonts at /fonts/
    fontsFs := http.FileServer(http.Dir("fonts"))
    mux.Handle("/fonts/", http.StripPrefix("/fonts/", fontsFs))

    // No request logging: use mux directly so individual requests are not printed
    handler := mux

    // Port from env (Render/Railway) or fallback to 8080 for local dev
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    addr := ":" + port

    srv := &http.Server{
        Addr:    addr,
        Handler: handler,
    }

    // Graceful shutdown on SIGINT/SIGTERM
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        // Message demandé par l'utilisateur (affiche le port utilisé)
        log.Printf("Serveur Modul-space démarré sur http://localhost:%s", port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    <-stop
    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("graceful shutdown failed: %v", err)
    }
    log.Println("Server stopped")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    p := r.URL.Path
    // Serve root or index.html explicitly
    if p == "/" || p == "" || p == "/index.html" {
        http.ServeFile(w, r, "index.html")
        return
    }

    // Prevent directory traversal
    if strings.Contains(p, "..") {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // If the path looks like a file in static, try to serve it
    if strings.HasPrefix(p, "/static/") {
        // map /static/... to ./static/...
        name := filepath.Clean(p[len("/static/"):])
        path := filepath.Join("static", name)
        http.ServeFile(w, r, path)
        return
    }

    // Fallback: return index.html (useful for SPA routing)
    http.ServeFile(w, r, "index.html")
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}
