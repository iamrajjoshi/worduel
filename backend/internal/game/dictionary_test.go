package game

import (
	"testing"
)

func TestNewDictionary(t *testing.T) {
	dict := NewDictionary()
	
	if dict == nil {
		t.Fatal("Expected dictionary to be created, got nil")
	}
	
	commonCount := dict.GetCommonWordsCount()
	validCount := dict.GetValidWordsCount()
	
	if commonCount == 0 {
		t.Error("Expected common words to be loaded")
	}
	
	if validCount == 0 {
		t.Error("Expected valid words to be loaded")
	}
	
	t.Logf("Loaded %d common words and %d valid words", commonCount, validCount)
}

func TestIsValidGuess(t *testing.T) {
	dict := NewDictionary()
	
	tests := []struct {
		word     string
		expected bool
	}{
		{"about", true},  // Should be in the word list
		{"hello", false}, // Not in our word list
		{"ABOUT", true},  // Should handle uppercase
		{"abc", false},   // Too short
		{"abcdef", false}, // Too long
		{"", false},      // Empty string
	}
	
	for _, test := range tests {
		result := dict.IsValidGuess(test.word)
		if result != test.expected {
			t.Errorf("IsValidGuess(%q) = %v, expected %v", test.word, result, test.expected)
		}
	}
}

func TestGetRandomTarget(t *testing.T) {
	dict := NewDictionary()
	
	word := dict.GetRandomTarget()
	
	if len(word) != 5 {
		t.Errorf("Expected 5-letter word, got %q (length %d)", word, len(word))
	}
	
	if !dict.IsValidGuess(word) {
		t.Errorf("Random target word %q should be valid", word)
	}
	
	// Test that we get different words (with some probability)
	words := make(map[string]bool)
	for i := 0; i < 10; i++ {
		words[dict.GetRandomTarget()] = true
	}
	
	// We should get at least 2 different words in 10 attempts (very high probability)
	if len(words) < 2 {
		t.Log("Warning: GetRandomTarget might not be sufficiently random")
	}
}

func TestDictionaryConcurrency(t *testing.T) {
	dict := NewDictionary()
	
	done := make(chan bool, 10)
	
	// Test concurrent access
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// Test concurrent reads
			for j := 0; j < 100; j++ {
				word := dict.GetRandomTarget()
				if !dict.IsValidGuess(word) {
					t.Errorf("Concurrent test failed: random word %q is not valid", word)
					return
				}
			}
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}