## Introduction
[![GoDoc](https://godoc.org/github.com/ITcathyh/alloter?status.svg)](https://godoc.org/github.com/guoquanwei/go-alloter)

Alloter is a concurrent toolkit to help execute functions concurrently in an efficient and safe way. 
It supports specifying the overall timeout to avoid blocking.

## How to use
Generally it can be set as a singleton to save memory. There are some example to use it.
### Normal Alloter
Alloter is a base struct to execute functions concurrently.
```
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
	opt := &Options{TimeOut:DurationPtr(time.Millisecond*50)}
	c := NewPooledAlloter(5, opt)
	
	err := c.Exec(...)
	
	if err != nil {
		// ...do sth
	}
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
