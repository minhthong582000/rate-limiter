# Rate limiter

## Description

Simple rate limiter implementation using many different strategies:

- Fixed window
- Sliding window
- Token bucket (default)
- Leaky bucket

## Usage

```bash
go run main.go run [flags]
```

You can adjust the requests simulation by specifying the following flags:

- `--num-requests`: Number of requests to simulate.
- `--wait-time`: Time to wait between requests in milliseconds.
- `--jitter`: Jitter in milliseconds to add to the wait time. The actual wait time will be `wait-time + rand(-jitter, jitter)`.

## Comparison

### 1. Token bucket

Run:

```bash
go run main.go run --engine=token-bucket --capacity=5 --fill-duration=0.2 --num-requests=20 --wait-time=100
```

This strategy base on the idea of a bucket that can hold a limited number of tokens (`capacity`). A request comming in will consume a token from the bucket if there are sufficient tokens available. If the bucket is empty, the request will be rejected. The bucket will also be refilled at a constant `refillRate = 1/fill-duration` (token/s).

For example, you want to handle 300 requests every minute:

- capacity=300, fill-duration=0.2 (1 token every 0.2s)

In this configuration, we can handle on average 1 req for every 0.2s and allow users to burst up to 300 requests at once.

Key points:

- This strategy is effective in controlling bursty traffic but not sustained overload traffic.
- Pretty easy to implement and memory-efficient.

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

- capacity=300, window-size=60

In this configuration, we can handle on average 300 requests for every window of 60s. If the number of requests exceeds 300 in a window, the request will be rejected.

But if the requests is distributed near the end of the window and at the very beginning of the next window, it might be accepted twice the number of requests allowed (600).

Key points:

- Pretty simple to implement.
- Not accurate and allow twice the configured number of requests in the worst case.

### 4. Sliding window

TODO
