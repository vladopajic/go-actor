# Best practices

To enhance the code quality of projects that heavily rely on the actor model and utilize the `go-actor` library, it's recommended to adhere following best practices.

## Forget about `sync` package

Projects that fully relay on actor model and `go-actor` library shouldn't use any synchronization primitives from `sync` package. Therefore repositories based on `go-actor` could add linter that will warn them if `sync` package is used, eg:

```yml
linters-settings:
  forbidigo:
      forbid:
        - 'sync.*(# sync package is forbidden)?'
```

While the general rule is to avoid `sync` package usage in actor-based code, there might be specific situations where its use becomes necessary. In such cases, the linter can be temporarily disabled to accommodate these exceptional needs.

## Use `goleak` library

[goleak](https://github.com/uber-go/goleak) is vary helpful library that detects goroutine leaks in unit tests. This can be helpful because it can identify actors that are still running after unit test have ended. Gracefully ending actors is important to test because it can identify various problems in implementation.

## Start select statements in `DoWork` with context

Workers should always respond to `Context.Done()` channel and return `actor.WorkerEnd` status in order to end it's actor. As a rule of thumb it's advised to always list this case first since it should be included in every `select` statement.

```go
func (w *worker) DoWork(ctx actor.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done():
		return actor.WorkerEnd
	case <-w.fooMbx.ReceiveC():
		handleFoo()
	case <-w.barMbx.ReceiveC():
		handleBar()
	}
}
```
 
---

`// page is not yet complete`