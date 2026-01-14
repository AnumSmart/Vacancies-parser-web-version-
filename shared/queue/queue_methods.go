package queue

import "sync/atomic"

// метод для добавления нового элемента в очередь
func (q *FIFOQueue[T]) Enqueue(item T) bool {
	// Атомарная проверка - не блокирует другие горутины
	if atomic.LoadInt32(&q.closed) == 1 {
		return false
	}

	select {
	case q.items <- item:
		return true
	default:
		return false // очередь переполнена
	}
}

// метод для получения элемента из очереди
func (q *FIFOQueue[T]) Dequeue() (T, bool) {
	var zeroVal T
	select {
	case item, ok := <-q.items:
		if ok {
			return item, true
		} else {
			return zeroVal, false
		}
	default:
		return zeroVal, false // очередь пуста
	}
}

// метод для получения размера очереди в данный момент
func (q *FIFOQueue[T]) Size() int {
	return len(q.items)
}

// Close безопасно закрывает очередь
// После закрытия добавление элементов невозможно, а чтение вернет оставшиеся элементы
func (q *FIFOQueue[T]) Close() {
	// CAS гарантирует, что закрываем только один раз
	if atomic.CompareAndSwapInt32(&q.closed, 0, 1) {
		close(q.items)
	}
}

// Clear безопасно очищает очередь, только если очередб не была закрыта.
func (q *FIFOQueue[T]) Clear() {
	// Быстрая проверка без блокировки
	if atomic.LoadInt32(&q.closed) == 1 {
		return
	}

	// Вычитываем все элементы неблокирующим способом
	for {
		select {
		case <-q.items:
			// Элемент удален из канала
			// Если T содержит ресурсы, они будут собраны GC
		default:
			// Канал пуст - завершаем очистку
			return
		}
	}
}
