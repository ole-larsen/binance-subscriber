package storage

import (
	"sort"
	"sync"
)

type Data struct {
	Symbol string `json:"symbol"`
	Bid    string `json:"bid"`
	Ask    string `json:"ask"`
}

type MemStorage struct {
	storage map[string]Data
	mx      sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: make(map[string]Data),
	}
}

func (m *MemStorage) Set(data Data) {
	m.mx.Lock()
	m.storage[data.Symbol] = data
	m.mx.Unlock()
}

func (m *MemStorage) Get(symbol string) *Data {
	m.mx.Lock()

	defer m.mx.Unlock()

	if data, ok := m.storage[symbol]; ok {
		return &data
	}

	return nil
}

func (m *MemStorage) GetAll() []*Data {
	m.mx.Lock()

	defer m.mx.Unlock()

	symbols := make([]*Data, 0, len(m.storage))

	for _, data := range m.storage {
		symbols = append(symbols, &data)
	}

	// sort to keep ordering
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Symbol < symbols[j].Symbol
	})

	return symbols
}
