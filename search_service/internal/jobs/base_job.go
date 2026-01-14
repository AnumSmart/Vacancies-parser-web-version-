// каждая джоба выполняется одним воркером
package jobs

import (
	"log"
	"sync"
	"time"
)

// структура результата выполнения джобы
type JobOutput struct {
	Success bool
	Data    interface{}
	Error   error
}

// структура базовой джобы, которая будет общей для всех других джоб (задач в очереди)
type BaseJob struct {
	ID         string
	ResultChan chan *JobOutput // обязательно при создании экземплярар джобы нужно делать буферизированный канал, 1
	CreatedAt  time.Time
	notified   sync.Once
}

// метод у структуры джоб, который отправляет результат в результирующий канал
func (j *BaseJob) Complete(data interface{}, err error) {
	j.notified.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				// Канал закрыт (очень редкий случай)
				log.Printf("job %s: канал закрыт при отправке результата", j.ID)
			}
		}()
		if err == nil {
			j.ResultChan <- &JobOutput{Success: true, Data: data, Error: err}
		}

		j.ResultChan <- &JobOutput{Success: false, Data: data, Error: err}
	})
}

// возвращает ID джобы
func (j *BaseJob) GetID() string {
	return j.ID
}
