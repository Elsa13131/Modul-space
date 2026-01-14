package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

// InitDB initialise la connexion à la base de données
func InitDB() error {
	// Récupérer l'URL de connexion depuis les variables d'environnement
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("DATABASE_URL non définie, mode sans base de données")
		return nil
	}

	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("erreur connexion DB: %v", err)
	}

	// Tester la connexion
	if err = db.Ping(); err != nil {
		return fmt.Errorf("erreur ping DB: %v", err)
	}

	log.Println("✅ Connecté à la base de données PostgreSQL")

	// Créer les tables si elles n'existent pas
	if err = createTables(); err != nil {
		return fmt.Errorf("erreur création tables: %v", err)
	}

	return nil
}

// createTables crée les tables nécessaires
func createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS demandes_devis (
		id SERIAL PRIMARY KEY,
		nom VARCHAR(100) NOT NULL,
		prenom VARCHAR(100) NOT NULL,
		email VARCHAR(255) NOT NULL,
		telephone VARCHAR(20),
		produit VARCHAR(255) NOT NULL,
		date_creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		statut VARCHAR(50) DEFAULT 'nouveau'
	);

	CREATE INDEX IF NOT EXISTS idx_email ON demandes_devis(email);
	CREATE INDEX IF NOT EXISTS idx_date ON demandes_devis(date_creation);
	`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	log.Println("✅ Tables créées/vérifiées")
	return nil
}

// DemandeDevis représente une demande de devis
type DemandeDevis struct {
	ID           int
	Nom          string
	Prenom       string
	Email        string
	Telephone    string
	Produit      string
	DateCreation time.Time
	Statut       string
}

// SaveDemandeDevis enregistre une demande de devis dans la base
func SaveDemandeDevis(nom, prenom, email, telephone, produit string) error {
	if db == nil {
		log.Println("⚠️ Base de données non configurée, demande non enregistrée")
		return nil
	}

	query := `
		INSERT INTO demandes_devis (nom, prenom, email, telephone, produit)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := db.Exec(query, nom, prenom, email, telephone, produit)
	if err != nil {
		return fmt.Errorf("erreur enregistrement devis: %v", err)
	}

	log.Printf("✅ Demande de devis enregistrée: %s %s - %s", prenom, nom, produit)
	return nil
}

// GetAllDemandesDevis récupère toutes les demandes de devis
func GetAllDemandesDevis() ([]DemandeDevis, error) {
	if db == nil {
		return nil, fmt.Errorf("base de données non configurée")
	}

	query := `
		SELECT id, nom, prenom, email, telephone, produit, date_creation, statut
		FROM demandes_devis
		ORDER BY date_creation DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var demandes []DemandeDevis
	for rows.Next() {
		var d DemandeDevis
		err := rows.Scan(&d.ID, &d.Nom, &d.Prenom, &d.Email, &d.Telephone, &d.Produit, &d.DateCreation, &d.Statut)
		if err != nil {
			return nil, err
		}
		demandes = append(demandes, d)
	}

	return demandes, nil
}

// CloseDB ferme la connexion à la base de données
func CloseDB() {
	if db != nil {
		db.Close()
		log.Println("Base de données fermée")
	}
}
