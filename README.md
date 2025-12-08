Un petit jeu du Pendu développé en Go, avec une interface web simple (HTML/CSS) et un backend minimaliste qui expose deux routes JSON. Idéal pour apprendre les bases du serveur HTTP en Go, la gestion de templates, et une logique de jeu côté serveur + mises à jour côté client.

Fonctionnalités

Jeu du pendu classique avec PV (lives), mot masqué, lettres correctes / incorrectes.
Clavier virtuel AZERTY pour jouer à la souris.
Dessin du pendu en SVG dans le template HTML, mis à jour dynamiquement.
Backend Go avec routes JSON : POST /new (nouvelle partie) et POST /guess (proposer une lettre).
Templates intégrés (templates/index/index.html).
Assets statiques servis par le FileServer (assets/style/style.css).

Prérequis

Go 1.20+ (ou version compatible avec net/http et html/template).
Un navigateur récent (Chrome/Firefox/Edge/Safari).
Aucun certificat requis (pas de HTTPS/mkcert) — le serveur tourne en HTTP local.

Lancer l'application
Dans le répertoire du projet (où se trouve main.go) :
Télécharger les dépendances (si vous utilisez des modules)

go mod tidy

Démarrer le serveur

go run main.go

Merci de jouer à notre jeu !

Crée par Théo Delourneau et Augustin Benoît.