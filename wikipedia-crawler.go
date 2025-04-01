package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type WikipediaScraper struct {
	client     *http.Client
	knownNames map[string]bool
	mu         sync.RWMutex
}

func NewWikipediaScraper() *WikipediaScraper {
	return &WikipediaScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		knownNames: make(map[string]bool),
	}
}

// ScrapeHistoricalFigure scrapes Wikipedia for info about a historical figure
func (ws *WikipediaScraper) ScrapeHistoricalFigure(name string) (*Person, error) {
	// Format name for URL
	formattedName := strings.ReplaceAll(name, " ", "_")
	url := fmt.Sprintf("https://en.wikipedia.org/wiki/%s", formattedName)

	// Make request to Wikipedia
	resp, err := ws.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request to Wikipedia: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse HTML using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	// Extract basic information
	person := &Person{
		ID:   createIDFromName(name),
		Name: name,
	}

	// Extract birth and death years from infobox
	ws.extractLifespan(doc, person)

	// Extract profession and era
	ws.extractProfessionAndEra(doc, person)

	// Extract country/nationality
	ws.extractCountry(doc, person)

	// Extract biographical information
	ws.extractBio(doc, person)

	// Set a default group based on era/profession (can be refined later)
	person.Group = determineGroup(person.Era, person.Profession)

	// Add this person to known names
	ws.mu.Lock()
	ws.knownNames[strings.ToLower(name)] = true
	ws.mu.Unlock()

	return person, nil
}

// FindRelationships analyzes a Wikipedia page to find relationships with other historical figures
func (ws *WikipediaScraper) FindRelationships(personID string) ([]Connection, error) {
	// Get the person's name from ID
	formattedName := strings.ReplaceAll(personID, "-", " ")
	url := fmt.Sprintf("https://en.wikipedia.org/wiki/%s", strings.ReplaceAll(formattedName, " ", "_"))

	// Make request to Wikipedia
	resp, err := ws.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request to Wikipedia: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse HTML using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	// Extract content paragraphs
	content := ws.extractContent(doc)

	// Find relationships
	return ws.analyzeRelationships(personID, content)
}

// Analyze text to find relationships with other known historical figures
func (ws *WikipediaScraper) analyzeRelationships(sourceID string, content string) ([]Connection, error) {
	var connections []Connection

	// Get all known people for checking
	ws.mu.RLock()
	knownNames := make(map[string]bool)
	for name := range ws.knownNames {
		knownNames[name] = true
	}
	ws.mu.RUnlock()

	// Keywords indicating relationships
	relationshipPatterns := map[string][]string{
		"mentor":     {"mentor", "teacher", "taught", "tutored", "educated", "guided"},
		"student":    {"student", "pupil", "studied under", "learned from", "disciple"},
		"colleague":  {"colleague", "associate", "worked with", "collaborated", "partnered"},
		"influenced": {"influenced", "inspired", "impact on", "affected the thinking", "shaped the views"},
		"rival":      {"rival", "opponent", "adversary", "competed", "disagreed", "disputed", "contested"},
		"friend":     {"friend", "companion", "close to", "confidant"},
		"admired":    {"admired", "respected", "honored", "looked up to", "esteemed"},
	}

	// Check content for each known person
	for name := range knownNames {
		targetID := createIDFromName(name)

		// Skip self-relationships
		if targetID == sourceID {
			continue
		}

		// Look for the name in content
		if strings.Contains(strings.ToLower(content), name) {
			// Find relationship type by analyzing surrounding text
			relationType, strength, description := ws.determineRelationship(content, name, relationshipPatterns)

			if relationType != "" {
				connection := Connection{
					Source:      sourceID,
					Target:      targetID,
					Type:        relationType,
					Strength:    strength,
					Description: description,
				}
				connections = append(connections, connection)
			}
		}
	}

	return connections, nil
}

