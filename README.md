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

### init Alloter
```go
// simple concurrency
func NewAlloter() *Alloter

// concurrency control
func NewCtrlAlloter(workerNum int) *Alloter

// goroutines pool from ants/v2
func NewPooledAlloter(workerNum int) *Alloter

```

### exec Alloter
```go
type Task func() error

func (c *Alloter) Exec(tasks *[]Task) error

func (c *Alloter) ExecWithContext(ctx context.Context, tasks *[]Task) error

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
        p := alloter.NewCtrlAlloter(1)
        err = p.ExecWithContext(ctx.Request().Context(), &tasks)
        if err != nil {
            ctx.JSON(500, err.Error())
            return
        }
        ctx.JSON(200, users)
        return
    }

```