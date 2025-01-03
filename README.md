# Rate limiter

## Description

Simple rate limiter implementation using many different strategies:

- Fixed window
- Sliding window log
- Sliding window counter
- Token bucket
- Leaky bucket

### Dependencies

- [go 1.23 or above]([https://golang.org/](https://go.dev/doc/install))

## Usage

```bash
go run main.go run [flags]
```

You can adjust the requests simulation by specifying the following flags:

- `--num-requests`: Number of requests to simulate.
- `--wait-time`: Time to wait between requests in milliseconds.
- `--jitter`: Jitter in milliseconds to add to the wait time. The actual wait time will be `wait-time + rand(-jitter, jitter)`.
- `--parallel`: Number of parallel workers to simulate requests. Each worker will simulate `num-requests` requests.

## Comparison

### 1. Token bucket

Run:

```bash
go run main.go run --engine=token-bucket --capacity=5 --fill-duration=200 --num-requests=20 --wait-time=10
```

This strategy base on the idea of a bucket that can hold a limited number of tokens (`capacity`). A request comming in will consume a token from the bucket if there are sufficient tokens available. If the bucket is empty, the request will be rejected. The bucket will also be refilled at a constant `refillRate = 1/fill-duration` (token/ms).

For example, you want to handle 300 requests every minute:

- capacity=300, fill-duration=200ms (1 token every 0.2s)

In this configuration, we can handle on average 1 req for every 0.2s and allow users to burst up to 300 requests at once.

Key points:

- This strategy is effective in controlling bursty traffic but not sustained overload traffic.
- Pretty easy to implement and memory-efficient. But requires locking or atomic operations to update the token counter in high concurrency scenarios.

### 2. Leaky bucket

Run:

```bash
go run main.go run --engine=leaky-bucket --capacity=5 --drain-duration=200 --num-requests=20 --wait-time=100
```

Think of this strategy as a bucket with a hole at the bottom. The bucket can hold a limited number of requests (`capacity`). When a request comes in, it will be added to the bucket (`queue`). If the bucket is full, the request will be rejected. The bucket will also be drained at a constant `drainRate = 1000/drain-duration` (requests/s).

For example, you want to handle 300 requests every minute:

- capacity=300, drain-duration=200ms (process 1 req every 200ms)

In this configuration, first 300 requests will be enqueued and processed at a rate of 1 req every 0.2s. Other requests will be rejected after the bucket is full.

Key points:

- This strategy is effective if you want a smooth and steady rate of requests. Doesn't allow bursty traffic.
- Requires a queue to store incoming requests. Queue might be full with old requests and cause starvation for more recent requests.

### 3. Fixed window

Run:

```bash
go run main.go run --engine=fixed-window --capacity=5 --window-size=1000 --num-requests=20 --wait-time=100
```

This strategy devide time into fixed windows of duration `windowSize`. Each window has a limited number of requests (`capacity`). When a request comes in, it will be added to the current window. If the request counter exceeds the capacity, the request will be rejected. The window will be reset after `windowSize` time.

For example, you want to handle 300 requests every minute:

- capacity=300, window-size=60000 (ms)

In this configuration, we can handle on average 300 requests for every window of 60s. If the number of requests exceeds 300 in a window, the request will be rejected.

But if the requests is distributed near the end of the window and at the very beginning of the next window, it might be accepted twice the number of requests allowed (600).

Key points:

- Pretty simple to implement. But will envolve locking or atomic operations to update the request counter in high concurrency scenarios.
- Not accurate and allow twice the configured number of requests in the worst case.

### 4. Sliding window Log

Run:

```bash
go run main.go run --engine=sliding-window-log --capacity=5 --window-size=1000 --num-requests=20 --wait-time=100
```

This strategy is an improvement over the fixed window strategy. It stores the timestamp of each request in a log (requestLog), this log can be implemented using a queue, Redis sorted set, etc. When a request comes in, it will discard all requests that are outside the current window (`now - requestTS > window-size`). After that, it will count the number of requests in the remaining log. If the number of requests exceeds the capacity, the request will be rejected.

For example, you want to handle 300 requests every minute:

- capacity=300, window-size=60000 (ms)

In this configuration, we can handle on average 300 requests for every window of 60s. If the number of requests exceeds 300 in a window, the request will be rejected. This strategy will not suffer from the same issue as the fixed window strategy.

Key points:

- Very precise. Does not suffer from boundary issues.
- Not memory-efficient. Requires storing all requests in the log.
- CPU-intensive. Requires scanning the log for each new request to filter out old requests and count the number of requests in the current window.
- 2 issues above lead to scalability problems when the number of requests and the window size increase.

### 5. Sliding window Counter

Run:

```bash
go run main.go run --engine=sliding-window-counter --capacity=5 --window-size=1000 --num-requests=30 --wait-time=100
```

To reduce the memory and CPU overhead of the sliding window log strategy, we can use a sliding window counter. By using only 2 counters of `previous` and `current` windows, we can estimate total requests in the actual window (`now - window-size`) by taking the weighted count of `previous` and add it to the count of `current` window.

For example, `previous` window is starting at `0 -> 100` with `30` requests, `current` window is starting at `100 -> 200` and currently has `10` requests, and new request comes in at `160`. The approximate number of requests in the actual window is:

```go
currWindowWeight = (160 - 100)/100 = 0.6
prevWindowWeight = 1 - currWindowWeight = 1 - 0.6 = 0.4
estimatedCount = 30 * 0.4 + 10 = 22
```

About the configuration, let's say you want to handle 300 requests every minute:

- capacity=300, window-size=60000 (ms)

In this configuration, we can handle on average 300 requests for every (sliding) window of 60s. If the estimated number of requests exceeds 300 in a window, the request will be rejected.

The formula use in this strategy assumes that the number of requests is uniformly distributed in all windows which is why it's an approximation. But in reality, Cloudflare has been using this strategy in their rate limiter and shown [good results](https://blog.cloudflare.com/counting-things-a-lot-of-different-things/).

Key points:

- Trade-off between the accuracy of the rate limiter and memory/CPU overhead. But still more accurate than the fixed window strategy and does not suffer from boundary issues.
- Need locking or atomic operations to update the counters in high concurrency scenarios.

## Conclusion

Choosing the right rate-limiting strategy depends on a combination of your system’s requirements and constraints. Below are some factors to consider:

- Traffic Pattern: The pattern of incoming requests to your system.
  - Steady Traffic: Leaky Bucket ensures a consistent, predictable flow of requests.
  - Bursty Traffic: Token Bucket can efficiently handle sudden burst while maintaining average limits.
  - Mixed Traffic: Sliding Window Log/Counter offers a good enoug to handle both steady and bursty traffic.

- Accuracy:
  - High Precision: Sliding Window Log and Leaky Bucket provide precise control at the cost of higher memory or computational overhead.
  - Approximate Control: Sliding Window Counter and Token Bucket offer good enough accuracy for most use cases with better resource utilization.

- Resource Efficiency (Memory/CPU Overhead): The system resources available for rate-limiting operations.
  - High: Sliding Window Log and Leaky Bucket can be resource-intensive, especially with large window and request volumes.
  - Low: Token Bucket and Sliding Window Counter only require a few counters to track requests but may not be as precise.

- Scalability: The ability of the rate limiter to scale with increasing traffic, while ensuring fault tolerance.
  - Highly Scalable: Token Bucket and Sliding Window Counter are easier to scale by integrating with distributed caching system like Redis, Memcached... with lower resource overhead.
  - Limited Scalability: Sliding Window Log and Leaky Bucket may face bottlenecks under heavy traffic as they require way more memory.

Additionally, your implementation of the rate-limiter also contributes to the overall performance. Factors such as programming language, concurrency mechanisms (mutex locks, atomic operations (CAS), transaction...) can affect the efficiency of the rate limiter.

## Milestones

- [x] Implement rate limiter using different strategies:
  - [x] Fixed window
  - [x] Sliding window log
  - [x] Sliding window counter
  - [x] Token bucket
  - [x] Leaky bucket
- [x] Implement request simulator
- [ ] Implement a simple HTTP and expose its metrics

## References

- <https://blog.cloudflare.com/counting-things-a-lot-of-different-things>
- <https://konghq.com/blog/engineering/how-to-design-a-scalable-rate-limiting-algorithm>
- <https://raphaeldelio.com/2024/12/23/rate-limiting-with-redis-an-essential-guide/>
- <https://github.com/uber-go/ratelimit>
- <https://blog.algomaster.io/p/rate-limiting-algorithms-explained-with-code>
- <https://www.figma.com/blog/an-alternative-approach-to-rate-limiting/>