// determineRelationship analyzes text to determine relationship type, strength, and description
func (ws *WikipediaScraper) determineRelationship(content, targetName string, patterns map[string][]string) (string, int, string) {
	//lowerContent := strings.ToLower(content)
	lowerName := strings.ToLower(targetName)

	// Find paragraphs that mention the target person
	paragraphs := strings.Split(content, "\n")
	var relevantParagraphs []string
	for _, para := range paragraphs {
		if strings.Contains(strings.ToLower(para), lowerName) {
			relevantParagraphs = append(relevantParagraphs, para)
		}
	}

	if len(relevantParagraphs) == 0 {
		return "", 0, ""
	}

	// Check for relationship keywords
	for relType, keywords := range patterns {
		for _, keyword := range keywords {
			for _, para := range relevantParagraphs {
				lowerPara := strings.ToLower(para)
				if strings.Contains(lowerPara, keyword) && strings.Contains(lowerPara, lowerName) {
					// Get a suitable description (sentence containing both keyword and name)
					description := ws.extractRelevantSentence(para, keyword, targetName)

					// Calculate strength based on how close the keyword is to the name
					// and how prominent/frequent the relationship appears
					strength := ws.calculateRelationshipStrength(relevantParagraphs, keyword, lowerName)

					return relType, strength, description
				}
			}
		}
	}

	// If no specific relationship is found but they are mentioned together,
	// consider it a general "connection"
	if len(relevantParagraphs) > 0 {
		return "associated", 3, ws.extractRelevantSentence(relevantParagraphs[0], "", targetName)
	}

	return "", 0, ""
}

// extractRelevantSentence finds the sentence that best describes the relationship
func (ws *WikipediaScraper) extractRelevantSentence(paragraph, keyword, name string) string {
	// Split paragraph into sentences
	sentences := splitIntoSentences(paragraph)

	// Look for sentences containing both keyword and name
	for _, sentence := range sentences {
		lowerSentence := strings.ToLower(sentence)
		if (keyword == "" || strings.Contains(lowerSentence, keyword)) &&
			strings.Contains(lowerSentence, strings.ToLower(name)) {
			// Clean up the sentence
			cleaned := cleanText(sentence)
			if len(cleaned) > 200 {
				cleaned = cleaned[:197] + "..."
			}
			return cleaned
		}
	}

	// If no sentence contains both, return the first sentence mentioning the name
	for _, sentence := range sentences {
		if strings.Contains(strings.ToLower(sentence), strings.ToLower(name)) {
			cleaned := cleanText(sentence)
			if len(cleaned) > 200 {
				cleaned = cleaned[:197] + "..."
			}
			return cleaned
		}
	}

	return "Connected in historical context."
}

// calculateRelationshipStrength estimates the strength of a relationship
func (ws *WikipediaScraper) calculateRelationshipStrength(paragraphs []string, keyword, name string) int {
	// Count mentions of the relationship
	mentionCount := 0
	for _, para := range paragraphs {
		lowerPara := strings.ToLower(para)
		if strings.Contains(lowerPara, keyword) && strings.Contains(lowerPara, name) {
			mentionCount++
		}
	}

	// More mentions = stronger relationship
	// Scale from 1-10
	switch {
	case mentionCount >= 5:
		return 10
	case mentionCount >= 4:
		return 8
	case mentionCount >= 3:
		return 7
	case mentionCount >= 2:
		return 5
	case mentionCount >= 1:
		return 4
	default:
		return 3 // Default for associated but relationship not explicitly described
	}
}

// Helper functions for information extraction

func (ws *WikipediaScraper) extractLifespan(doc *goquery.Document, person *Person) {
	// Look for birth and death dates in the infobox
	birthDateText := doc.Find(".infobox .bday").Text()
	deathDateText := doc.Find(".infobox .dday").Text()

	// Extract years using regex
	birthYear := extractYear(birthDateText)
	deathYear := extractYear(deathDateText)

	// If not found in specific fields, try to find in general text
	if birthYear == 0 || deathYear == 0 {
		// Check the first paragraph for birth-death pattern
		firstPara := doc.Find("#mw-content-text p").First().Text()
		years := extractYearsFromText(firstPara)

		if len(years) >= 1 && birthYear == 0 {
			birthYear = years[0]
		}
		if len(years) >= 2 && deathYear == 0 {
			deathYear = years[1]
		}
	}

	person.YearBirth = birthYear
	if deathYear != 0 {
		person.YearDeath = deathYear
	}
}

