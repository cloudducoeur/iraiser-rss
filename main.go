package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

var (
	yearlyGoals = map[string]int{}
	yearlyAdd   = map[string]int{}
	collected   = 0
	donations   = 0
	percent     = 0.0
	lastUpdated time.Time
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime)

	// Charger la config pour 2025
	if val := os.Getenv("IRAISER_GOAL_2025"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			yearlyGoals["2025"] = parsed
		}
	} else {
		yearlyGoals["2025"] = 100000 // valeur par défaut
	}

	if val := os.Getenv("IRAISER_ADD_2025"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			yearlyAdd["2025"] = parsed
		}
	}
}

func fetchData() {
	log.Println("[FETCH] Querying iRaiser API...")

	resp, err := http.Get("https://services.iraiser.eu/counter-api/restosducoeur")
	if err != nil {
		log.Println("[ERROR] Failed to fetch data:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// convertir en JSON valide
	raw := regexp.MustCompile(`^var iraiser_counter = `).ReplaceAllString(string(body), "")
	raw = regexp.MustCompile(`(\w+):`).ReplaceAllString(raw, `"$1":`)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		log.Println("[ERROR] Failed to parse JSON:", err)
		return
	}

	// On ne garde que 2025
	valDon, okDon := data["RE2025_nb"].(float64)
	valCol, okCol := data["RE2025_value"].(float64)

	if okDon {
		donations = int(valDon)
	}
	if okCol {
		additional := yearlyAdd["2025"]
		collected = int(valCol) + additional
		goal := yearlyGoals["2025"]
		percent = float64(collected) / float64(goal) * 100
	}
	lastUpdated = time.Now()
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	item := Item{
		Title: "iRaiser 2025 – " + strconv.Itoa(collected) + "€ collectés",
		Description: strconv.Itoa(collected) + "€ collectés, " +
			strconv.Itoa(donations) + " dons (" + strconv.FormatFloat(percent, 'f', 2, 64) + "% de l’objectif)",
		PubDate: lastUpdated.Format(time.RFC1123Z),
		GUID:    "iraiser-2025-" + lastUpdated.Format("20060102150405"),
	}

	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:       "iRaiser Collecte 2025",
			Link:        "https://services.iraiser.eu/counter-api/restosducoeur",
			Description: "Flux RSS iRaiser (2025 uniquement)",
			Items:       []Item{item},
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	xml.NewEncoder(w).Encode(rss)
}

func main() {
	go func() {
		for {
			fetchData()
			time.Sleep(60 * time.Second) // maj chaque minute
		}
	}()

	http.HandleFunc("/rss", rssHandler)
	log.Println("[INFO] iRaiser RSS feed available on http://localhost:9191/rss")
	log.Fatal(http.ListenAndServe(":9191", nil))
}
