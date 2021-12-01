package alloter

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"runtime"
	"sync"
)

var ErrorUsingAlloter = fmt.Errorf("ErrorUsingActuator")

type GoroutinePool interface {
	Submit(f func()) error
	Release()
}

type PooledAlloter struct {
	workerNum int
	pool      GoroutinePool
	initOnce sync.Once
}

func NewPooledAlloter(workerNum int) *PooledAlloter {
	return &PooledAlloter{
		workerNum: workerNum,
	}
}

// WithPool will support for using custom goroutine pool
func (c *PooledAlloter) WithPool(pool GoroutinePool) *PooledAlloter {
	newAlloter := c.clone()
	newAlloter.pool = pool
	return newAlloter
}

// Exec is used to run tasks concurrently
func (c *PooledAlloter) Exec(tasks *[]Task) error {
	return c.ExecWithContext(context.Background(), tasks)
}

func (c *PooledAlloter) ExecWithContext(ctx context.Context, tasks *[]Task) error {
	defer c.Release()
	c.initOnce.Do(func() {
		c.initPooledAlloter()
	})

	if c.workerNum == -1 {
		return ErrorUsingAlloter
	}

	return c.execTasks(ctx, tasks)
}

func (c *PooledAlloter) Release() {
	if c.pool != nil {
		c.pool.Release()
	}
}

func (c *PooledAlloter) initPooledAlloter() {
	if c.pool != nil {
		c.workerNum = 1
		return
	}

	if c.workerNum <= 0 {
		c.workerNum = runtime.NumCPU() << 1
	}

	var err error
	c.pool, err = ants.NewPool(c.workerNum)

	if err != nil {
		c.workerNum = -1
		panic(err)
	}
}

// clone will clone this PooledAlloter without goroutine pool
func (c *PooledAlloter) clone() *PooledAlloter {
	return &PooledAlloter{
		workerNum: c.workerNum,
		initOnce:  sync.Once{},
	}
}

func (c *PooledAlloter) execTasks(ctx context.Context, tasks *[]Task) error {
	size := len(*tasks)
	if size == 0 {
		return nil
	}

	resChan := make(chan error, size)
	errChan := make(chan error, size)
	wg := sync.WaitGroup{}
	wg.Add(size)

	for _, task := range *tasks {
		end, err := noBlockGo(ctx, &errChan)
		if end {
			return err
		}
		f := wrapperTask(ctx, task, &wg, &resChan, &errChan)
		err = c.pool.Submit(f)
		if err != nil {
			return err
		}
	}

	// When error, wo can't close resChan, maybe some goroutines just finished.
	// So, when error, wo just can wait auto GC.
	child, cancel := context.WithCancel(ctx)
	go func() {
		wg.Wait()
		cancel()
		close(resChan)
		close(errChan)
	}()
	return blockGo(child, &errChan)
}
