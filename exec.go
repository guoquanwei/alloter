package alloter

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// execTasks uses customized function to
// execute every task, such as using the simplyRun
func execTasks(parent context.Context, c TimedAlloter,
	execFunc func(f func()), tasks *[]Task) error {
	size := len(*tasks)
	if size == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(parent)
	resChan := make(chan error, size)
	wg := &sync.WaitGroup{}
	wg.Add(size)

	timeout := c.GetTimeout()
	for _, task := range *tasks {
		// If find error, fast return.
		if timeout != nil && time.Now().After(*timeout) {
			return ErrorTimeOut
		}
		select {
		case <-ctx.Done():
			return nil
		case err := <-resChan:
			if err != nil {
				cancel()
				return err
			}
		default:
		}

		child, _ := context.WithCancel(ctx)
		f := wrapperTask(child, task, wg, resChan)
		execFunc(f)
	}

	// Just support normal close resChan.
	// When error, wo can't close resChan, maybe some goroutines just finished.
	// So, when error, wo just can wait auto GC.
	go func() {
		wg.Wait()
		cancel()
		close(resChan)
	}()

	for {
		if timeout != nil && time.Now().After(*timeout) {
			return ErrorTimeOut
		}
		select {
		case <-ctx.Done():
			return nil
		case err := <-resChan:
			if err != nil {
				cancel()
				return err
			}
		default:
		}
	}
}

// simplyRun uses a new goroutine to run the function
func simplyRun(f func()) {
	go f()
}

// Exec simply runs the tasks concurrently
// True will be returned is all tasks complete successfully
// otherwise false will be returned
func Exec(tasks *[]Task) bool {
	var c int32
	wg := &sync.WaitGroup{}
	wg.Add(len(*tasks))

	for _, t := range *tasks {
		go func(task Task) {
			defer func() {
				if r := recover(); r != nil {
					atomic.StoreInt32(&c, 1)
					fmt.Printf("alloter panic:%v\n%s\n", r, string(debug.Stack()))
				}

				wg.Done()
			}()

			if err := task(); err != nil {
				atomic.StoreInt32(&c, 1)
			}
		}(t)
	}

	wg.Wait()
	return c == 0
}
