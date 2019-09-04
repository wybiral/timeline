# timeline
Self hosted news aggregate for collecting real-time article updates using [stream-sources](https://github.com/wybiral/stream-sources). The main application itself is a server that renders a the timeline for browser visitors as and provides a WebSocket endpoint for accepting a stream of articles from stream-sources. The tool [timeline-push](https://github.com/wybiral/timeline/tree/master/cmd/timeline-push) provides and easy way to pipe stream-sources output into the WebSocket endpoint.

## Run main server
```go run main.go```

## Stream updates
(from stream-sources directory)

``` python3 main.py | timeline-push ```
