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

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var dbDriver string

func main() {
	_ = godotenv.Load()

	if err := InitDB(); err != nil {
		log.Printf("⚠️ Erreur DB: %v", err)
	}
	defer CloseDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/api/quote", quoteHandler)
	mux.HandleFunc("/api/user", userHandler)
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

// InitDB initialise la base de données PostgreSQL (Scalingo) ou MySQL local en fallback
func InitDB() error {
	var err error
	postgresDSN := os.Getenv("DATABASE_URL")
	if postgresDSN != "" {
		dbDriver = "postgres"
		db, err = sql.Open("postgres", postgresDSN)
		if err != nil {
			return fmt.Errorf("erreur connexion PostgreSQL: %v", err)
		}

		if err = db.Ping(); err != nil {
			return fmt.Errorf("erreur ping PostgreSQL: %v", err)
		}

		if err := createTablesPostgres(); err != nil {
			return err
		}

		log.Println("✅ Connecté à PostgreSQL (DATABASE_URL)")
		log.Println("✅ Tables users et quotes créées/vérifiées")
		return nil
	}

	dbDriver = "mysql"
	rootDSN := getEnv("MYSQL_ROOT_DSN", "root:@tcp(localhost:3306)/?parseTime=true")
	tempDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("erreur connexion MySQL: %v", err)
	}
	defer tempDB.Close()

	if _, err := tempDB.Exec("CREATE DATABASE IF NOT EXISTS modulspace"); err != nil {
		return fmt.Errorf("erreur création base de données: %v", err)
	}

	mysqlDSN := getEnv("MYSQL_DSN", "root:@tcp(localhost:3306)/modulspace?parseTime=true")
	db, err = sql.Open("mysql", mysqlDSN)
	if err != nil {
		return fmt.Errorf("erreur connexion MySQL modulspace: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("erreur ping MySQL: %v", err)
	}

	if err := createTablesMySQL(); err != nil {
		return err
	}

	log.Println("✅ Connecté à MySQL local")
	log.Println("✅ Tables users et quotes créées/vérifiées")
	return nil
}

func createTablesMySQL() error {
	queryUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		nom VARCHAR(100),
		prenom VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(queryUsers); err != nil {
		return fmt.Errorf("erreur création table users: %v", err)
	}

	queryQuotes := `
	CREATE TABLE IF NOT EXISTS quotes (
		id INT AUTO_INCREMENT PRIMARY KEY,
		nom VARCHAR(100) NOT NULL,
		prenom VARCHAR(100) NOT NULL,
		email VARCHAR(255) NOT NULL,
		telephone VARCHAR(20),
		produit VARCHAR(255) NOT NULL,
		message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(queryQuotes); err != nil {
		return fmt.Errorf("erreur création table quotes: %v", err)
	}

	return nil
}

func createTablesPostgres() error {
	queryUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		nom VARCHAR(100),
		prenom VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(queryUsers); err != nil {
		return fmt.Errorf("erreur création table users: %v", err)
	}

	queryQuotes := `
	CREATE TABLE IF NOT EXISTS quotes (
		id SERIAL PRIMARY KEY,
		nom VARCHAR(100) NOT NULL,
		prenom VARCHAR(100) NOT NULL,
		email VARCHAR(255) NOT NULL,
		telephone VARCHAR(20),
		produit VARCHAR(255) NOT NULL,
		message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(queryQuotes); err != nil {
		return fmt.Errorf("erreur création table quotes: %v", err)
	}

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
		func() string {
			if dbDriver == "postgres" {
				return "INSERT INTO users (email, password_hash, nom, prenom) VALUES ($1, $2, $3, $4)"
			}
			return "INSERT INTO users (email, password_hash, nom, prenom) VALUES (?, ?, ?, ?)"
		}(),
		email, string(hashedPassword), nom, prenom,
	)
	if err != nil {
		log.Printf("Erreur insertion utilisateur %s: %v", email, err)
	}
	return err
}

// GetUserByEmail récupère un utilisateur par email
func GetUserByEmail(email string) (*User, error) {
	if db == nil {
		return nil, fmt.Errorf("base de données non configurée")
	}

	user := &User{}
	err := db.QueryRow(
		func() string {
			if dbDriver == "postgres" {
				return "SELECT id, email, password_hash, nom, prenom FROM users WHERE email = $1"
			}
			return "SELECT id, email, password_hash, nom, prenom FROM users WHERE email = ?"
		}(),
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
		func() string {
			if dbDriver == "postgres" {
				return "INSERT INTO quotes (nom, prenom, email, telephone, produit, message) VALUES ($1, $2, $3, $4, $5, $6)"
			}
			return "INSERT INTO quotes (nom, prenom, email, telephone, produit, message) VALUES (?, ?, ?, ?, ?, ?)"
		}(),
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

		// Validation
		errors := make(map[string]string)
		
		if email == "" {
			errors["email"] = "Email requis"
		}
		if password == "" {
			errors["password"] = "Mot de passe requis"
		}
		if nom == "" {
			errors["nom"] = "Nom requis"
		}
		if prenom == "" {
			errors["prenom"] = "Prénom requis"
		}

		existing, _ := GetUserByEmail(email)
		if existing != nil {
			errors["email"] = "Cet email existe déjà"
		}

		// Si erreurs, afficher le formulaire avec erreurs
		if len(errors) > 0 {
			registerFormWithErrors(w, email, nom, prenom, errors)
			return
		}

		if err := CreateUser(email, password, nom, prenom); err != nil {
			errors["general"] = "Erreur création compte"
			registerFormWithErrors(w, email, nom, prenom, errors)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "user_email",
			Value: email,
			Path:  "/",
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func registerFormWithErrors(w http.ResponseWriter, email, nom, prenom string, errors map[string]string) {
	html := `<!DOCTYPE html>
<html lang="fr">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Inscription - MODULSPACE</title>
	<link rel="stylesheet" href="/static/css/style.css?v=31">
	<style>
		.auth-container {
			max-width: 450px;
			margin: 3rem auto;
			padding: 2.5rem;
			background: white;
			border-radius: 12px;
			box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
		}
		.auth-container h2 {
			text-align: center;
			color: #6161AB;
			margin-bottom: 2rem;
			font-size: 1.8rem;
		}
		.form-error {
			background: #ffebee;
			border-left: 4px solid #ef5350;
			color: #c62828;
			padding: 1rem;
			border-radius: 4px;
			margin-bottom: 1.5rem;
			font-weight: 500;
			display: flex;
			align-items: center;
			gap: 0.5rem;
		}
		.form-error::before {
			content: '⚠';
			font-size: 1.2rem;
		}
		.form-group {
			margin-bottom: 1.5rem;
		}
		.form-group label {
			display: block;
			margin-bottom: 0.5rem;
			font-weight: 600;
			color: #333;
		}
		.form-group input {
			width: 100%;
			padding: 0.75rem;
			border: 2px solid #e0e0e0;
			border-radius: 6px;
			font-size: 1rem;
			box-sizing: border-box;
			transition: border-color 0.3s;
		}
		.form-group input:focus {
			outline: none;
			border-color: #6161AB;
		}
		.form-group input.error {
			border-color: #ef5350;
		}
		.field-error {
			display: block;
			color: #ef5350;
			font-size: 0.85rem;
			margin-top: 0.4rem;
			font-weight: 500;
		}
		.btn-submit {
			width: 100%;
			padding: 0.9rem;
			background: #6161AB;
			color: white;
			border: none;
			border-radius: 6px;
			font-size: 1rem;
			font-weight: 600;
			cursor: pointer;
			transition: all 0.3s;
		}
		.btn-submit:hover {
			background: #4d4d8f;
			transform: translateY(-2px);
			box-shadow: 0 4px 12px rgba(97, 97, 171, 0.3);
		}
		.auth-link {
			text-align: center;
			margin-top: 1.5rem;
			color: #666;
		}
		.auth-link a {
			color: #6161AB;
			text-decoration: none;
			font-weight: 600;
		}
		.auth-link a:hover {
			text-decoration: underline;
		}
	</style>
</head>
<body>
	<header class="site-header">
		<div class="container">
			<span class="header-welcome">BIENVENUE</span>
			<img src="/static/img/logo.png" alt="Logo" class="site-logo">
		</div>
	</header>

	<div class="sub-banner">
		<div class="container">
			<nav>
				<ul>
					<li><a href="/">Accueil</a></li>
					<li><a href="/apropos.html">A Propos</a></li>
					<li><a href="/produit.html">Produit</a></li>
					<li><a href="/contact.html">Contact</a></li>
				</ul>
			</nav>
		</div>
	</div>

	<main class="container">
		<div class="auth-container">
			<h2>Inscription</h2>`

	if errors["general"] != "" {
		html += `<div class="form-error">` + errors["general"] + `</div>`
	}

	html += `<form method="POST">
				<div class="form-group">
					<label for="nom">Nom</label>
					<input type="text" id="nom" name="nom" value="` + nom + `" class="` + 
					(func() string {
						if errors["nom"] != "" {
							return "error"
						}
						return ""
					})() + `">` + 
					(func() string {
						if errors["nom"] != "" {
							return `<span class="field-error">` + errors["nom"] + `</span>`
						}
						return ""
					})() + `
				</div>

				<div class="form-group">
					<label for="prenom">Prénom</label>
					<input type="text" id="prenom" name="prenom" value="` + prenom + `" class="` + 
					(func() string {
						if errors["prenom"] != "" {
							return "error"
						}
						return ""
					})() + `">` + 
					(func() string {
						if errors["prenom"] != "" {
							return `<span class="field-error">` + errors["prenom"] + `</span>`
						}
						return ""
					})() + `
				</div>

				<div class="form-group">
					<label for="email">Email</label>
					<input type="email" id="email" name="email" value="` + email + `" class="` + 
					(func() string {
						if errors["email"] != "" {
							return "error"
						}
						return ""
					})() + `">` + 
					(func() string {
						if errors["email"] != "" {
							return `<span class="field-error">` + errors["email"] + `</span>`
						}
						return ""
					})() + `
				</div>

				<div class="form-group">
					<label for="password">Mot de passe</label>
					<input type="password" id="password" name="password" class="` + 
					(func() string {
						if errors["password"] != "" {
							return "error"
						}
						return ""
					})() + `">` + 
					(func() string {
						if errors["password"] != "" {
							return `<span class="field-error">` + errors["password"] + `</span>`
						}
						return ""
					})() + `
				</div>

				<button type="submit" class="btn-submit">S'inscrire</button>
			</form>
			<p class="auth-link">
				Vous avez déjà un compte ? <a href="/login.html">Connectez-vous</a>
			</p>
		</div>
	</main>

	<footer class="site-footer">
		<div class="container">© 2025 Ydays</div>
	</footer>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "templates/login.html")
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		errors := make(map[string]string)

		if email == "" {
			errors["email"] = "Email requis"
		}
		if password == "" {
			errors["password"] = "Mot de passe requis"
		}

		user, err := GetUserByEmail(email)
		if err != nil || user == nil || !VerifyPassword(user.PasswordHash, password) {
			errors["general"] = "Email ou mot de passe incorrect"
		}

		if len(errors) > 0 {
			loginFormWithErrors(w, email, errors)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "user_email",
			Value: email,
			Path:  "/",
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func loginFormWithErrors(w http.ResponseWriter, email string, errors map[string]string) {
	html := `<!DOCTYPE html>
<html lang="fr">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Connexion - MODULSPACE</title>
	<link rel="stylesheet" href="/static/css/style.css?v=31">
	<style>
		.auth-container {
			max-width: 450px;
			margin: 3rem auto;
			padding: 2.5rem;
			background: white;
			border-radius: 12px;
			box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
		}
		.auth-container h2 {
			text-align: center;
			color: #6161AB;
			margin-bottom: 2rem;
			font-size: 1.8rem;
		}
		.form-error {
			background: #ffebee;
			border-left: 4px solid #ef5350;
			color: #c62828;
			padding: 1rem;
			border-radius: 4px;
			margin-bottom: 1.5rem;
			font-weight: 500;
			display: flex;
			align-items: center;
			gap: 0.5rem;
		}
		.form-error::before {
			content: '⚠';
			font-size: 1.2rem;
		}
		.form-group {
			margin-bottom: 1.5rem;
		}
		.form-group label {
			display: block;
			margin-bottom: 0.5rem;
			font-weight: 600;
			color: #333;
		}
		.form-group input {
			width: 100%;
			padding: 0.75rem;
			border: 2px solid #e0e0e0;
			border-radius: 6px;
			font-size: 1rem;
			box-sizing: border-box;
			transition: border-color 0.3s;
		}
		.form-group input:focus {
			outline: none;
			border-color: #6161AB;
		}
		.form-group input.error {
			border-color: #ef5350;
		}
		.field-error {
			display: block;
			color: #ef5350;
			font-size: 0.85rem;
			margin-top: 0.4rem;
			font-weight: 500;
		}
		.btn-submit {
			width: 100%;
			padding: 0.9rem;
			background: #6161AB;
			color: white;
			border: none;
			border-radius: 6px;
			font-size: 1rem;
			font-weight: 600;
			cursor: pointer;
			transition: all 0.3s;
		}
		.btn-submit:hover {
			background: #4d4d8f;
			transform: translateY(-2px);
			box-shadow: 0 4px 12px rgba(97, 97, 171, 0.3);
		}
		.auth-link {
			text-align: center;
			margin-top: 1.5rem;
			color: #666;
		}
		.auth-link a {
			color: #6161AB;
			text-decoration: none;
			font-weight: 600;
		}
		.auth-link a:hover {
			text-decoration: underline;
		}
	</style>
</head>
<body>
	<header class="site-header">
		<div class="container">
			<span class="header-welcome">BIENVENUE</span>
			<img src="/static/img/logo.png" alt="Logo" class="site-logo">
		</div>
	</header>

	<div class="sub-banner">
		<div class="container">
			<nav>
				<ul>
					<li><a href="/">Accueil</a></li>
					<li><a href="/apropos.html">A Propos</a></li>
					<li><a href="/produit.html">Produit</a></li>
					<li><a href="/contact.html">Contact</a></li>
				</ul>
			</nav>
		</div>
	</div>

	<main class="container">
		<div class="auth-container">
			<h2>Connexion</h2>`

	if errors["general"] != "" {
		html += `<div class="form-error">` + errors["general"] + `</div>`
	}

	html += `<form method="POST">
				<div class="form-group">
					<label for="email">Email</label>
					<input type="email" id="email" name="email" value="` + email + `" class="` + 
					(func() string {
						if errors["email"] != "" {
							return "error"
						}
						return ""
					})() + `">` + 
					(func() string {
						if errors["email"] != "" {
							return `<span class="field-error">` + errors["email"] + `</span>`
						}
						return ""
					})() + `
				</div>

				<div class="form-group">
					<label for="password">Mot de passe</label>
					<input type="password" id="password" name="password" class="` + 
					(func() string {
						if errors["password"] != "" {
							return "error"
						}
						return ""
					})() + `">` + 
					(func() string {
						if errors["password"] != "" {
							return `<span class="field-error">` + errors["password"] + `</span>`
						}
						return ""
					})() + `
				</div>

				<button type="submit" class="btn-submit">Se connecter</button>
			</form>
			<p class="auth-link">
				Pas encore inscrit ? <a href="/register.html">Créez un compte</a>
			</p>
		</div>
	</main>

	<footer class="site-footer">
		<div class="container">© 2025 Ydays</div>
	</footer>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier si l'utilisateur est connecté
	user := GetUserFromSession(r)
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"message": "Vous devez être connecté pour demander un devis",
		})
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

func userHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	w.Header().Set("Content-Type", "application/json")
	if user == nil {
		w.Write([]byte(`{"loggedIn":false}`))
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"loggedIn": true,
		"email":    user.Email,
		"prenom":   user.Prenom,
		"nom":      user.Nom,
	})
}
