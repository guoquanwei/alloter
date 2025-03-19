package alloter

import (
	"context"
	"sync"
)

type Alloter struct{}

func NewAlloter() *Alloter {
	return &Alloter{}
}

func (c *Alloter) Exec(tasks []Task) error {
	return c.execTasks(context.Background(), tasks)
}

func (c *Alloter) ExecWithContext(ctx context.Context, tasks []Task) error {
	return c.execTasks(ctx, tasks)
}

func (c *Alloter) execTasks(ctx context.Context, tasks []Task) error {
	size := len(tasks)
	if size == 0 {
		return nil
	}
	errChan := make(chan error, size)
	wg := sync.WaitGroup{}
	wg.Add(size)

	for _, task := range tasks {
		f := wrapperSimpleTask(task, &wg, &errChan)
		go f()
	}

	// When error, wo can't close resChan, maybe some goroutines just finished.
	// So, when error, wo just can wait auto GC.
	child, cancel := context.WithCancel(ctx)
	go func() {
		wg.Wait()
		cancel()
		close(errChan)
	}()
	return blockGo(child, &errChan)
}
