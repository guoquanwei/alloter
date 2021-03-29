## Introduction

[![GoDoc](https://godoc.org/github.com/ITcathyh/alloter?status.svg)](https://godoc.org/github.com/guoquanwei/alloter)

Alloter is a goroutine's concurrent toolkit to help execute functions concurrently in an efficient and safe way.

It was inspired by a Node.js package's function, [bluebird](https://npmjs.com/package/bluebird) .map()

* It supports concurrency limits.
* It supports recovery goroutine's panic.
* It supports specifying the overall timeout to avoid blocking.
* It supports the use of goroutines pool(invoke [ants/v2](https://github.com/panjf2000/ants)).
* It supports context passing; listen ctx.Done(), will return.
* It supports ending other tasks when an error occurs.
* It supports fast return when an error occurs.

## How to use

Generally it can be set as a singleton to save memory. There are some example to use it.

### Pooled Alloter

Pooled alloter uses the goroutine pool to execute functions. In some times it is a more efficient way.

```go
   tasks := []alloter.Task{
	func() error {
    	    time.Sleep(1 * time.Second)
    	    return nil
	}, 
	func() error {
	    time.Sleep(1 * time.Second)
	    return nil
	},
   }
    // 'limit' needs >= 0,default is runtime.NumCPU()
    // 'option' is not necessary, can be use 'nil'
    c := NewPooledAlloter(1, &Options{TimeOut: time.Now().Add(8 * time.Second)})

    err := c.ExecWithContext(context.Background(), &tasks)
    if err != nil {
        ...do sth
    }
    // can also be used c.ExecWithContext()
    // err := c.ExecWithContext(context, &tasks) 
```

### Normal Alloter, like 'errgroup'.
Alloter is a base struct to execute functions concurrently.
```go
    // 'option' is not necessary, can be use 'nil'
    c := NewAlloter(&Options{TimeOut: time.Now().Add(3 * time.Second)})
    err := c.ExecWithContext(context.Background(), &[]alloter.Task{
        func() error {
            time.Sleep(time.Second * 2)
            fmt.Println(1)
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
    })
    if err != nil {
    	...do sth
    }

```
### Demo!!!
```go
    func (that *Controller) TestGoRunLock(ctx echo.Context) {
        userIds := []string{
            "uuid_1",
            "uuid_2",
        }
        tasks := []alloter.Task{}
        users := []third_parts.User{}
        mux := sync.Mutex{}
        for _, uid := range userIds {
            func (uid string) {
                tasks = append(tasks, func() error {
                user, resErr := third_parts.GetUserById(uid)
                if resErr != nil {
                    return resErr
                }
                mux.Lock()
                users = append(users, user)
                mux.Unlock()
                return nil
                })
            }(uid)
        }
        p := alloter.NewPooledAlloter(1, &alloter.Options{TimeOut: time.Now().Add(3 * time.Second)})
        err = p.ExecWithContext(ctx.Request().Context(), &tasks)
        if err != nil {
            ctx.JSON(500, err.Error())
            return
        }
        ctx.JSON(200, users)
        return
    }

```