func (ws *WikipediaScraper) extractProfessionAndEra(doc *goquery.Document, person *Person) {
	// First paragraph often contains profession
	firstPara := doc.Find("#mw-content-text p").First().Text()

	// Common profession keywords
	professions := []string{"philosopher", "scientist", "physicist", "mathematician",
		"writer", "artist", "politician", "leader", "general",
		"composer", "inventor", "explorer", "king", "queen",
		"emperor", "empress", "president", "prime minister"}

	for _, profession := range professions {
		if strings.Contains(strings.ToLower(firstPara), profession) {
			person.Profession = strings.Title(profession)
			break
		}
	}

	if person.Profession == "" {
		person.Profession = "Historical Figure"
	}

	// Determine era based on birth year
	if person.YearBirth < -800 {
		person.Era = "Ancient (Pre-Classical)"
	} else if person.YearBirth < -500 {
		person.Era = "Classical Antiquity"
	} else if person.YearBirth < 476 {
		person.Era = "Ancient"
	} else if person.YearBirth < 1000 {
		person.Era = "Early Medieval"
	} else if person.YearBirth < 1300 {
		person.Era = "High Medieval"
	} else if person.YearBirth < 1500 {
		person.Era = "Late Medieval"
	} else if person.YearBirth < 1650 {
		person.Era = "Renaissance"
	} else if person.YearBirth < 1800 {
		person.Era = "Early Modern"
	} else if person.YearBirth < 1914 {
		person.Era = "Modern"
	} else {
		person.Era = "Contemporary"
	}
}

func (ws *WikipediaScraper) extractCountry(doc *goquery.Document, person *Person) {
	// Try to find nationality or country in the infobox
	infobox := doc.Find(".infobox")

	// Look for common nationality/country fields
	nationLabels := []string{"Nationality", "Country", "Born", "Citizenship"}

	for _, label := range nationLabels {
		infobox.Find("tr").Each(func(i int, s *goquery.Selection) {
			headerText := s.Find("th").Text()
			if strings.Contains(headerText, label) {
				country := s.Find("td").Text()
				// Clean up the text
				country = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(country, "")
				country = strings.TrimSpace(country)

				// If "Born" field, extract just the country
				if label == "Born" {
					// Usually last part of the field is country
					parts := strings.Split(country, ",")
					if len(parts) > 0 {
						country = strings.TrimSpace(parts[len(parts)-1])
					}
				}

				if country != "" {
					person.Country = country
					return
				}
			}
		})

		if person.Country != "" {
			break
		}
	}

	// Default if not found
	if person.Country == "" {
		person.Country = "Unknown"
	}
}

func (ws *WikipediaScraper) extractBio(doc *goquery.Document, person *Person) {
	// Get the first paragraph as a brief bio
	firstPara := doc.Find("#mw-content-text p").First().Text()
	// Clean up the text
	bio := cleanText(firstPara)

	if len(bio) > 500 {
		bio = bio[:497] + "..."
	}

	person.Info = bio
}

func (ws *WikipediaScraper) extractContent(doc *goquery.Document) string {
	var content strings.Builder

	// Extract all paragraphs from the main content
	doc.Find("#mw-content-text p").Each(func(i int, s *goquery.Selection) {
		content.WriteString(s.Text())
		content.WriteString("\n")
	})

	// Also get headings and text from sections
	doc.Find("#mw-content-text h2, #mw-content-text h3").Each(func(i int, s *goquery.Selection) {
		content.WriteString(s.Text())
		content.WriteString("\n")
	})

	return content.String()
}

