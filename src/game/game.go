package game

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ---------- État du jeu ----------
type GameState struct {
	Word        string
	Errors      int
	MaxErrors   int
	UsedLetters []rune
	Slots       []string
	Alphabet    []string
	Over        bool
	Message     string
}

var (
	mu      sync.Mutex
	game    *GameState
	tpl     *template.Template // home/index (déjà utilisé par HandleIndex)
	gameTpl *template.Template // game/index (nouveau pour /game)
)

func sub(a, b int) int { return a - b } // helper pour templates

// ---------- Templates ----------
func mustInitTemplate() { // déjà utilisé par HandleIndex -> templates/index/index.html
	funcMap := template.FuncMap{"sub": sub}
	indexPath := filepath.Join("templates", "index", "index.html")

	if wd, err := os.Getwd(); err == nil {
		log.Printf("[DEBUG] CWD=%s, template=%s", wd, indexPath)
	}

	var err error
	tpl, err = template.New("index.html").Funcs(funcMap).ParseFiles(indexPath)
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}
}

// 🆕 Loader pour la page /game -> templates/game/index.html
func mustInitTemplateGame() {
	funcMap := template.FuncMap{"sub": sub}
	path := filepath.Join("templates", "game", "index.html")
	var err error
	gameTpl, err = template.New("game.html").Funcs(funcMap).ParseFiles(path)
	if err != nil {
		log.Fatalf("parse game template %s: %v", path, err)
	}
}

func buildAlphabet() []string {
	alpha := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	out := make([]string, 0, len(alpha))
	for _, r := range alpha {
		out = append(out, string(r))
	}
	return out
}

func newGame(difficulty string) *GameState {
	max := 6
	switch strings.ToLower(difficulty) {
	case "easy":
		max = 8
	case "hard":
		max = 5
	}
	word := "GOLEM"

	slots := make([]string, len(word))
	for i := range slots {
		if word[i] == ' ' || word[i] == '-' {
			slots[i] = string(word[i])
		} else {
			slots[i] = "_"
		}
	}

	return &GameState{
		Word:        word,
		Errors:      0,
		MaxErrors:   max,
		UsedLetters: []rune{},
		Slots:       slots,
		Alphabet:    buildAlphabet(),
		Over:        false,
		Message:     "",
	}
}

// ---------- Handlers exportés ----------
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	if tpl == nil {
		mustInitTemplate()
	}
	mu.Lock()
	defer mu.Unlock()

	if game == nil {
		game = newGame("normal")
	}

	data := struct {
		Errors      int
		MaxErrors   int
		UsedLetters string
		Slots       []string
		Alphabet    []string
		Over        bool
		Message     string
	}{
		Errors:      game.Errors,
		MaxErrors:   game.MaxErrors,
		UsedLetters: usedLettersString(game.UsedLetters),
		Slots:       game.Slots,
		Alphabet:    game.Alphabet,
		Over:        game.Over,
		Message:     game.Message,
	}

	if err := tpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 🆕 /game : rendu page de jeu avec pendu SVG adaptatif
func HandleGame(w http.ResponseWriter, r *http.Request) {
	if gameTpl == nil {
		mustInitTemplateGame()
	}
	mu.Lock()
	defer mu.Unlock()

	if game == nil {
		game = newGame("normal")
	}

	vm := struct {
		Title       string
		Errors      int
		MaxErrors   int
		Remaining   int
		UsedLetters string
		Slots       []string
		Alphabet    []string
		Over        bool
		Message     string
	}{
		Title:       "Jeu — Hangman Classic",
		Errors:      game.Errors,
		MaxErrors:   game.MaxErrors,
		Remaining:   sub(game.MaxErrors, game.Errors),
		UsedLetters: usedLettersString(game.UsedLetters),
		Slots:       game.Slots,
		Alphabet:    game.Alphabet,
		Over:        game.Over,
		Message:     game.Message,
	}

	if err := gameTpl.ExecuteTemplate(w, "game.html", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HandleNewGame(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	diff := r.Form.Get("hdifficulty")
	mu.Lock()
	game = newGame(diff)
	mu.Unlock()

	// Retour sur la page de jeu pour voir le pendu réinitialisé
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func HandleGuess(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	ch := r.Form.Get("ch")
	if ch == "" {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}
	letter := []rune(strings.ToUpper(ch))[0]

	mu.Lock()
	defer mu.Unlock()

	if game == nil {
		game = newGame("normal")
	}
	if game.Over {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	// ignore si déjà utilisée
	for _, used := range game.UsedLetters {
		if used == letter {
			http.Redirect(w, r, "/game", http.StatusSeeOther)
			return
		}
	}
	game.UsedLetters = append(game.UsedLetters, letter)

	// révèle les lettres
	found := false
	for i, r := range []rune(game.Word) {
		if r == letter {
			game.Slots[i] = string(r)
			found = true
		}
	}
	if !found {
		game.Errors++
		if game.Errors >= game.MaxErrors {
			game.Over = true
			game.Message = "Perdu… Le mot était: " + game.Word
		}
	} else {
		win := true
		for _, s := range game.Slots {
			if s == "_" {
				win = false
				break
			}
		}
		if win {
			game.Over = true
			game.Message = "Bravo !"
		}
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func usedLettersString(letters []rune) string {
	var b strings.Builder
	for i, r := range letters {
		if i > 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
