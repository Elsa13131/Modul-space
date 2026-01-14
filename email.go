package main

import (
	"fmt"
	"net/smtp"
	"os"
)

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
