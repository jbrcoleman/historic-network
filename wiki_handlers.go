package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

// WikipediaService combines the scraper and NLP analyzer
type WikipediaService struct {
	scraper  *WikipediaScraper
	analyzer *NLPAnalyzer
	inProgress map[string]bool // track ongoing scraping operations
	mu         sync.RWMutex
}

// NewWikipediaService creates a new Wikipedia service
func NewWikipediaService() *WikipediaService {
	return &WikipediaService{
		scraper:    NewWikipediaScraper(),
		analyzer:   NewNLPAnalyzer(),
		inProgress: make(map[string]bool),
	}
}

// SearchWikipedia handles searching for historical figures
func (ws *WikipediaService) SearchWikipedia(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Use the Wikipedia API to search for matches
	results, err := ws.searchWikipedia(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search Wikipedia: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// searchWikipedia performs the actual search
func (ws *WikipediaService) searchWikipedia(query string) ([]map[string]string, error) {
	// This is a simplified implementation
	// In a real-world scenario, you'd use Wikipedia's API for more accurate results
	
	url := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=opensearch&search=%s&limit=10&namespace=0&format=json",
		strings.ReplaceAll(query, " ", "%20"))
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var searchResults []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&searchResults); err != nil {
		return nil, err
	}
	
	// Extract results from the OpenSearch response
	if len(searchResults) < 4 {
		return nil, fmt.Errorf("unexpected response format")
	}
	
	titles, ok := searchResults[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid titles format")
	}
	
	descriptions, ok := searchResults[2].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid descriptions format")
	}
	
	urls, ok := searchResults[3].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid URLs format")
	}
	
	var results []map[string]string
	for i := 0; i < len(titles); i++ {
		title, _ := titles[i].(string)
		description, _ := descriptions[i].(string)
		url, _ := urls[i].(string)
		
		results = append(results, map[string]string{
			"title":       title,
			"description": description,
			"url":         url,
		})
	}
	
	return results, nil
}

// ScrapeHistoricalFigure handles scraping a single historical figure
func (ws *WikipediaService) ScrapeHistoricalFigure(w http.ResponseWriter, r *http.Request) {
	// Parse request body to get name
	var request struct {
		Name string `json:"name"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	
	if request.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	
	// Check if this figure is already being processed
	ws.mu.Lock()
	if ws.inProgress[request.Name] {
		ws.mu.Unlock()
		http.Error(w, "Already processing this historical figure", http.StatusConflict)
		return
	}
	
	// Mark as in progress
	ws.inProgress[request.Name] = true
	ws.mu.Unlock()
	
	// Ensure we mark as no longer in progress when done
	defer func() {
		ws.mu.Lock()
		delete(ws.inProgress, request.Name)
		ws.mu.Unlock()
	}()
	
	// Scrape the figure
	person, err := ws.scraper.ScrapeHistoricalFigure(request.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to scrape historical figure: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Add to graph data
	mu.Lock()
	
	// Check if person already exists
	exists := false
	for _, node := range graphData.Nodes {
		if node.ID == person.ID {
			exists = true
			break
		}
	}
	
	if !exists {
		graphData.Nodes = append(graphData.Nodes, *person)
	}
	
	mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(person)
}

// FindRelationships handles extracting relationships for a figure
func (ws *WikipediaService) FindRelationships(w http.ResponseWriter, r *http.Request) {
	// Get person ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	
	if id == "" {
		http.Error(w, "Person ID is required", http.StatusBadRequest)
		return
	}
	
	// Find relationships
	connections, err := ws.scraper.FindRelationships(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to find relationships: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Add new connections to graph data
	mu.Lock()
	
	// Check for existing connections to avoid duplicates
	for _, conn := range connections {
		exists := false
		for _, existingConn := range graphData.Links {
			if existingConn.Source == conn.Source && existingConn.Target == conn.Target {
				exists = true
				break
			}
		}
		
		if !exists {
			graphData.Links = append(graphData.Links, conn)
		}
	}
	
	mu.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

// BatchScrape handles scraping multiple historical figures
func (ws *WikipediaService) BatchScrape(w http.ResponseWriter, r *http.Request) {
	// Parse request body to get names
	var request struct {
		Names []string `json:"names"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	
	if len(request.Names) == 0 {
		http.Error(w, "At least one name is required", http.StatusBadRequest)
		return
	}
	
	// Flag in-progress names
	ws.mu.Lock()
	for _, name := range request.Names {
		ws.inProgress[name] = true
	}
	ws.mu.Unlock()
	
	// Ensure we mark as no longer in progress when done
	defer func() {
		ws.mu.Lock()
		for _, name := range request.Names {
			delete(ws.inProgress, name)
		}
		ws.mu.Unlock()
	}()
	
	// Scrape the figures
	people, err := ws.scraper.BatchScrapeHistoricalFigures(request.Names)
	if err != nil {
		http.Error(w, fmt.Sprintf("Some scraping operations failed: %v", err), http.StatusInternalServerError)
		// Continue with partial results
	}
	
	// Add to graph data
	mu.Lock()
	
	for _, person := range people {
		// Check if person already exists
		exists := false
		for _, node := range graphData.Nodes {
			if node.ID == person.ID {
				exists = true
				break
			}
		}
		
		if !exists {
			graphData.Nodes = append(graphData.Nodes, *person)
		}
	}
	
	mu.Unlock()
	
	// Extract person IDs for relationship analysis
	var personIDs []string
	for _, person := range people {
		personIDs = append(personIDs, person.ID)
	}
	
	// Find relationships
	connections, err := ws.scraper.BatchFindRelationships(personIDs)
	if err != nil {
		// Log error but continue with partial results
		fmt.Printf("Some relationship analyses failed: %v\n", err)
	}
	
	// Add new connections to graph data
	mu.Lock()
	
	// Check for existing connections to avoid duplicates
	for _, conn := range connections {
		exists := false
		for _, existingConn := range graphData.Links {
			if existingConn.Source == conn.Source && existingConn.Target == conn.Target {
				exists = true
				break
			}
		}
		
		if !exists {
			graphData.Links = append(graphData.Links, conn)
		}
	}
	
	mu.Unlock()
	
	// Create response
	response := struct {
		People      []*Person     `json:"people"`
		Connections []Connection `json:"connections"`
	}{
		People:      people,
		Connections: connections,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ExtractEntitiesFromText handles extracting named entities from text
func (ws *WikipediaService) ExtractEntitiesFromText(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var request struct {
		Text string `json:"text"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	
	if request.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}
	
	// Extract entities
	entities := ws.analyzer.ExtractNamedEntities(request.Text)
	
	response := struct {
		Entities []string `json:"entities"`
	}{
		Entities: entities,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AnalyzeTextRelationships handles analyzing text for relationship indicators
func (ws *WikipediaService) AnalyzeTextRelationships(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var request struct {
		Text   string `json:"text"`
		Source string `json:"source"`
		Target string `json:"target"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	
	if request.Text == "" || request.Source == "" || request.Target == "" {
		http.Error(w, "Text, source, and target are required", http.StatusBadRequest)
		return
	}
	
	// Analyze text
	relType, strength, description := ws.analyzer.DetermineRelationshipFromText(
		request.Text, request.Source, request.Target)
	
	response := struct {
		Type        string `json:"type"`
		Strength    int    `json:"strength"`
		Description string `json:"description"`
	}{
		Type:        relType,
		Strength:    strength,
		Description: description,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}