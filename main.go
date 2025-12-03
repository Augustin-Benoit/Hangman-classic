package main

import (
	"hangman-classic/src/game"
	"log"
	"net/http"
	"os"
)

func serveFile(w http.ResponseWriter, path, ctype string) {
	b, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	if ctype != "" {
		w.Header().Set("Content-Type", ctype)
	}
	w.Write(b)
}

func main() {
	// Routes jeu (rendu serveur, pas de JS)
	http.HandleFunc("/", game.HandleIndex)      // page courante
	http.HandleFunc("/new", game.HandleNewGame) // nouvelle partie
	http.HandleFunc("/guess", game.HandleGuess) // jouer une lettre

	// Assets statiques (CSS, images...)
	http.Handle("/assets/",
		http.StripPrefix("/assets/",
			http.FileServer(http.Dir("assets"))))

	log.Println("Hangman Classic sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
