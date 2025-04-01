package main

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
)

// NLPAnalyzer provides natural language processing functions for historical relationship analysis
type NLPAnalyzer struct {
	// Maps to store word frequencies for different relationship types
	relationshipCorpus map[string]map[string]int
	mu                 sync.RWMutex
}

// NewNLPAnalyzer creates a new NLP analyzer with pre-trained data
func NewNLPAnalyzer() *NLPAnalyzer {
	analyzer := &NLPAnalyzer{
		relationshipCorpus: make(map[string]map[string]int),
	}
	
	// Initialize with known relationship words
	analyzer.initializeCorpus()
	
	return analyzer
}

// Initialize the corpus with known relationship indicators
func (na *NLPAnalyzer) initializeCorpus() {
	// Mentor relationship words
	na.relationshipCorpus["mentor"] = map[string]int{
		"mentor": 10, "teacher": 9, "taught": 8, "guide": 7, "instruct": 7, 
		"train": 6, "tutor": 9, "educate": 7, "master": 6, "professor": 6,
		"advise": 5, "supervise": 5, "coach": 5, "counsel": 4, "direct": 3,
	}
	
	// Student relationship words
	na.relationshipCorpus["student"] = map[string]int{
		"student": 10, "pupil": 9, "disciple": 8, "apprentice": 8, "protégé": 7,
		"follower": 6, "studied under": 9, "learned from": 8, "trainee": 6, "mentee": 7,
		"educated by": 7, "tutored by": 8, "guided by": 6, "influenced by": 5, "school of": 5,
	}
	
	// Colleague relationship words
	na.relationshipCorpus["colleague"] = map[string]int{
		"colleague": 10, "associate": 8, "collaborator": 9, "partner": 8, "coworker": 8,
		"ally": 6, "contemporary": 5, "peer": 7, "fellow": 6, "worked with": 9,
		"collaborated with": 9, "joined forces": 7, "teamed up": 7, "together": 4, "alongside": 6,
	}
	
	// Influenced relationship words
	na.relationshipCorpus["influenced"] = map[string]int{
		"influenced": 10, "inspired": 9, "affected": 7, "shaped": 8, "impacted": 8,
		"changed": 6, "transformed": 7, "informed": 6, "guided": 5, "swayed": 6,
		"impressed": 5, "sway over": 6, "impact on": 8, "effect on": 7, "inspiration for": 9,
	}
	
	// Rival relationship words
	na.relationshipCorpus["rival"] = map[string]int{
		"rival": 10, "opponent": 9, "competitor": 8, "adversary": 9, "enemy": 7,
		"foe": 7, "antagonist": 8, "critic": 6, "contested": 7, "challenged": 6,
		"disputed with": 8, "disagreed with": 7, "opposed": 8, "contended with": 7, "conflict": 6,
	}
	
	// Friend relationship words
	na.relationshipCorpus["friend"] = map[string]int{
		"friend": 10, "companion": 8, "ally": 7, "confidant": 9, "close": 6,
		"intimate": 8, "buddy": 7, "pal": 6, "associate": 5, "comrade": 7,
		"acquaintance": 4, "fellowship": 6, "friendship": 10, "friendly": 5, "amicable": 6,
	}
	
	// Admired relationship words
	na.relationshipCorpus["admired"] = map[string]int{
		"admired": 10, "respected": 8, "revered": 9, "esteemed": 8, "venerated": 9,
		"looked up to": 8, "honored": 7, "praised": 6, "acclaimed": 7, "celebrated": 6,
		"idolized": 9, "hero": 8, "model": 6, "idol": 8, "exemplar": 7,
	}
}

// AnalyzeText determines the most likely relationship types in a given text
func (na *NLPAnalyzer) AnalyzeText(text string) map[string]float64 {
	// Preprocess the text
	processedText := na.preprocessText(text)
	words := strings.Fields(processedText)
	
	// Calculate scores for each relationship type
	scores := make(map[string]float64)
	
	na.mu.RLock()
	defer na.mu.RUnlock()
	
	for relType, corpus := range na.relationshipCorpus {
		var score float64
		
		// Check for each word/phrase in the corpus
		for phrase, weight := range corpus {
			// For multi-word phrases
			if strings.Contains(phrase, " ") {
				if strings.Contains(processedText, phrase) {
					// Give higher weight to exact phrases
					score += float64(weight) * 1.5
				}
			} else {
				// For single words, count occurrences
				for _, word := range words {
					if word == phrase {
						score += float64(weight)
					}
					// Check for stemming/variations
					if strings.HasPrefix(word, phrase) && len(word) <= len(phrase)+3 {
						score += float64(weight) * 0.7
					}
				}
			}
		}
		
		// Normalize score by text length to avoid bias toward longer texts
		normalizedScore := score / math.Sqrt(float64(len(words)))
		if normalizedScore > 0 {
			scores[relType] = normalizedScore
		}
	}
	
	return scores
}

