package queue

// структура для очереди
type FIFOQueue[T any] struct {
	items  chan T
	closed int32 // 0 = открыт, 1 = закрыт
}

// конструктор для очереди
func NewFIFOQueue[T any](capacity int) *FIFOQueue[T] {
	return &FIFOQueue[T]{
		items:  make(chan T, capacity),
		closed: 0, // при создании экзмепляра очереди устанавливаем в флаг 0. Канал открыт
	}
}
