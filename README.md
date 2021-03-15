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
    deadline := time.Now().Add(8 * time.Second)
    c := NewPooledAlloter(1, &Options{TimeOut: &deadline})

    err := c.Exec(&tasks)
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
    deadline := time.Now().Add(8 * time.Second)
    c := NewAlloter(&Options{TimeOut: &deadline})
    err := c.Exec(
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
	}, )
	
	if err != nil {
		...do sth
	}

```
### Demo!!!
```go
    func getRequestDeadLine(ctx *echo.Context) time.Time {
    // timeout middleware will fill Deadline().
	deadline, _ := (*ctx).Request().Context().Deadline()
	// default timeout is 8s.
	if deadline.IsZero() {
	    deadline = time.Now().Add(8 * time.Second)
	}
	return deadline
    }
    
    type UsersLock struct {
        sync.Mutex
        users []third_parts.User
    }
    
    func wrapFunc(tasks *[]alloter.Task, lockStore *UsersLock, uid string) {
        *tasks = append (*tasks, func() error {
            // request/db operations.
            user, err := third_parts.GetUserById(uid)
            if err != nil {
                return err
            }
            lockStore.Lock()
            lockStore.users = append(lockStore.users, user)
            lockStore.Unlock()
            return nil
        })
    }
    
    func (that *Controller) TestGoRunLock(ctx echo.Context) {
        userIds := []string{
            `uuid_1`,
            `uuid_2`,
        }
        tasks := []alloter.Task{}
        result := UsersLock{}
        for _, uid := range userIds {
            wrapFunc(&tasks, &result, uid)
        }
        poolDeadline := getRequestDeadLine(&ctx)
        p := alloter.NewPooledAlloter(1, &alloter.Options{TimeOut: &poolDeadline})
        err = p.ExecWithContext(ctx.Request().Context(), &tasks)
        if err != nil {
            ctx.JSON(500, err.Error())
            return
        }
        ctx.JSON(200, result.users)
        return
    }

```