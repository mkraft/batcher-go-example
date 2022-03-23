Toying with the code for [a WebSocket event batching proxy idea](https://gist.github.com/mkraft/9a4cf09e898b65b9d21e8183a8215aa0).

```
go run .
go test -v
```

TODO:

- [ ] there are races (`go test -race` and `go run -race . -v`)
- [ ] more tests
- [ ] benchmark
- [ ] is there a better way of controlling shared access to the `queues map[string]chan *event` other than the `sync.Mutex`?
