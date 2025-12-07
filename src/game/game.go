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
	mu   sync.Mutex
	game *GameState
	tpl  *template.Template
)

// helper utilisé par le template
func sub(a, b int) int { return a - b }

// ---------- Template (Option 3) ----------
func mustInitTemplate() {
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

func HandleNewGame(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	diff := r.Form.Get("hdifficulty")
	mu.Lock()
	game = newGame(diff)
	mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleGuess(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	ch := r.Form.Get("ch")
	if ch == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	letter := []rune(strings.ToUpper(ch))[0]

	mu.Lock()
	defer mu.Unlock()

	if game == nil {
		game = newGame("normal")
	}
	if game.Over {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// ignore si déjà utilisée
	for _, used := range game.UsedLetters {
		if used == letter {
			http.Redirect(w, r, "/", http.StatusSeeOther)
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

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
