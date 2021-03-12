package alloter

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"runtime"
	"sync"
	"time"
)

var (
	// ErrorUsingAlloter is the error when goroutine pool has exception
	ErrorUsingAlloter = fmt.Errorf("ErrorUsingAlloter")
)

// GoroutinePool is the base routine pool interface
// User can use custom goroutine pool by implementing this interface
type GoroutinePool interface {
	Submit(f func()) error
	Release()
}

// PooledAlloter is a alloter which has a worker pool
type PooledAlloter struct {
	timeout *time.Duration

	workerNum int
	pool      GoroutinePool

	initOnce sync.Once
}

// NewPooledAlloter creates an PooledAlloter instance
func NewPooledAlloter(workerNum int, opt ...*Options) *PooledAlloter {
	c := &PooledAlloter{
		workerNum: workerNum,
	}
	setOptions(c, opt...)
	return c
}

// WithPool will support for using custom goroutine pool
func (c *PooledAlloter) WithPool(pool GoroutinePool) *PooledAlloter {
	newAlloter := c.clone()
	newAlloter.pool = pool
	return newAlloter
}

// Exec is used to run tasks concurrently
func (c *PooledAlloter) Exec(tasks ...Task) error {
	return c.ExecWithContext(context.Background(), tasks...)
}

// ExecWithContext uses goroutine pool to run tasks concurrently
// Return nil when tasks are all completed successfully,
// or return error when some exception happen such as timeout
func (c *PooledAlloter) ExecWithContext(ctx context.Context, tasks ...Task) error {
	// Finaly, close pool.
	defer c.Release()
	// ensure the alloter can init correctly
	c.initOnce.Do(func() {
		c.initPooledAlloter()
	})

	if c.workerNum == -1 {
		return ErrorUsingAlloter
	}

	return execTasks(ctx, c, c.runWithPool, tasks...)
}

// GetTimeout return the timeout set before
func (c *PooledAlloter) GetTimeout() *time.Duration {
	return c.timeout
}

// Release will release the pool
func (c *PooledAlloter) Release() {
	if c.pool != nil {
		c.pool.Release()
	}
}

// initPooledAlloter init the pooled alloter once while the runtime
// If the workerNum is zero or negative,
// default worker num will be used
func (c *PooledAlloter) initPooledAlloter() {
	if c.pool != nil {
		// just pass
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
		fmt.Println("initPooledAlloter err")
	}
}

// runWithPool used the goroutine pool to execute the tasks
func (c *PooledAlloter) runWithPool(f func()) {
	err := c.pool.Submit(f)
	if err != nil {
		fmt.Println("submit task err:", err.Error())
	}
}

// setTimeout sets the timeout
func (c *PooledAlloter) setTimeout(timeout *time.Duration) {
	c.timeout = timeout
}

// clone will clone this PooledAlloter without goroutine pool
func (c *PooledAlloter) clone() *PooledAlloter {
	return &PooledAlloter{
		timeout:   c.timeout,
		workerNum: c.workerNum,
		initOnce:  sync.Once{},
	}
}
