package interfaces

// Интерфейс с дженериком для FIFO очереди
type FIFOQueueInterface[T any] interface {
	Enqueue(item T) bool // потокобезопасен, так как очередь построена на базе каналов
	Dequeue() (T, bool)
	Size() int
}
