package closer

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Closer - управляет закрытием зависимостей
type Closer struct {
	closers []io.Closer
	mu      sync.Mutex
	closed  bool
}

// New - создает новый экземпляр Closer
func New() *Closer {
	return &Closer{
		closers: make([]io.Closer, 0),
	}
}

// Add - добавляет зависимость для закрытия
func (c *Closer) Add(closer io.Closer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		_ = closer.Close()
		return
	}

	c.closers = append(c.closers, closer)
}

// AddFunc - добавляет функцию как io.Closer
func (c *Closer) AddFunc(fn func() error) {
	c.Add(closerFunc(fn))
}

// Close - закрывает все зависимости в обратном порядке
func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	closers := c.closers
	c.closers = nil
	c.mu.Unlock()

	var errs []error

	// Закрываем в обратном порядке (LIFO)
	for i := len(closers) - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return fmt.Errorf("close interrupted: %w", ctx.Err())
		default:
			if err := closers[i].Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// Len - количество зарегистрированных зависимостей
func (c *Closer) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.closers)
}

// closerFunc - адаптер для функции
type closerFunc func() error

func (f closerFunc) Close() error {
	return f()
}
