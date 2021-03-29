package alloter

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

// wrapperTask will wrapper the task in order to notice execution result
// to the main process
func wrapperTask(ctx context.Context, task Task,
	wg *sync.WaitGroup, resChan chan error) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("alloter panic:%v\n%s", r, string(debug.Stack()))
				resChan <- err
			}

			wg.Done()
		}()

		select {
		case <-ctx.Done():
			return // fast return
		case resChan <- task():
		}
	}
}

// setOptions set the options for alloter
func setOptions(c TimedAlloter, options *Options) {
	if options == nil {
		return
	}
	if !options.TimeOut.IsZero() {
		c.setTimeout(options.TimeOut)
	}
}