// Utility functions

func createIDFromName(name string) string {
	// Convert name to lowercase
	id := strings.ToLower(name)
	// Replace spaces with hyphens
	id = strings.ReplaceAll(id, " ", "-")
	// Remove any special characters
	id = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(id, "")
	return id
}

func extractYear(dateStr string) int {
	re := regexp.MustCompile(`\b\d{4}\b`)
	matches := re.FindStringSubmatch(dateStr)
	if len(matches) > 0 {
		var year int
		fmt.Sscanf(matches[0], "%d", &year)
		return year
	}
	return 0
}

func extractYearsFromText(text string) []int {
	re := regexp.MustCompile(`\b\d{4}\b`)
	matches := re.FindAllString(text, -1)

	var years []int
	for _, match := range matches {
		var year int
		fmt.Sscanf(match, "%d", &year)
		years = append(years, year)
	}

	return years
}

func splitIntoSentences(text string) []string {
	// Basic sentence splitting - can be improved
	re := regexp.MustCompile(`[.!?]["\s)]`)
	sentences := re.Split(text, -1)

	var result []string
	for _, s := range sentences {
		if len(strings.TrimSpace(s)) > 0 {
			result = append(result, strings.TrimSpace(s))
		}
	}

	return result
}

func cleanText(text string) string {
	// Remove Wikipedia citations [1], [2], etc.
	text = regexp.MustCompile(`\[\d+\]`).ReplaceAllString(text, "")
	// Remove other brackets
	text = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(text, "")
	// Replace multiple spaces with a single space
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func determineGroup(era, profession string) int {
	// Simple group assignment based on era and profession
	// This can be expanded to be more sophisticated
	switch era {
	case "Ancient", "Classical Antiquity", "Ancient (Pre-Classical)":
		return 1
	case "Early Medieval", "High Medieval", "Late Medieval":
		return 2
	case "Renaissance":
		return 3
	case "Early Modern":
		return 4
	case "Modern":
		return 5
	case "Contemporary":
		return 6
	default:
		return 7
	}
}

// WikipediaAPI provides a way to use Wikipedia's API for more structured data
func (ws *WikipediaScraper) WikipediaAPI(query string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=query&format=json&prop=extracts&exintro=true&redirects=1&titles=%s",
		strings.ReplaceAll(query, " ", "_"))

	resp, err := ws.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// BatchScrapeHistoricalFigures scrapes information for multiple historical figures
func (ws *WikipediaScraper) BatchScrapeHistoricalFigures(names []string) ([]*Person, error) {
	var people []*Person
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, len(names))

	for _, name := range names {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			// Throttle requests to be kind to Wikipedia
			time.Sleep(1 * time.Second)

			person, err := ws.ScrapeHistoricalFigure(name)
			if err != nil {
				log.Printf("Error scraping %s: %v", name, err)
				errCh <- err
				return
			}

			mu.Lock()
			people = append(people, person)
			mu.Unlock()
		}(name)
	}

	wg.Wait()
	close(errCh)

	// Check if there were any errors
	for err := range errCh {
		if err != nil {
			return people, fmt.Errorf("some scraping operations failed: %w", err)
		}
	}

	return people, nil
}

// BatchFindRelationships finds relationships for multiple historical figures
func (ws *WikipediaScraper) BatchFindRelationships(personIDs []string) ([]Connection, error) {
	var allConnections []Connection
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, len(personIDs))

	for _, id := range personIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// Throttle requests to be kind to Wikipedia
			time.Sleep(1 * time.Second)

			connections, err := ws.FindRelationships(id)
			if err != nil {
				log.Printf("Error finding relationships for %s: %v", id, err)
				errCh <- err
				return
			}

			mu.Lock()
			allConnections = append(allConnections, connections...)
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	close(errCh)

	// Check if there were any errors
	for err := range errCh {
		if err != nil {
			return allConnections, fmt.Errorf("some relationship operations failed: %w", err)
		}
	}

	return allConnections, nil
}
