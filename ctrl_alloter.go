package alloter

import (
	"context"
	"sync"
)

type CtrlAlloter struct {
	workerNum int
	ctrlChan  chan struct{}
}

func NewCtrlAlloter(workerNum int) *CtrlAlloter {
	return &CtrlAlloter{
		workerNum: workerNum,
		ctrlChan:  make(chan struct{}, workerNum),
	}
}

func (c *CtrlAlloter) Exec(tasks []Task) error {
	return c.execTasks(context.Background(), tasks)
}

func (c *CtrlAlloter) ExecWithContext(ctx context.Context, tasks []Task) error {
	return c.execTasks(ctx, tasks)
}

func (c *CtrlAlloter) execTasks(ctx context.Context, tasks []Task) error {
	size := len(tasks)
	if size == 0 {
		return nil
	}
	resChan := make(chan error, size)
	errChan := make(chan error, size)
	wg := sync.WaitGroup{}
	wg.Add(size)

	for _, task := range tasks {
		c.ctrlChan <- struct{}{}
		end, err := noBlockGo(ctx, &errChan)
		if end {
			return err
		}
		f := wrapperTask(ctx, task, &wg, &resChan, &errChan)
		go func() {
			f()
			<-c.ctrlChan
		}()
	}

	// When error, wo can't close resChan, maybe some goroutines just finished.
	// So, when error, wo just can wait auto GC.
	go func() {
		wg.Wait()
		close(resChan)
		close(errChan)
	}()
	return blockGo(ctx, &errChan)
}
