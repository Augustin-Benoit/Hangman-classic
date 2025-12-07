package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func mustTpl(root, rel string, funcs template.FuncMap) *template.Template {
	p := filepath.Join(root, rel)
	t, err := template.New(filepath.Base(p)).Funcs(funcs).Option("missingkey=error").ParseFiles(p)
	if err != nil {
		log.Fatalf("Parse template %s: %v", p, err)
	}
	return t
}

func main() {
	root, _ := os.Getwd()
	log.Printf("Working dir: %s", root)

	funcs := template.FuncMap{}
	homeTpl := mustTpl(root, "templates/home/index.html", funcs)
	gameTpl := mustTpl(root, "templates/game/index.html", funcs)

	http.Handle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(http.Dir(filepath.Join(root, "assets"))),
	))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := homeTpl.Execute(w, struct{ Title string }{"Accueil — Hangman Classic"}); err != nil {
			log.Printf("homeTpl error: %v", err)
			http.Error(w, "render error", 500)
		}
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := struct {
			Title, Difficulty string
			Lives             int
		}{"Jeu — Hangman Classic", "Normal", 6}
		if err := gameTpl.Execute(w, data); err != nil {
			log.Printf("gameTpl error: %v", err)
			http.Error(w, "render error", 500)
		}
	})

	log.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
