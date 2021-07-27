package alloter

import (
	"context"
	"sync"
	"time"
)

type CtrlAlloter struct {
	timeout time.Time
	workerNum int
	ctrlChan chan struct{}
}

func NewCtrlAlloter(workerNum int, opt *Options) *CtrlAlloter {
	c := &CtrlAlloter{
		workerNum: workerNum,
		ctrlChan: make(chan struct{}, workerNum),
	}
	if opt != nil && opt.TimeOut != 0 {
		c.timeout = time.Now().Add(opt.TimeOut)
	}
	return c
}

func (c *CtrlAlloter) Exec(tasks *[]Task) error {
	return c.execTasks(context.Background(), tasks)
}

func (c *CtrlAlloter) ExecWithContext(ctx context.Context, tasks *[]Task) error {
	return c.execTasks(ctx, tasks)
}

func (c *CtrlAlloter) GetTimeout() time.Time {
	return c.timeout
}

func (c *CtrlAlloter) setTimeout(timeout time.Time) {
	c.timeout = timeout
}

func (c *CtrlAlloter) execTasks(parent context.Context, tasks *[]Task) error {
	size := len(*tasks)
	if size == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	resChan := make(chan error, size)
	errChan := make(chan error, size)
	wg := sync.WaitGroup{}
	wg.Add(size)

	timeout := c.GetTimeout()
	for _, task := range *tasks {
		c.ctrlChan <- struct{}{}
		end, err := noblockGo(ctx, cancel, &errChan, timeout)
		if end {
			return err
		}
		f := wrapperTask(ctx, cancel, task, &wg, &resChan, &errChan, timeout)
		go func() {
			f()
			<- c.ctrlChan
		}()
	}

	// When error, wo can't close resChan, maybe some goroutines just finished.
	// So, when error, wo just can wait auto GC.
	go func() {
		wg.Wait()
		cancel()
		close(resChan)
		close(errChan)
	}()
	return blockGo(ctx, cancel, &errChan, timeout)
}
