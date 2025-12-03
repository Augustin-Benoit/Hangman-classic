package main

import (
    "bufio"
    "fmt"
    "math/rand"
    "os"
    "strings"
    "time"
)

const TENTATIVES = 10 // nombre total de tentatives

// ------------------------------------------------------------
// Chargement des mots du fichier mots.txt
// ------------------------------------------------------------
func chargerMots(fichier string) []string {
    data, err := os.ReadFile(fichier)
    if err != nil {
        fmt.Println("Erreur : impossible d’ouvrir le fichier de mots :", fichier)
        os.Exit(1)
    }

    lignes := strings.Split(string(data), "\n")
    mots := []string{}

    for _, ligne := range lignes {
        mot := strings.TrimSpace(ligne)
        if mot != "" {
            mots = append(mots, strings.ToLower(mot))
        }
    }
    return mots
}

// ------------------------------------------------------------
// Chargement des 10 positions du pendu (José) dans hangman.txt
// Chaque position fait 7 lignes
// ------------------------------------------------------------
func chargerPositions(fichier string) []string {
    file, err := os.Open(fichier)
    if err != nil {
        fmt.Println("Erreur : impossible de lire hangman.txt :", err)
        os.Exit(1)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    positions := []string{}
    var bloc strings.Builder
    ligneCount := 0

    for scanner.Scan() {
        bloc.WriteString(scanner.Text() + "\n")
        ligneCount++

        if ligneCount == 7 {
            positions = append(positions, bloc.String())
            bloc.Reset()
            ligneCount = 0
        }
    }

    if len(positions) != TENTATIVES {
        fmt.Printf("Erreur : hangman.txt doit contenir exactement %d positions de 7 lignes.\n", TENTATIVES)
        os.Exit(1)
    }

    return positions
}

// ------------------------------------------------------------
// Affiche le mot avec les lettres trouvées
// ------------------------------------------------------------
func afficherMot(mot string, trouvees map[rune]bool) string {
    var res strings.Builder
    for _, c := range mot {
        if trouvees[c] {
            res.WriteRune(c)
        } else {
            res.WriteRune('_')
        }
        res.WriteRune(' ')
    }
    return res.String()
}

// ------------------------------------------------------------
// Fonction principale appelée par main.go
// ------------------------------------------------------------
func Run() {
    if len(os.Args) != 3 {
        fmt.Println("Utilisation : go run main.go mots.txt hangman.txt")
        return
    }

    mots := chargerMots(os.Args[1])
    positions := chargerPositions(os.Args[2])

    rand.Seed(time.Now().UnixNano())
    mot := mots[rand.Intn(len(mots))]

    fmt.Println("🎮 Bienvenue dans le jeu du Pendu !")
    fmt.Printf("Le mot contient %d lettres.\n", len([]rune(mot)))

    lettresTrouvees := map[rune]bool{}

    // ------------------------------------------------------------
    // PART 1 — révélation de n lettres au début
    // n = len(mot)/2 - 1
    // ------------------------------------------------------------
    n := len([]rune(mot))/2 - 1
    if n < 0 {
        n = 0
    }

    indices := rand.Perm(len([]rune(mot)))[:n]
    for _, i := range indices {
        lettresTrouvees[rune(mot[i])] = true
    }

    fmt.Printf("%d lettres ont été révélées au hasard.\n", n)

    // ------------------------------------------------------------
    // Boucle principale
    // ------------------------------------------------------------
    tentatives := TENTATIVES
    reader := bufio.NewReader(os.Stdin)

    for tentatives > 0 {
        fmt.Println("\n----------------------------------------")
        fmt.Println("Position de José (tentatives restantes :", tentatives, ")")
        fmt.Println(positions[TENTATIVES-tentatives]) // position selon le nombre de tentatives déjà perdues

        fmt.Println("Mot :", afficherMot(mot, lettresTrouvees))

        fmt.Print("Propose une lettre : ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(strings.ToLower(input))

        if len(input) != 1 || input < "a" || input > "z" {
            fmt.Println("❌ Erreur : entre une seule lettre valide.")
            continue
        }

        lettre := rune(input[0])

        if lettresTrouvees[lettre] {
            fmt.Println("Lettre déjà trouvée.")
            continue
        }

        if strings.ContainsRune(mot, lettre) {
            fmt.Println("✔️ Bonne lettre !")
            lettresTrouvees[lettre] = true
        } else {
            fmt.Println("❌ Mauvaise lettre.")
            tentatives--
        }

        // Vérification du mot complet
        complet := true
        for _, c := range mot {
            if !lettresTrouvees[c] {
                complet = false
                break
            }
        }

        if complet {
            fmt.Println("\n🎉 BRAVO ! Vous avez trouvé le mot :", mot)
            return
        }
    }

    fmt.Println("\n💀 Vous avez perdu !")
    fmt.Println("Le mot était :", mot)
    fmt.Println("José finit pendu...")
}
