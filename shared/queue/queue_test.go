package queue

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestNewFIFOQueue проверяет создание новой очереди
func TestNewFIFOQueue(t *testing.T) {
	capacity := 5
	q := NewFIFOQueue[int](capacity)

	if q == nil {
		t.Fatal("Очередь не должна быть nil")
	}

	if cap(q.items) != capacity {
		t.Errorf("Неверная емкость очереди. Ожидалось: %d, получено: %d", capacity, cap(q.items))
	}

	if atomic.LoadInt32(&q.closed) != 0 {
		t.Error("При создании очередь должна быть открыта")
	}
}

// TestEnqueueDequeue проверяет базовые операции добавления и извлечения
func TestEnqueueDequeue(t *testing.T) {
	q := NewFIFOQueue[int](3)

	// Тестируем добавление
	if !q.Enqueue(1) {
		t.Error("Не удалось добавить элемент 1")
	}
	if !q.Enqueue(2) {
		t.Error("Не удалось добавить элемент 2")
	}
	if !q.Enqueue(3) {
		t.Error("Не удалось добавить элемент 3")
	}

	// Очередь должна быть полной
	if q.Enqueue(4) {
		t.Error("Очередь не должна принимать элементы при переполнении")
	}

	// Проверяем размер
	if size := q.Size(); size != 3 {
		t.Errorf("Неверный размер очереди. Ожидалось: 3, получено: %d", size)
	}

	// Проверяем извлечение в правильном порядке (FIFO)
	val, ok := q.Dequeue()
	if !ok || val != 1 {
		t.Errorf("Неверный элемент извлечен. Ожидалось: 1, получено: %v, ok: %v", val, ok)
	}

	val, ok = q.Dequeue()
	if !ok || val != 2 {
		t.Errorf("Неверный элемент извлечен. Ожидалось: 2, получено: %v, ok: %v", val, ok)
	}

	// Добавляем еще один элемент после освобождения места
	if !q.Enqueue(4) {
		t.Error("Не удалось добавить элемент 4 после освобождения места")
	}

	val, ok = q.Dequeue()
	if !ok || val != 3 {
		t.Errorf("Неверный элемент извлечен. Ожидалось: 3, получено: %v, ok: %v", val, ok)
	}

	val, ok = q.Dequeue()
	if !ok || val != 4 {
		t.Errorf("Неверный элемент извлечен. Ожидалось: 4, получено: %v, ok: %v", val, ok)
	}

	// Проверяем извлечение из пустой очереди
	val, ok = q.Dequeue()
	if ok || val != 0 {
		t.Error("Извлечение из пустой очереди должно возвращать false и zero value")
	}
}

// TestClose проверяет закрытие очереди
func TestClose(t *testing.T) {
	q := NewFIFOQueue[string](3)

	// Добавляем элементы перед закрытием
	q.Enqueue("first")
	q.Enqueue("second")

	q.Close()

	// После закрытия нельзя добавлять новые элементы
	if q.Enqueue("third") {
		t.Error("Нельзя добавлять элементы в закрытую очередь")
	}

	// Но можно извлекать оставшиеся элементы
	val, ok := q.Dequeue()
	if !ok || val != "first" {
		t.Errorf("Ожидалось 'first', получено: %v, ok: %v", val, ok)
	}

	val, ok = q.Dequeue()
	if !ok || val != "second" {
		t.Errorf("Ожидалось 'second', получено: %v, ok: %v", val, ok)
	}

	// После извлечения всех элементов возвращаем false
	val, ok = q.Dequeue()
	if ok {
		t.Error("После извлечения всех элементов из закрытой очереди должен возвращаться false")
	}

	// Повторное закрытие не должно паниковать
	q.Close()
}

