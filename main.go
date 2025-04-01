package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Person represents a historical figure
type Person struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Era        string   `json:"era"`
	Profession string   `json:"profession"`
	ImageURL   string   `json:"imageUrl,omitempty"`
	YearBirth  int      `json:"yearBirth"`
	YearDeath  int      `json:"yearDeath,omitempty"`
	Country    string   `json:"country"`
	Info       string   `json:"info,omitempty"`
	Group      int      `json:"group"` // For visualization grouping
}

// Connection represents a relationship between two historical figures
type Connection struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	Type        string `json:"type"`        // e.g., "mentor", "colleague", "rival", "influenced"
	Strength    int    `json:"strength"`    // 1-10 scale
	Description string `json:"description"`
}

// GraphData represents the complete network data
type GraphData struct {
	Nodes []Person     `json:"nodes"`
	Links []Connection `json:"links"`
}

var (
	graphData GraphData
	mu        sync.RWMutex
	wikiService *WikipediaService
)

func main() {
	// Initialize with sample data
	initSampleData()
	
	// Initialize Wikipedia service
	wikiService = NewWikipediaService()

	r := mux.NewRouter()

	// Original API endpoints
	r.HandleFunc("/api/graph", getGraphData).Methods("GET")
	r.HandleFunc("/api/people", getPeople).Methods("GET")
	r.HandleFunc("/api/people/{id}", getPersonDetails).Methods("GET")
	r.HandleFunc("/api/connections", getConnections).Methods("GET")
	r.HandleFunc("/api/people", addPerson).Methods("POST")
	r.HandleFunc("/api/connections", addConnection).Methods("POST")

	// Wikipedia API endpoints
	r.HandleFunc("/api/wikipedia/search", wikiService.SearchWikipedia).Methods("GET")
	r.HandleFunc("/api/wikipedia/scrape", wikiService.ScrapeHistoricalFigure).Methods("POST")
	r.HandleFunc("/api/wikipedia/relationships/{id}", wikiService.FindRelationships).Methods("GET")
	r.HandleFunc("/api/wikipedia/batch-scrape", wikiService.BatchScrape).Methods("POST")
	r.HandleFunc("/api/wikipedia/extract-entities", wikiService.ExtractEntitiesFromText).Methods("POST")
	r.HandleFunc("/api/wikipedia/analyze-relationship", wikiService.AnalyzeTextRelationships).Methods("POST")

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	// Start server
	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func getGraphData(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphData)
}

func getPeople(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphData.Nodes)
}

func getPersonDetails(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	
	vars := mux.Vars(r)
	id := vars["id"]

	for _, person := range graphData.Nodes {
		if person.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(person)
			return
		}
	}

	http.NotFound(w, r)
}

func getConnections(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphData.Links)
}

func addPerson(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graphData.Nodes = append(graphData.Nodes, person)
	w.WriteHeader(http.StatusCreated)
}

func addConnection(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var connection Connection
	if err := json.NewDecoder(r.Body).Decode(&connection); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate that both source and target exist
	sourceExists, targetExists := false, false
	for _, person := range graphData.Nodes {
		if person.ID == connection.Source {
			sourceExists = true
		}
		if person.ID == connection.Target {
			targetExists = true
		}
	}

	if !sourceExists || !targetExists {
		http.Error(w, "Source or target person does not exist", http.StatusBadRequest)
		return
	}

	graphData.Links = append(graphData.Links, connection)
	w.WriteHeader(http.StatusCreated)
}

func initSampleData() {
	// Sample historical figures
	graphData.Nodes = []Person{
		{ID: "socrates", Name: "Socrates", Era: "Ancient", Profession: "Philosopher", YearBirth: -470, YearDeath: -399, Country: "Greece", Group: 1, Info: "Classical Greek philosopher credited as the founder of Western philosophy"},
		{ID: "plato", Name: "Plato", Era: "Ancient", Profession: "Philosopher", YearBirth: -428, YearDeath: -348, Country: "Greece", Group: 1, Info: "Student of Socrates and teacher of Aristotle"},
		{ID: "aristotle", Name: "Aristotle", Era: "Ancient", Profession: "Philosopher", YearBirth: -384, YearDeath: -322, Country: "Greece", Group: 1, Info: "Student of Plato and founder of the Lyceum"},
		{ID: "alexander", Name: "Alexander the Great", Era: "Ancient", Profession: "Military Leader", YearBirth: -356, YearDeath: -323, Country: "Macedonia", Group: 2, Info: "Student of Aristotle who created one of the largest empires of the ancient world"},
		{ID: "newton", Name: "Isaac Newton", Era: "Modern", Profession: "Physicist", YearBirth: 1643, YearDeath: 1727, Country: "England", Group: 3, Info: "Mathematician, physicist, and key figure in the scientific revolution"},
		{ID: "einstein", Name: "Albert Einstein", Era: "Modern", Profession: "Physicist", YearBirth: 1879, YearDeath: 1955, Country: "Germany/USA", Group: 3, Info: "Developed the theory of relativity"},
		{ID: "darwin", Name: "Charles Darwin", Era: "Modern", Profession: "Naturalist", YearBirth: 1809, YearDeath: 1882, Country: "England", Group: 4, Info: "Known for his contributions to evolutionary theory"},
		{ID: "davinci", Name: "Leonardo da Vinci", Era: "Renaissance", Profession: "Polymath", YearBirth: 1452, YearDeath: 1519, Country: "Italy", Group: 5, Info: "Renaissance polymath: painter, sculptor, architect, scientist, and engineer"},
	}
}