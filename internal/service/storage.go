package service

import (
	"github.com/sirupsen/logrus"
	"strings"
	"sync"

	"github.com/kljensen/snowball"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

type Preset struct {
	ID             int
	Query          string
	ProcessedQuery string
}

type Storage struct {
	mu      sync.RWMutex
	presets []Preset
}

func NewStorage() *Storage {
	return &Storage{
		presets: make([]Preset, 0),
	}
}

// preprocessText: lowercase + stemming
func preprocessText(query string) string {
	query = strings.ToLower(query)
	queryWords := strings.Fields(query)
	for i, word := range queryWords {
		stemmed, err := snowball.Stem(word, "russian", true)
		if err == nil {
			queryWords[i] = stemmed
		}
	}
	return strings.Join(queryWords, " ")
}

func (s *Storage) AddPreset(preset Preset) {
	s.mu.Lock()
	defer s.mu.Unlock()

	preset.ProcessedQuery = preprocessText(preset.Query)
	s.presets = append(s.presets, preset)
}

func (s *Storage) FindClosestPreset(query string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	processedQuery := preprocessText(query)

	bestPresetID := -1
	bestDistance := -1 // меньше — лучше (ближе)

	for _, preset := range s.presets {
		distance := levenshtein.DistanceForStrings(
			[]rune(processedQuery),
			[]rune(preset.ProcessedQuery),
			levenshtein.DefaultOptions,
		)

		// выбираем самый близкий пресет
		if bestDistance == -1 || distance < bestDistance {
			bestDistance = distance
			bestPresetID = preset.ID
		}
	}

	// Порог можно менять. Например, допустимая дистанция до 4 символов
	const maxDistance = 4
	if bestDistance > maxDistance {
		logrus.Info(bestDistance, bestPresetID)
		return 0, false
	}

	return bestPresetID, true
}
