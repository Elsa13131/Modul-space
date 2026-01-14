package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
)

// QuoteRequest représente une demande de devis
type QuoteRequest struct {
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Email     string `json:"email"`
	Telephone string `json:"telephone"`
	Produit   string `json:"produit"`
}

// IndexHandler gère la page d'accueil et les pages HTML
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	
	// Serve root or index.html explicitly
	if p == "/" || p == "" || p == "/index.html" {
		http.ServeFile(w, r, "templates/index.html")
		return
	}

	// Prevent directory traversal
	if strings.Contains(p, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// If the path looks like a file in static, try to serve it
	if strings.HasPrefix(p, "/static/") {
		name := filepath.Clean(p[len("/static/"):])
		path := filepath.Join("static", name)
		http.ServeFile(w, r, path)
		return
	}

	// Try to serve .html files from templates
	if strings.HasSuffix(p, ".html") {
		name := filepath.Clean(strings.TrimPrefix(p, "/"))
		filePath := filepath.Join("templates", name)
		http.ServeFile(w, r, filePath)
		return
	}

	// Fallback: return index.html
	http.ServeFile(w, r, "templates/index.html")
}

// QuoteHandler gère les demandes de devis
func QuoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validation basique
	if req.Nom == "" || req.Prenom == "" || req.Email == "" || req.Produit == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Envoi de l'email
	if err := SendQuoteEmail(req.Nom, req.Prenom, req.Email, req.Telephone, req.Produit); err != nil {
		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Demande de devis envoyée"})
}
