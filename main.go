package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"hangman-classic/src/api"
)

func mustTpl(root, rel string) *template.Template {
	p := filepath.Join(root, rel)
	t, err := template.ParseFiles(p)
	if err != nil {
		log.Fatalf("Parse template %s: %v", p, err)
	}
	return t
}

func main() {
	root, _ := os.Getwd()
	home := mustTpl(root, "templates/home/index.html")
	game := mustTpl(root, "templates/game/index.html")

	http.Handle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(http.Dir(filepath.Join(root, "assets"))),
	))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := home.Execute(w, struct{ Title string }{"Hangman Classic 👋"}); err != nil {
			log.Printf("homeTpl error: %v", err)
			http.Error(w, "render error", 500)
		}
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := game.Execute(w, struct{ Title string }{"Jeu — Hangman Classic"}); err != nil {
			log.Printf("gameTpl error: %v", err)
			http.Error(w, "render error", 500)
		}
	})

	api.Register()

	log.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
