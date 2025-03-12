package service

import (
	"strings"
	"sync"

	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"github.com/kljensen/snowball"
	"gonum.org/v1/gonum/mat"
)

type Preset struct {
	ID    int
	Query string
}

type Storage struct {
	mu          sync.RWMutex
	presets     []Preset
	vectoriser  *nlp.CountVectoriser
	transformer *nlp.TfidfTransformer
}

func NewStorage() *Storage {
	vectoriser := nlp.NewCountVectoriser()
	transformer := nlp.NewTfidfTransformer()

	return &Storage{
		presets:     make([]Preset, 0),
		vectoriser:  vectoriser,
		transformer: transformer,
	}
}

func (s *Storage) AddPreset(preset Preset) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.presets = append(s.presets, preset)
}

func (s *Storage) FindClosestPreset(query string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	queryWords := strings.Fields(query)
	for i, word := range queryWords {
		stemmed, _ := snowball.Stem(word, "russian", true)
		queryWords[i] = stemmed
	}
	processedQuery := strings.Join(queryWords, " ")

	//TODO реалиховать сохранение пресетов с уже ловеркейсом и чекать на совпадение ниже
	for _, preset := range s.presets {
		if strings.ToLower(preset.Query) == processedQuery {
			return preset.ID, true
		}
	}

	var bestPresetID int
	var bestScore float64
	found := false

	queries := make([]string, len(s.presets))
	for i, preset := range s.presets {
		queries[i] = preset.Query
	}

	//TODO: Рассмотреть вариант переиспользования данной матрицы
	matrix, _ := s.vectoriser.FitTransform(queries...)
	tfidf, _ := s.transformer.FitTransform(matrix)

	queryMatrix, _ := s.vectoriser.Transform(processedQuery)
	queryTfidf, _ := s.transformer.Transform(queryMatrix)

	queryVector := queryTfidf.(*mat.Dense).RowView(0)

	for i, preset := range s.presets {
		presetVector := tfidf.(*mat.Dense).RowView(i)

		score := pairwise.CosineSimilarity(queryVector, presetVector)
		if score > bestScore {
			bestScore = score
			bestPresetID = preset.ID
			found = true
		}
	}

	if bestScore < 0.5 { // Порог можно настроить
		return 0, false
	}

	return bestPresetID, found
}
