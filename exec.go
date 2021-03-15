package alloter

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// wait waits for the notification of execution result
func wait(ctx context.Context, c TimedAlloter,
	resChan chan error, cancel context.CancelFunc) error {
	if timeout := c.GetTimeout(); timeout != nil {
		return waitWithTimeout(ctx, resChan, *timeout, cancel)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-resChan:
			if err != nil {
				cancel()
				return err
			}
		}
	}
}

// waitWithTimeout waits for the notification of execution result
// when the timeout is set
func waitWithTimeout(ctx context.Context, resChan chan error,
	timeout time.Duration, cancel context.CancelFunc) error {
	for {
		select {
		case <-time.After(timeout):
			cancel()
			return ErrorTimeOut
		case <-ctx.Done():
			return nil
		case err := <-resChan:
			if err != nil {
				cancel()
				return err
			}
		}
	}
}

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

	// support timeout.
	var timeoutErr error
	timeout := c.GetTimeout()
	if timeout != nil {
		go func() {
			time.Sleep(*timeout)
			timeoutErr = ErrorTimeOut
		}()
	}

	for _, task := range *tasks {
		// If find error, fast return.
		if timeoutErr != nil {
			return timeoutErr
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
		if timeoutErr != nil {
			return timeoutErr
		}
		select {
		case <-ctx.Done():
			return nil
		case err := <-resChan:
			if err != nil {
				cancel()
				return err
			}
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

// ExecWithError simply runs the tasks concurrently
// nil will be returned is all tasks complete successfully
// otherwise custom error will be returned
func ExecWithError(tasks *[]Task) error {
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(len(*tasks))

	for _, t := range *tasks {
		go func(task Task) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("alloter panic:%v\n%s\n", r, string(debug.Stack()))
				}

				wg.Done()
			}()

			if e := task(); e != nil {
				err = e
			}
		}(t)
	}

	wg.Wait()
	return err
}