// TestConcurrentAccess проверяет конкурентный доступ
func TestConcurrentAccess(t *testing.T) {
	q := NewFIFOQueue[int](100)

	const workers = 10
	const itemsPerWorker = 100

	var wg sync.WaitGroup
	wg.Add(workers * 2) // Писатели и читатели

	// Атомарные счетчики для статистики
	var writtenCount, readCount int64

	// Писатели
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerWorker; j++ {
				value := id*itemsPerWorker + j
				if q.Enqueue(value) {
					atomic.AddInt64(&writtenCount, 1)
				}
				// Если не удалось записать (очередь полна) - это нормально
			}
		}(i)
	}

	// Читатели
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < itemsPerWorker; j++ {
				_, ok := q.Dequeue()
				if ok {
					atomic.AddInt64(&readCount, 1)
				}
			}
		}()
	}

	wg.Wait()

	// Проверяем базовые инварианты
	written := atomic.LoadInt64(&writtenCount)
	read := atomic.LoadInt64(&readCount)

	t.Logf("Статистика: записано %d, прочитано %d", written, read)

	// 1. Не может быть прочитано больше, чем записано
	if read > written {
		t.Errorf("Прочитано больше, чем записано: %d > %d", read, written)
	}

	// 2. Размер очереди должен соответствовать разнице
	finalSize := q.Size()
	expectedSize := int(written - read)
	if finalSize != expectedSize {
		t.Errorf("Несоответствие размера очереди: ожидалось %d, получено %d",
			expectedSize, finalSize)
	}
}

// TestClear проверяет очистку очереди (очередь должна быть открыта)
func TestClear(t *testing.T) {
	q := NewFIFOQueue[int](10)

	// Заполняем очередь
	for i := 0; i < 10; i++ {
		q.Enqueue(i)
	}

	if size := q.Size(); size != 10 {
		t.Errorf("Перед очисткой размер должен быть 10, но получено: %d", size)
	}

	// Очищаем
	q.Clear()

	if size := q.Size(); size != 0 {
		t.Errorf("После очистки размер должен быть 0, но получено: %d", size)
	}

	// Проверяем, что очередь работает после очистки
	if !q.Enqueue(42) {
		t.Error("Не удалось добавить элемент после очистки")
	}

	val, ok := q.Dequeue()
	if !ok || val != 42 {
		t.Errorf("Не удалось извлечь элемент после очистки. Ожидалось: 42, получено: %v", val)
	}

	// Очистка закрытой очереди не должна паниковать
	q.Enqueue(1)
	q.Enqueue(2)
	q.Close()
	q.Clear() // Должно работать без паники
}

// TestTypeGeneric проверяет работу с разными типами
func TestTypeGeneric(t *testing.T) {
	// Тестируем с int
	q1 := NewFIFOQueue[int](3)
	q1.Enqueue(42)
	val, _ := q1.Dequeue()
	if val != 42 {
		t.Errorf("Для int: ожидалось 42, получено %v", val)
	}

	// Тестируем со string
	q2 := NewFIFOQueue[string](3)
	q2.Enqueue("hello")
	str, _ := q2.Dequeue()
	if str != "hello" {
		t.Errorf("Для string: ожидалось 'hello', получено %v", str)
	}

	// Тестируем со структурой
	type Person struct {
		Name string
		Age  int
	}
	q3 := NewFIFOQueue[Person](3)
	person := Person{Name: "Alice", Age: 30}
	q3.Enqueue(person)
	p, _ := q3.Dequeue()
	if p.Name != "Alice" || p.Age != 30 {
		t.Errorf("Для структуры: ожидалось {Alice 30}, получено %v", p)
	}
}

// TestStress тестирует производительность
func TestStressLimitedCapacity(t *testing.T) {
	const capacity = 1000
	const totalOperations = 10000

	q := NewFIFOQueue[int](capacity)

	var wg sync.WaitGroup
	wg.Add(2)

	// Счетчики для проверки
	var enqueuedCount, dequeuedCount int32

	// Писатель - пытается записать все элементы
	go func() {
		defer wg.Done()
		for i := 0; i < totalOperations; i++ {
			// Пытаемся записать с повторными попытками
			for !q.Enqueue(i) {
			}
			atomic.AddInt32(&enqueuedCount, 1)
		}
	}()

	// Читатель - читает с задержкой
	go func() {
		defer wg.Done()
		for i := 0; i < totalOperations; i++ {
			// Пытаемся прочитать
			for {
				val, ok := q.Dequeue()
				if ok {
					if val != i {
						t.Errorf("Неверный порядок элементов: ожидалось %d, получено %d", i, val)
					}
					atomic.AddInt32(&dequeuedCount, 1)
					break
				}
			}
		}
	}()

	wg.Wait()

	// Проверяем, что все элементы обработаны
	if enqueuedCount != totalOperations {
		t.Errorf("Записано не все: %d из %d", enqueuedCount, totalOperations)
	}
	if dequeuedCount != totalOperations {
		t.Errorf("Прочитано не все: %d из %d", dequeuedCount, totalOperations)
	}
}
