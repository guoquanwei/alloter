package alloter

import (
	"context"
	"fmt"
	"time"
)

// BaseAlloter is the alloter interface
type BaseAlloter interface {
	Exec(tasks *[]Task) error
	ExecWithContext(ctx context.Context, tasks *[]Task) error
}

// TimedAlloter is the alloter interface within timeout method
type TimedAlloter interface {
	BaseAlloter
	GetTimeout() *time.Duration
	setTimeout(timeout *time.Duration)
}

// ErrorTimeOut is the error when executes tasks timeout
var ErrorTimeOut = fmt.Errorf("TimeOut")

// Task Type
type Task func() error

// Alloter is the base struct
type Alloter struct {
	timeout *time.Duration
}

// NewAlloter creates an Alloter instance
func NewAlloter(opt *Options) *Alloter {
	c := &Alloter{}
	setOptions(c, opt)
	return c
}

// Exec is used to run tasks concurrently
func (c *Alloter) Exec(tasks *[]Task) error {
	return c.ExecWithContext(context.Background(), tasks)
}

// ExecWithContext is used to run tasks concurrently
// Return nil when tasks are all completed successfully,
// or return error when some exception happen such as timeout
func (c *Alloter) ExecWithContext(ctx context.Context, tasks *[]Task) error {
	return execTasks(ctx, c, simplyRun, tasks)
}

// GetTimeout return the timeout set before
func (c *Alloter) GetTimeout() *time.Duration {
	return c.timeout
}

// setTimeout sets the timeout
func (c *Alloter) setTimeout(timeout *time.Duration) {
	c.timeout = timeout
}
