package alloter

import (
	"context"
	"sync"
	"time"
)

type Alloter struct {
	timeout time.Time
}

func NewAlloter(opt *Options) *Alloter {
	c := &Alloter{}
	if opt != nil && opt.TimeOut != 0 {
		c.timeout = time.Now().Add(opt.TimeOut)
	}
	return c
}

func (c *Alloter) Exec(tasks *[]Task) error {
	return c.execTasks(context.Background(), tasks)
}

func (c *Alloter) ExecWithContext(ctx context.Context, tasks *[]Task) error {
	return c.execTasks(ctx, tasks)
}

func (c *Alloter) GetTimeout() time.Time {
	return c.timeout
}

func (c *Alloter) setTimeout(timeout time.Time) {
	c.timeout = timeout
}

func (c *Alloter) execTasks(parent context.Context, tasks *[]Task) error {
	size := len(*tasks)
	if size == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	errChan := make(chan error, size)
	wg := sync.WaitGroup{}
	wg.Add(size)

	timeout := c.GetTimeout()
	for _, task := range *tasks {
		f := wrapperSimpleTask(task, &wg, &errChan)
		go f()
	}

	// When error, wo can't close resChan, maybe some goroutines just finished.
	// So, when error, wo just can wait auto GC.
	go func() {
		wg.Wait()
		cancel()
		close(errChan)
	}()
	return blockGo(ctx, cancel, &errChan, timeout)
}

