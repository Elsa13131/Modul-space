package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {
	if err := InitDB(); err != nil {
		log.Printf("⚠️ Erreur DB: %v", err)
	}
	defer CloseDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/dashboard", dashboardHandler)
	mux.HandleFunc("/api/quote", quoteHandler)
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

// InitDB initialise la base de données
func InitDB() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("DATABASE_URL non définie, mode sans base de données")
		return nil
	}

	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("erreur connexion: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("erreur ping: %v", err)
	}

	log.Println("✅ Connecté à PostgreSQL")

	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		nom VARCHAR(100),
		prenom VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_email ON users(email);
	
	CREATE TABLE IF NOT EXISTS quotes (
		id SERIAL PRIMARY KEY,
		nom VARCHAR(100) NOT NULL,
		prenom VARCHAR(100) NOT NULL,
		email VARCHAR(255) NOT NULL,
		telephone VARCHAR(20),
		produit VARCHAR(255) NOT NULL,
		message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_quote_email ON quotes(email);
	`

	if _, err := db.Exec(query); err != nil {
		return err
	}

	log.Println("✅ Tables users et quotes créées/vérifiées")
	return nil
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// User représente un utilisateur
type User struct {
	ID           int
	Email        string
	PasswordHash string
	Nom          string
	Prenom       string
}

// CreateUser crée un nouvel utilisateur
func CreateUser(email, password, nom, prenom string) error {
	if db == nil {
		return fmt.Errorf("base de données non configurée")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO users (email, password_hash, nom, prenom) VALUES ($1, $2, $3, $4)",
		email, string(hashedPassword), nom, prenom,
	)
	return err
}

// GetUserByEmail récupère un utilisateur par email
func GetUserByEmail(email string) (*User, error) {
	if db == nil {
		return nil, fmt.Errorf("base de données non configurée")
	}

	user := &User{}
	err := db.QueryRow(
		"SELECT id, email, password_hash, nom, prenom FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nom, &user.Prenom)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// VerifyPassword vérifie le mot de passe
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetUserFromSession récupère l'utilisateur connecté
func GetUserFromSession(r *http.Request) *User {
	cookie, err := r.Cookie("user_email")
	if err != nil {
		return nil
	}

	user, err := GetUserByEmail(cookie.Value)
	if err != nil {
		return nil
	}
	return user
}

// Quote représente une demande de devis
type Quote struct {
	ID        int
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Email     string `json:"email"`
	Telephone string `json:"telephone"`
	Produit   string `json:"produit"`
	Message   string `json:"message"`
}

// CreateQuote enregistre une demande de devis
func CreateQuote(nom, prenom, email, telephone, produit, message string) error {
	if db == nil {
		return fmt.Errorf("base de données non configurée")
	}

	_, err := db.Exec(
		"INSERT INTO quotes (nom, prenom, email, telephone, produit, message) VALUES ($1, $2, $3, $4, $5, $6)",
		nom, prenom, email, telephone, produit, message,
	)
	return err
}

// Handlers

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

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "templates/register.html")
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		nom := r.FormValue("nom")
		prenom := r.FormValue("prenom")

		if email == "" || password == "" {
			http.Error(w, "Email et mot de passe requis", http.StatusBadRequest)
			return
		}

		existing, _ := GetUserByEmail(email)
		if existing != nil {
			http.Error(w, "Cet email existe déjà", http.StatusBadRequest)
			return
		}

		if err := CreateUser(email, password, nom, prenom); err != nil {
			http.Error(w, "Erreur création compte", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "user_email",
			Value: email,
			Path:  "/",
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "templates/login.html")
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := GetUserByEmail(email)
		if err != nil || user == nil || !VerifyPassword(user.PasswordHash, password) {
			http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "user_email",
			Value: email,
			Path:  "/",
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "user_email",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	html := `<!DOCTYPE html>
<html lang="fr">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Dashboard - MODULSPACE</title>
	<link rel="stylesheet" href="/static/css/style.css?v=13">
</head>
<body>
	<header class="site-header">
		<div class="container">
			<img src="/static/img/logo.png" alt="Logo" class="site-logo">
		</div>
	</header>

	<main class="container" style="padding: 2rem;">
		<h1>Bienvenue ` + user.Prenom + `!</h1>
		<p>Email: ` + user.Email + `</p>
		<p>Nom: ` + user.Nom + ` ` + user.Prenom + `</p>
		<a href="/logout" style="padding: 0.5rem 1rem; background: #6161AB; color: white; text-decoration: none; border-radius: 6px;">Déconnexion</a>
	</main>

	<footer class="site-footer">
		<div class="container">© 2025 Ydays</div>
	</footer>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var quote Quote
	if err := json.NewDecoder(r.Body).Decode(&quote); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validation
	if quote.Nom == "" || quote.Prenom == "" || quote.Email == "" || quote.Produit == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Enregistrer dans la base de données
	if err := CreateQuote(quote.Nom, quote.Prenom, quote.Email, quote.Telephone, quote.Produit, quote.Message); err != nil {
		log.Printf("Erreur création devis: %v", err)
		http.Error(w, "Error saving quote", http.StatusInternalServerError)
		return
	}

	// Envoyer l'email
	if err := SendQuoteEmail(quote.Nom, quote.Prenom, quote.Email, quote.Telephone, quote.Produit); err != nil {
		log.Printf("Erreur envoi email: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Demande de devis enregistrée"})
}

// SendQuoteEmail envoie un email de demande de devis
func SendQuoteEmail(nom, prenom, email, telephone, produit string) error {
	// Configuration SMTP (utilise des variables d'environnement ou valeurs par défaut)
	smtpHost := getEnv("SMTP_HOST", "smtp.gmail.com")
	smtpPort := getEnv("SMTP_PORT", "587")
	smtpUser := getEnv("SMTP_USER", "")
	smtpPass := getEnv("SMTP_PASS", "")

	// Destinataire
	to := "elsachochon13@gmail.com"

	// Construction du message
	subject := fmt.Sprintf("Demande de devis - %s", produit)
	body := fmt.Sprintf(`Bonjour,

J'aimerais demander un devis pour le produit : %s

Mes coordonnées :
- Nom : %s
- Prénom : %s
- Email : %s
- Téléphone : %s

Merci de me renvoyer le devis pour ce produit.

Cordialement,
%s %s`, produit, nom, prenom, email, telephone, prenom, nom)

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		smtpUser, to, subject, body)

	// Si pas de config SMTP, on log juste (mode dev)
	if smtpUser == "" || smtpPass == "" {
		fmt.Printf("MODE DEV: Email qui serait envoyé:\n%s\n", message)
		return nil
	}

	// Authentification et envoi
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("erreur envoi email: %v", err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
