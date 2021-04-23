package alloter

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

var ErrorTimeOut = fmt.Errorf("TimeOut")

type BaseActuator interface {
	Exec(tasks *[]Task) error
	ExecWithContext(ctx context.Context, tasks *[]Task) error
}

type TimedAlloter interface {
	BaseActuator
	GetTimeout() time.Time
	setTimeout(timeout time.Time)
}

// Options use to init alloter
type Options struct {
	TimeOut time.Time
}

type Task func() error


// wrapperTask will wrapper the task in order to notice execution result
// to the main process
func wrapperTask(ctx context.Context, cancel context.CancelFunc, task Task,
	wg *sync.WaitGroup, resChan *chan error, errChan *chan error, timeout time.Time) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("alloter panic:%v\n%s", r, string(debug.Stack()))
				*errChan <- err
			}

			wg.Done()
		}()

		select {
		case <-time.After(timeout.Sub(time.Now())):
			*errChan <- ErrorTimeOut
		case <-ctx.Done():
			cancel()
		case *resChan <- task():
			err := <- *resChan
			if err != nil {
				*errChan <- err
			}
		}
	}
}

func wrapperSimpleTask(task Task, wg *sync.WaitGroup, resChan *chan error) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("alloter panic:%v\n%s", r, string(debug.Stack()))
				*resChan <- err
			}

			wg.Done()
		}()

		err := task()
		if err != nil {
			*resChan <- err
		}
	}
}
