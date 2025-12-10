# Modul-space — déploiement

Ce dépôt contient un petit serveur Go qui sert `index.html` et le dossier `static/`.

Objectif : préparer et déployer rapidement ce site sur une plateforme qui fournit une URL publique (ex : Render, Railway).

Pré-requis
- Un compte GitHub
- Avoir poussé ce dépôt sur un repo GitHub
- Avoir Go installé localement pour tests (optionnel)

Options de déploiement recommandées

1) Render (recommandé)
 - Aller sur https://render.com et créer un "Web Service" en connectant votre repo GitHub.
 - Build Command: `go build -o main .`
 - Start Command: `./main`
 - Render détectera et lancera le service. Une URL publique du type `https://votre-service.onrender.com` sera fournie automatiquement après le déploiement.

2) Railway
 - Aller sur https://railway.app, créer un nouveau projet et connecter le repo GitHub.
 - Définir la commande de démarrage: `./main` (ou `go run main.go` pour exécuter sans construire).
 - Railway vous fournira aussi une URL publique.

3) Déploiement manuel (exécuter localement)
 - Build : `go build -o main .`
 - Lancer : `.\main.exe` (Windows) ou `./main` (Linux/macOS)
 - Le serveur écoute sur le port fourni par la variable d'environnement `PORT`. Si non fournie, il écoute sur le port `8080`.

Que faire après le déploiement ?
- L'interface de la plateforme (Render/Railway) affichera l'URL publique. Copiez-la et ouvrez-la dans votre navigateur.

Dépannage rapide
- Si la page retourne 404 pour des fichiers statiques : vérifiez que `static/` est bien dans la racine du repo et que les chemins dans `index.html` commencent par `/static/`.
- Vérifiez les logs du service sur Render/Railway pour voir les erreurs de build ou d'exécution.

Notes techniques
- L'application lit `PORT` depuis l'environnement (convention PaaS). Le serveur lie `0.0.0.0` pour être accessible publiquement.
- Un `Procfile` est inclus pour faciliter le déploiement sur des plateformes compatibles.

Si tu veux, je peux te donner les étapes précises pour connecter le repo GitHub à Render et te montrer l'URL une fois déployé — mais je ne peux pas déployer à ta place sans accès au compte.

Ajouter ton image fournie (logo) dans le projet
------------------------------------------------

Pour que l'image que tu as jointe dans la conversation apparaisse dans la barre d'info en haut à droite, fais ceci :

1. Sauvegarde l'image fournie (clic droit -> "Enregistrer l'image sous...") en lui donnant le nom exact :

	static/img/info-logo.png

	(crée le dossier `static/img` s'il n'existe pas)

2. Vérifie que `index.html` contient bien la balise suivante (déjà en place) :

	<img src="/static/img/info-logo.png" alt="Logo info" id="infoLogo">

3. Commit et pousse les changements sur GitHub :

```powershell
git add static/img/info-logo.png
git commit -m "Ajout du logo info pour la barre supérieure"
git push origin main
```

4. Si ton site est déjà déployé (Render/Railway), le déploiement automatique prendra la nouvelle image et ton site en ligne affichera le logo.

Remarque : si tu préfères que j'ajoute l'image directement dans le dépôt, copie-colle ici l'image en base64 (je peux alors la déposer au bon endroit), ou autorise l'accès à un URL public où je peux la télécharger.
