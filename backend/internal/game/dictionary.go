package game

import (
	_ "embed"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Embed word list files
//go:embed common.txt
var commonWordsData string

//go:embed valid.txt
var validWordsData string

// Dictionary manages word lists and validation
type Dictionary struct {
	commonWords []string
	validWords  map[string]bool
	rand        *rand.Rand
	mutex       sync.RWMutex
}

// NewDictionary creates a new dictionary with embedded word lists
func NewDictionary() *Dictionary {
	d := &Dictionary{
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		validWords: make(map[string]bool),
	}

	// Parse common words (used for target words)
	commonLines := strings.Split(strings.TrimSpace(commonWordsData), "\n")
	for _, line := range commonLines {
		word := strings.TrimSpace(strings.ToLower(line))
		if len(word) == 5 { // Only include 5-letter words
			d.commonWords = append(d.commonWords, word)
		}
	}

	// Parse valid words (used for guess validation)
	validLines := strings.Split(strings.TrimSpace(validWordsData), "\n")
	for _, line := range validLines {
		word := strings.TrimSpace(strings.ToLower(line))
		if len(word) == 5 { // Only include 5-letter words
			d.validWords[word] = true
		}
	}

	// Add common words to valid words set
	for _, word := range d.commonWords {
		d.validWords[word] = true
	}

	return d
}

// IsValidGuess checks if a word is a valid guess
func (d *Dictionary) IsValidGuess(word string) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	normalized := strings.TrimSpace(strings.ToLower(word))
	if len(normalized) != 5 {
		return false
	}

	return d.validWords[normalized]
}

// GetRandomTarget returns a random word from the common words list
func (d *Dictionary) GetRandomTarget() string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if len(d.commonWords) == 0 {
		return "wordle" // Fallback word
	}

	index := d.rand.Intn(len(d.commonWords))
	return d.commonWords[index]
}

// GetCommonWordsCount returns the number of words available for targets
func (d *Dictionary) GetCommonWordsCount() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.commonWords)
}

// GetValidWordsCount returns the number of words available for validation
func (d *Dictionary) GetValidWordsCount() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.validWords)
}