// DetermineRelationshipFromText identifies the most probable relationship type from text
func (na *NLPAnalyzer) DetermineRelationshipFromText(text, source, target string) (string, int, string) {
	scores := na.AnalyzeText(text)
	
	// Find the highest scoring relationship type
	var bestType string
	var highestScore float64
	
	for relType, score := range scores {
		if score > highestScore {
			highestScore = score
			bestType = relType
		}
	}
	
	// If no strong relationship found, check for co-occurrence
	if bestType == "" || highestScore < 1.0 {
		if strings.Contains(strings.ToLower(text), strings.ToLower(target)) {
			bestType = "associated"
			highestScore = 0.5
		}
	}
	
	// Calculate strength (1-10 scale)
	var strength int
	if highestScore > 10 {
		strength = 10
	} else if highestScore > 0 {
		strength = int(math.Max(1, math.Min(10, math.Ceil(highestScore))))
	} else {
		strength = 0
	}
	
	// Extract a relevant description
	description := na.extractRelevantDescription(text, source, target, bestType)
	
	return bestType, strength, description
}

// ExtractNamedEntities identifies potential historical figures in text
func (na *NLPAnalyzer) ExtractNamedEntities(text string) []string {
	// Basic named entity recognition for people
	// This is a simplified implementation - in production, you'd use a more robust NER solution
	
	// First, look for titles followed by capitalized words
	titlePattern := `(Mr\.|Mrs\.|Ms\.|Dr\.|Prof\.|Sir|Lord|Lady|King|Queen|Emperor|Empress|Prince|Princess|Duke|Duchess|Pope|Saint|President|Prime Minister)\s+([A-Z][a-z]+)(\s+[A-Z][a-z]+){0,3}`
	
	// Also look for capitalized words that might be names
	namePattern := `\b([A-Z][a-z]+)(\s+[A-Z][a-z]+){0,3}\b`
	
	// Find matches
	titleRegex := regexp.MustCompile(titlePattern)
	nameRegex := regexp.MustCompile(namePattern)
	
	titleMatches := titleRegex.FindAllString(text, -1)
	nameMatches := nameRegex.FindAllString(text, -1)
	
	// Combine and deduplicate matches
	allMatches := append(titleMatches, nameMatches...)
	
	// Filter out common false positives
	filteredMatches := na.filterNamedEntities(allMatches)
	
	// Deduplicate
	uniqueNames := make(map[string]bool)
	var results []string
	
	for _, name := range filteredMatches {
		if !uniqueNames[name] {
			uniqueNames[name] = true
			results = append(results, name)
		}
	}
	
	return results
}

// filterNamedEntities removes common false positives from named entity extraction
func (na *NLPAnalyzer) filterNamedEntities(entities []string) []string {
	// List of common words that might be mistaken as names
	stopWords := map[string]bool{
		"The": true, "A": true, "An": true, "This": true, "That": true,
		"These": true, "Those": true, "It": true, "They": true, "I": true,
		"We": true, "You": true, "He": true, "She": true, "His": true,
		"Her": true, "Their": true, "Our": true, "Your": true, "Its": true,
		"January": true, "February": true, "March": true, "April": true,
		"May": true, "June": true, "July": true, "August": true,
		"September": true, "October": true, "November": true, "December": true,
		"Monday": true, "Tuesday": true, "Wednesday": true, "Thursday": true,
		"Friday": true, "Saturday": true, "Sunday": true,
	}
	
	var filtered []string
	for _, entity := range entities {
		words := strings.Fields(entity)
		if len(words) == 0 {
			continue
		}
		
		// Skip if it's a single word that's in our stop list
		if len(words) == 1 && stopWords[words[0]] {
			continue
		}
		
		// Skip if it starts with a stop word and is only two words
		if len(words) == 2 && stopWords[words[0]] {
			continue
		}
		
		filtered = append(filtered, entity)
	}
	
	return filtered
}

// preprocessText cleans and normalizes text for analysis
func (na *NLPAnalyzer) preprocessText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	
	// Remove punctuation except for apostrophes in contractions
	text = regexp.MustCompile(`[^\w\s']`).ReplaceAllString(text, " ")
	
	// Replace multiple spaces with a single space
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}

