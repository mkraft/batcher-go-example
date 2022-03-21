Toying with the code for [a WebSocket event batching proxy idea](https://gist.github.com/mkraft/9a4cf09e898b65b9d21e8183a8215aa0).

```
go run .
```

TODO:

- [ ] test
- [ ] do I need to notify the backend client that the proxy is stopped or is the context interface enough?
- [ ] is there a better way of controlling shared access to the `queues map[string]chan *event` other than the `sync.Mutex`?
- [ ] should the `(*proxy).Publish` method continue to function once the proxy has been cancelled (bypassing the handlers)?