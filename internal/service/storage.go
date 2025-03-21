package service

import (
	"sort"
	"strings"
	"sync"

	"github.com/kljensen/snowball"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

const maxDistance = 4

type Product struct {
	Id    uint32
	Name  string
	Price uint32
	Score int16
}

type Preset struct {
	Id             uint32
	processedQuery string
	products       []Product
	isDone         bool
}

type Storage struct {
	mu      sync.RWMutex
	presets []Preset
}

func NewStorage() *Storage {
	return &Storage{}
}

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

func (s *Storage) CreateNewPreset(query string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	preset := Preset{
		Id:             uint32(len(s.presets)),
		processedQuery: preprocessText(query),
	}

	s.presets = append(s.presets, preset)
	//TODO: Здесь будет отправка запроса на майнинг в кафку
}

// TODO: Скорее всего мы будем слушать кафку и обновлять продукты || продукты будем получать сразу отсортированные по скору, чтобы не грузить сервис выдачи
func (s *Storage) UpdateProductsPreset(presetId int, products []Product) {
	if presetId >= len(s.presets) || presetId < 0 || len(products) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.presets[presetId].products = products
	s.presets[presetId].isDone = true
}

//TODO: Тут так же будем слушать скорее всего кафку. Будет возможность обновлять отдельно товары(в случае если они только появились тобишь новые)

type PresetsToProductsScore struct {
	PresetId uint32
	Score    int16
}
type ProductPresets struct {
	ProductId uint32
	Name      string
	Price     uint32
	Presets   []PresetsToProductsScore
}

func (s *Storage) UpdateProductByPresets(product ProductPresets) {
	p := Product{
		Id:    product.ProductId,
		Name:  product.Name,
		Price: product.Price,
	}

	for _, presetUpdate := range product.Presets {
		if int(presetUpdate.PresetId) >= len(s.presets) {
			continue
		}

		s.mu.RLock()
		preset := s.presets[presetUpdate.PresetId]
		if !preset.isDone {
			continue
		}

		presetProducts := make([]Product, len(preset.products), len(preset.products)+1)
		copy(presetProducts, preset.products)
		s.mu.RUnlock()

		p.Score = presetUpdate.Score
		presetProducts = append(presetProducts, p)

		sort.Slice(preset.products, func(i, j int) bool {
			return preset.products[i].Score > preset.products[j].Score
		})
	}
}

func (s *Storage) FindClosestPreset(query string) ([]Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	processedQuery := preprocessText(query)

	var (
		bestDistance = 1000
		products     []Product
	)

	for _, preset := range s.presets {
		if !preset.isDone {
			continue
		}

		distance := levenshtein.DistanceForStrings(
			[]rune(processedQuery),
			[]rune(preset.processedQuery),
			levenshtein.DefaultOptions,
		)

		if distance < bestDistance {
			bestDistance = distance
			products = preset.products
		}
	}

	if bestDistance > maxDistance {
		return products, false
	}

	return products, true
}

func (s *Storage) GetAllPresets() []Preset {
	return s.presets
}
