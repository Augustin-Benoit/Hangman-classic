package game

import (
	"html/template"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Game struct {
	Word       string        // mot en clair (UTF-8)
	Discovered []bool        // lettres révélées
	Used       map[rune]bool // lettres déjà jouées
	Errors     int
	MaxErrors  int
	Over       bool
	Message    string
	mu         sync.Mutex
}

var (
	tpl      = template.Must(template.ParseFiles("templates/index/index.html"))
	current  *Game
	wordlist = []string{
		"ordinateur", "developpeur", "algorithme", "interface", "reseau",
		"asynchrone", "evenement", "concurrence", "validation",
		"persistance", "arborescence", "compilation", "optimisation",
		"modulaire", "debogage", "acceleration", "gestionnaire", "navigateurs", "performances",
	}
)

func newGame(word string, maxErrors int) *Game {
	g := &Game{
		Word:       word,
		MaxErrors:  maxErrors,
		Used:       make(map[rune]bool),
		Discovered: make([]bool, len([]rune(word))),
	}
	// Révéler espaces/tirets
	for i, r := range []rune(word) {
		g.Discovered[i] = (r == ' ' || r == '-')
	}
	return g
}

func normalizeRune(r rune) rune {
	s := strings.ToLower(string(r))
	repl := strings.NewReplacer(
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"à", "a", "â", "a",
		"û", "u", "ù", "u",
		"ô", "o",
		"î", "i", "ï", "i",
		"ç", "c",
	)
	return []rune(repl.Replace(s))[0]
}

func (g *Game) Try(ch rune) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.Over {
		return
	}

	ch = normalizeRune(ch)
	if ch < 'a' || ch > 'z' || g.Used[ch] {
		return
	}
	g.Used[ch] = true

	hit := false
	wordRunes := []rune(g.Word)
	for i, r := range wordRunes {
		if normalizeRune(r) == ch {
			g.Discovered[i] = true
			hit = true
		}
	}
	if !hit {
		g.Errors++
	}

	// Fin de partie ?
	all := true
	for _, ok := range g.Discovered {
		if !ok {
			all = false
			break
		}
	}
	if all {
		g.Over = true
		g.Message = "Bravo !"
	} else if g.Errors >= g.MaxErrors {
		g.Over = true
		g.Message = "Dommage… Le mot était « " + g.Word + " »."
	}
}

// ---------- Handlers ----------

func HandleNewGame(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	word := wordlist[rand.Intn(len(wordlist))]
	// Difficulté optionnelle via query: ?hdifficulty=easy|normal|hard
	maxErr := 6
	switch r.URL.Query().Get("hdifficulty") {
	case "easy":
		maxErr = 8
	case "hard":
		maxErr = 5
	}
	current = newGame(word, maxErr)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleGuess(w http.ResponseWriter, r *http.Request) {
	if current == nil {
		http.Redirect(w, r, "/new", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	_ = r.ParseForm()
	letter := r.FormValue("ch")
	if letter == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	runes := []rune(letter)
	if len(runes) > 0 {
		current.Try(runes[0])
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	if current == nil {
		// première visite → créer une partie
		rand.Seed(time.Now().UnixNano())
		current = newGame(wordlist[rand.Intn(len(wordlist))], 6)
	}
	// Préparer le modèle
	type view struct {
		WordLen     int
		Slots       []string
		UsedLetters string
		Errors      int
		MaxErrors   int
		Over        bool
		Message     string
		Alphabet    []string
	}
	wordRunes := []rune(current.Word)
	slots := make([]string, len(wordRunes))
	for i := range wordRunes {
		if current.Discovered[i] {
			slots[i] = strings.ToUpper(string(wordRunes[i]))
		} else {
			if wordRunes[i] == ' ' || wordRunes[i] == '-' {
				slots[i] = string(wordRunes[i])
			} else {
				slots[i] = ""
			}
		}
	}
	// lettres utilisées
	used := make([]string, 0)
	for r := 'a'; r <= 'z'; r++ {
		if current.Used[r] {
			used = append(used, string(r))
		}
	}
	// alphabet (boutons)
	alpha := make([]string, 26)
	for i := 0; i < 26; i++ {
		alpha[i] = string(rune('a' + i))
	}

	data := view{
		WordLen:     len(wordRunes),
		Slots:       slots,
		UsedLetters: strings.Join(used, " "),
		Errors:      current.Errors,
		MaxErrors:   current.MaxErrors,
		Over:        current.Over,
		Message:     current.Message,
		Alphabet:    alpha,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.Execute(w, data)
}
