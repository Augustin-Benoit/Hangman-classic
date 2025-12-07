package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Game struct {
	Word      string   `json:"-"`       // pas envoyé au client
	Mask      string   `json:"mask"`    // mot masqué
	Lives     int      `json:"lives"`   // vies restantes
	Wrong     []string `json:"wrong"`   // lettres fausses
	Correct   []string `json:"correct"` // lettres correctes
	Completed bool     `json:"completed"`
}

var (
	mu    sync.Mutex
	games = map[string]*Game{} // sid -> game
	words = []string{
		"MAISON", "VOITURE", "JARDIN", "TABLE", "CHAISE", "FENETRE", "PORTE", "CUISINE", "SALON",
		"LIVRE", "STYLO", "ORDINATEUR", "TELEPHONE", "BOUTEILLE", "VERRE", "ASSIETTE", "CUILLERE",
		"FOURCHETTE", "PAIN", "FROMAGE", "FRUIT", "LEGUME", "POMME", "BANANE", "ORANGE", "TOMATE",
		"CAROTTE", "CHOCOLAT", "CAFE", "THE", "EAU", "SOLEIL", "NUAGE", "PLUIE", "VENT", "NEIGE",
		"MER", "MONTAGNE", "FORET", "ARBRE", "FLEUR", "OISEAU", "CHAT", "CHIEN", "POISSON", "CHEVAL",
		"VOYAGE", "TRAIN", "AVION", "BUS", "ROUTE", "VILLE", "RUE", "ECOLE", "PROFESSEUR", "ELEVE",
		"AMITIE", "BONHEUR", "JOIE", "TRAVAIL", "SPORT", "MUSIQUE", "DANSE", "FILM", "PHOTO", "JEU",
	}
)

func sid(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("sid"); err == nil && c.Value != "" {
		return c.Value
	}
	const cs = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 16)
	for i := range b {
		b[i] = cs[rand.Intn(len(cs))]
	}
	s := string(b)
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: s, Path: "/", HttpOnly: true})
	return s
}

func newGame() *Game {
	w := strings.ToUpper(words[rand.Intn(len(words))])
	return &Game{Word: w, Mask: strings.Repeat("_", len(w)), Lives: 6}
}

func recomputeMask(word string, correct []string) string {
	set := map[rune]bool{}
	for _, l := range correct {
		if len(l) == 1 {
			set[rune(l[0])] = true
		}
	}
	var b strings.Builder
	for _, r := range word {
		if set[r] {
			b.WriteRune(r)
		} else {
			b.WriteString("_")
		}
	}
	return b.String()
}

// Register expose les routes JSON /new et /guess
func Register() {
	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s := sid(w, r)
		mu.Lock()
		g := newGame()
		games[s] = g
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g)
	})

	http.HandleFunc("/guess", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s := sid(w, r)

		mu.Lock()
		g := games[s]
		if g == nil {
			g = newGame()
			games[s] = g
		}
		mu.Unlock()

		letter := strings.ToUpper(strings.TrimSpace(r.FormValue("letter")))
		if len(letter) != 1 || letter[0] < 'A' || letter[0] > 'Z' {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(g)
			return
		}

		// déjà essayé ?
		for _, a := range append(g.Correct, g.Wrong...) {
			if a == letter {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(g)
				return
			}
		}

		// présent dans le mot ?
		if strings.Contains(g.Word, letter) {
			g.Correct = append(g.Correct, letter)
			g.Mask = recomputeMask(g.Word, g.Correct)
			if g.Mask == g.Word {
				g.Completed = true
			}
		} else {
			g.Wrong = append(g.Wrong, letter)
			g.Lives--
			if g.Lives <= 0 {
				g.Completed = true
				g.Mask = g.Word
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g)
	})
}
