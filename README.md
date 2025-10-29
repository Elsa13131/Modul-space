# Base de site — Modul-space

Ceci est une base minimale pour un site statique (strictement HTML + CSS).

Structure actuelle :

- `index.html` — point d'entrée (HTML)
- `static/css/style.css` — styles de base (CSS)
- `main.go` — petit serveur Go qui ne sert que HTML et CSS

Comment lancer localement (Go)

1) Depuis la racine du projet, lancer le serveur :

```powershell
go run main.go
```

2) Ouvrir http://localhost:8080 dans le navigateur.

Comportement du serveur Go

- Sert `index.html` sur `/` ou `/index.html`.
- Sert uniquement les fichiers `*.css` situés dans `static/css/` via l'URL `/static/css/<name>.css`.
- Toute autre requête (JS, images, fichiers binaires, accès hors dossier) renvoie `403` ou `404`.

Si tu veux que je génère automatiquement une page supplémentaire ou organise les pages (ex : `pages/`), je peux le faire, mais j'ai laissé la structure minimale telle que demandée (HTML + CSS seulement).