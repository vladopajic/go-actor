# Best practices

To enhance the code quality of projects that heavily rely on the actor model and utilize the `go-actor` library, it's recommended to adhere following best practices.

## Forget about `sync` package

Projects that fully relay on actor model and `go-actor` library shouldn't use any synchronization primitives from `sync` package. Therefore repositories based on `go-actor` could add linter that will warn them if `sync` package is used, eg:

```
linters-settings:
  forbidigo:
      forbid:
        - 'sync.*(# sync package is forbidden)?'
```

While the general rule is to avoid `sync` package usage in actor-based code, there might be specific situations where its use becomes necessary. In such cases, the linter can be temporarily disabled to accommodate these exceptional needs.


---

`// page is not yet complete`