// extractRelevantDescription finds the most relevant sentence describing a relationship
func (na *NLPAnalyzer) extractRelevantDescription(text, source, target, relType string) string {
	// Split text into sentences
	sentences := na.splitIntoSentences(text)
	
	// Look for sentences containing both names and relationship keywords
	sourcePattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(source) + `\b`)
	targetPattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(target) + `\b`)
	
	// Get relevant keywords for the relationship type
	var keywords []string
	if relCorpus, exists := na.relationshipCorpus[relType]; exists {
		for word := range relCorpus {
			keywords = append(keywords, word)
		}
	}
	
	// Score each sentence based on relevance
	type scoredSentence struct {
		text  string
		score int
	}
	
	var scoredSentences []scoredSentence
	
	for _, sentence := range sentences {
		score := 0
		
		// Check if sentence contains both names
		if sourcePattern.MatchString(sentence) {
			score += 2
		}
		if targetPattern.MatchString(sentence) {
			score += 2
		}
		
		// Check for relationship keywords
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(sentence), keyword) {
				score += 3
				break
			}
		}
		
		if score > 0 {
			scoredSentences = append(scoredSentences, scoredSentence{
				text:  sentence,
				score: score,
			})
		}
	}
	
	// Sort by score (higher is better)
	var bestSentence string
	bestScore := -1
	
	for _, scored := range scoredSentences {
		if scored.score > bestScore {
			bestScore = scored.score
			bestSentence = scored.text
		}
	}
	
	// If no good sentence found, create a generic description
	if bestSentence == "" {
		return fmt.Sprintf("%s and %s were connected in historical context.", source, target)
	}
	
	// Clean up the sentence
	bestSentence = regexp.MustCompile(`\[\d+\]`).ReplaceAllString(bestSentence, "")
	bestSentence = regexp.MustCompile(`\s+`).ReplaceAllString(bestSentence, " ")
	
	// Truncate if too long
	if len(bestSentence) > 200 {
		bestSentence = bestSentence[:197] + "..."
	}
	
	return strings.TrimSpace(bestSentence)
}

// splitIntoSentences divides text into individual sentences
func (na *NLPAnalyzer) splitIntoSentences(text string) []string {
	// Basic sentence splitting with regex
	// This is simplified and could be improved
	sentenceEndings := regexp.MustCompile(`[.!?][\s)]`)
	boundaries := sentenceEndings.FindAllStringIndex(text, -1)
	
	if len(boundaries) == 0 {
		return []string{text}
	}
	
	var sentences []string
	lastStart := 0
	
	for _, boundary := range boundaries {
		end := boundary[1] - 1 // Exclude the space or parenthesis after the punctuation
		sentences = append(sentences, text[lastStart:end])
		lastStart = end + 1
	}
	
	// Add the last sentence if there's text remaining
	if lastStart < len(text) {
		sentences = append(sentences, text[lastStart:])
	}
	
	return sentences
}

// UpdateCorpusFromText learns new relationship patterns from text
func (na *NLPAnalyzer) UpdateCorpusFromText(text, relType string) {
	// Only proceed if it's a known relationship type
	na.mu.RLock()
	_, exists := na.relationshipCorpus[relType]
	na.mu.RUnlock()
	
	if !exists {
		return
	}
	
	// Preprocess text and extract words
	processedText := na.preprocessText(text)
	words := strings.Fields(processedText)
	
	// Extract significant words (excluding common stopwords)
	stopwords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, 
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "with": true, "by": true, "as": true, "of": true,
		"from": true, "was": true, "were": true, "is": true, "are": true,
		"be": true, "been": true, "has": true, "have": true, "had": true,
	}
	
	// Count word frequencies
	wordCounts := make(map[string]int)
	for _, word := range words {
		if !stopwords[word] && len(word) > 2 {
			wordCounts[word]++
		}
	}
	
	// Update the corpus with high-frequency words
	na.mu.Lock()
	defer na.mu.Unlock()
	
	for word, count := range wordCounts {
		if count >= 2 { // Only add if it appears multiple times
			// If word already exists, increase its weight
			if _, wordExists := na.relationshipCorpus[relType][word]; wordExists {
				na.relationshipCorpus[relType][word] += 1
				// Cap at maximum weight
				if na.relationshipCorpus[relType][word] > 10 {
					na.relationshipCorpus[relType][word] = 10
				}
			} else {
				// Add new word with low initial weight
				na.relationshipCorpus[relType][word] = 3
			}
		}
	}
}

// GetTopRelationshipIndicators returns the top words indicating a relationship type
func (na *NLPAnalyzer) GetTopRelationshipIndicators(relType string, topN int) []string {
	na.mu.RLock()
	defer na.mu.RUnlock()
	
	corpus, exists := na.relationshipCorpus[relType]
	if !exists {
		return nil
	}
	
	// Create a slice of word-weight pairs
	type wordWeight struct {
		word   string
		weight int
	}
	
	var pairs []wordWeight
	for word, weight := range corpus {
		pairs = append(pairs, wordWeight{word, weight})
	}
	
	// Sort by weight (descending)
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].weight < pairs[j].weight {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}
	
	// Get top N words
	var result []string
	for i := 0; i < min(topN, len(pairs)); i++ {
		result = append(result, pairs[i].word)
	}
	
	return result
}

// Helper function for Go versions that don't have built-in min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}