# Rate Limit

[![Go Report Card](https://goreportcard.com/badge/github.com/esnunes/ratelimit)](https://goreportcard.com/report/github.com/esnunes/ratelimit)

## Library

## Reverse Proxy

```bash
go get github.com/esnunes/ratelimit/cmd/rl-proxy

rl-proxy -h
# Rate Limit Proxy
#
# Usage:
#   rl-proxy [flags]
#
# Flags:
#   -a, --addr string   bind address ip:port (default ":8080")
#   -b, --burst int     maximum burst size of requests (default 1)
#   -h, --help          help for rl-proxy
#   -q, --queue int     maximum queued requests
#   -r, --rate float    requests per second (default 2)
```

- Leaky Bucket algorithm;
- Rate Limit based on target endpoint (considering connection latency);
- Queue of connections when bucket limit reached;

## TODO

- [ ] Store and retrieve context values using functions;
- [ ] Calculate rate limit value based on specific date (e.g. service startup time);
- [ ] Cover with tests;
- [ ] Create docs about library usage;
- [ ] Create docs about proxy usage;
- [ ] Add Prometheus metrics support;
  - [ ] Number of times a response status code occurred;
  - [ ] Number of no more slots error;
  - [ ] Number of requests in queue;
  - [ ] Number of requests running;
  - [ ] Histogram of requests response time;
  - [ ] General Go statistics;

## License

ratelimit is licensed under the MIT license. Check the [LICENSE](LICENSE) file for details.

## Author

[Eduardo Nunes](http://nunes.io)
