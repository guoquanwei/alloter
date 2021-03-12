## Introduction
[![GoDoc](https://godoc.org/github.com/ITcathyh/alloter?status.svg)](https://godoc.org/github.com/guoquanwei/alloter)

Alloter is a concurrent toolkit to help execute functions concurrently in an efficient and safe way.
* It supports concurrency limits.
* It supports specifying the overall timeout to avoid blocking.
* It supports the recovery of underlying goroutines.
* It supports the use of goroutines pool(invoke [ants/v2](https://github.com/panjf2000/ants)).
* It supports context passing; listen ctx.Done(), will return.
* It supports ending other tasks when an error occurs.
* It supports fast return when an error occurs.

## How to use
Generally it can be set as a singleton to save memory. There are some example to use it.
### Normal Alloter
Alloter is a base struct to execute functions concurrently.
```
    // option is not necessary
	opt := &Options{TimeOut:DurationPtr(time.Millisecond*50)}
	c := NewAlloter(opt)
	
	err := c.Exec(
		func() error {
			fmt.Println(1)
			time.Sleep(time.Second * 2)
			return nil
		},
		func() error {
			fmt.Println(2)
			return nil
		},
		func() error {
			time.Sleep(time.Second * 1)
			fmt.Println(3)
			return nil
		},
	)
	
	if err != nil {
		// ...do sth
	}
```
### Pooled Alloter
Pooled alloter uses the goroutine pool to execute functions. In some times it is a more efficient way.
```
    tasks := []alloter.Task{
		func() error {
			time.Sleep(1 * time.Second)
			fmt.Println(`1 end`)
			return nil
		},
		func() error {
			time.Sleep(1 * time.Second)
			fmt.Println(`2 end`)
			return nil
		},
	}
	
	
	// 'limit' needs >= 0,default is runtime.NumCPU()
	// 'option' is not necessary
	opt := &Options{TimeOut:DurationPtr(time.Millisecond*50)}
	c := NewPooledAlloter(1, opt)
	
	err := c.Exec(tasks...) 
	if err != nil {
		...do sth
	}
	// can also be used c.ExecWithContext()
	// err := c.ExecWithContext(context, tasks...) 

```
Use custom goroutine pool
```
	c := NewPooledAlloter(5).WithPool(pool)
```
### Simply exec using goroutine
```
	done := Exec(...)

	if !done {
		// ... do sth 
	}
```
