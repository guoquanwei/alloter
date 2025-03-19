package alloter

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

type BaseActuator interface {
	Exec(tasks []Task) error
	ExecWithContext(ctx context.Context, tasks []Task) error
}

type TimedAlloter interface {
	BaseActuator
}

type Task func() error

// wrapperTask will wrapper the task in order to notice execution result
// to the main process
func wrapperTask(ctx context.Context, task Task, wg *sync.WaitGroup, resChan *chan error, errChan *chan error) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("alloter panic:%v\n%s", r, string(debug.Stack()))
				*errChan <- err
			}

			wg.Done()
		}()
		select {
		case <-ctx.Done():
		case *resChan <- task():
			err := <-*resChan
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

func noBlockGo(ctx context.Context, errChan *chan error) (end bool, err error) {
	select {
	case <-ctx.Done():
		return true, ctx.Err()
	case err = <-*errChan:
		return true, err
	default:
	}
	return false, nil
}

func blockGo(ctx context.Context, errChan *chan error) (err error) {
	select {
	case <-ctx.Done():
		if ctx.Err() != nil {
			return ctx.Err()
		}
	case err = <-*errChan:
		return err
	}
	return nil
}
