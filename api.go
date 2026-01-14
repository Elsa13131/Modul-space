package main

import (
	"encoding/json"
	"net/http"
)

// QuoteRequest représente une demande de devis
type QuoteRequest struct {
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Email     string `json:"email"`
	Telephone string `json:"telephone"`
	Produit   string `json:"produit"`
}

// QuoteHandler gère les demandes de devis (enregistrement dans la DB)
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

	// Enregistrer dans la base de données
	if err := SaveDemandeDevis(req.Nom, req.Prenom, req.Email, req.Telephone, req.Produit); err != nil {
		http.Error(w, "Error saving quote request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Demande de devis enregistrée",
	})
}

// AdminDevisHandler affiche toutes les demandes de devis (page admin)
func AdminDevisHandler(w http.ResponseWriter, r *http.Request) {
	demandes, err := GetAllDemandesDevis()
	if err != nil {
		http.Error(w, "Error fetching quotes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(demandes)
}
