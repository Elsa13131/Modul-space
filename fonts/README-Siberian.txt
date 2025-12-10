Placez ici les fichiers de police "Siberian" si vous souhaitez utiliser cette fonte localement.

Fichiers attendus (exemples de noms) :
- Siberian-Regular.woff2
- Siberian-Regular.woff

Les fichiers doivent être copiés dans le dossier `fonts/` à la racine du projet. Le CSS dans `static/css/style.css` contient déjà la déclaration @font-face pointant vers `/fonts/Siberian-Regular.woff2` et `/fonts/Siberian-Regular.woff`.

Si vous n'avez pas ces fichiers, vous pouvez :
- obtenir la famille Siberian (format webfont) depuis vos sources de polices, ou
- remplacer ces noms par ceux de la police que vous possédez.

Après ajout, rechargez la page (Ctrl+F5) et vérifiez dans DevTools -> Network -> Font que `/fonts/Siberian-Regular.woff2` renvoie 200.