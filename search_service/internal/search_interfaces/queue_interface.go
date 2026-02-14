package search_interfaces

// Интерфейс с дженериком для FIFO очереди
type FIFOQueueInterface[T any] interface {
	Enqueue(item T) bool // должен быть потокобезопасен
	Dequeue() (T, bool)
	Size() int
}
