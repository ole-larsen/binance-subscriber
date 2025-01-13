package storage

type Storage interface {
	Set(data Data)
	Get(symbol string) *Data
	GetAll() []*Data
}
