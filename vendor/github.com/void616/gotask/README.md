# gotask

Simple routine group library for Golang. Main purpose is to gracefully shutdown complex multitask apps.

```sh
go get github.com/void616/gotask
```

# Examples

## Group

Make a group, add tasks, run it and wait for completion.

```go
t1, _ := gotask.NewTask("MyTask #1", func() { log.Println("Routine 1") })
t2, _ := gotask.NewTask("MyTask #2", func() { log.Println("Routine 2") })

g := gotask.NewGroup("MyGroup")
g.Add(t1)
g.Add(t2)

g.Run()
g.Wait()
```

Output:
```
Routine 2
Routine 1
```

## Task lifetime

Pass token into the routine to control it's lifetime.

```go
result := 0
routine := func(token *gotask.Token) {
  for {
    result++
    // sleep with alertness
    if token.Sleep(time.Minute * 60) {
      break
    }
  }
}

t, _ := gotask.NewTask("MyTask", routine)
token, waiter, _ := t.Run()

token.Stop()
waiter.Wait()

log.Println("Result", result)
```

Output:
```
Result 1
```

## Panic

Stop group of tasks on panic.

```go
g := gotask.NewGroup("MyGroup")

routine1 := func(token *gotask.Token) {
  log.Println("Routine 1 begin")
  defer log.Println("Routine 1 end")
  token.Sleep(time.Minute * 60)
}

routine2 := func(token *gotask.Token) {
  log.Println("Routine 2 begin")
  defer log.Println("Routine 2 end")
  
  panic("panic!")
}

t1, _ := gotask.NewTask("MyTask #1", routine1)
t2, _ := gotask.NewTask("MyTask #2", routine2)

t2.Recover(func(v interface{}) {
  g.Stop()
})

g.Add(t1)
g.Add(t2)

g.Run()
g.Wait()
```

Output:
```
Routine 1 begin
Routine 2 begin
Routine 2 end
Routine 1 end
```