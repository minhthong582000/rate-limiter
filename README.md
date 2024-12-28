# Rate limiter

## Description

Simple rate limiter implementation using many different strategies:

- Fixed window
- Sliding window
- Token bucket
- Leaky bucket

## Usage

```bash
go run main.go run -l traffic.txt --engine <engine>
```

## Comparison

### 1. Token bucket

Run:

```bash
go run main.go run -l traffic.txt --engine token-bucket
```

This strategy base on the idea of a bucket that can hold a limited number of tokens (`capacity`). A request comming in will consume a token from the bucket if there are sufficient tokens available. If the bucket is empty, the request will be rejected. The bucket will also be refilled at a constant `refillRate` (token/s).

For example, you want to handle 300 requests every minute:

- capacity=300, refillRate=5

In this configuration, we can handle on average 1 req for every 0.25s and allow users to burst up to 300 requests at once.

Key points:

- This strategy is effective in controlling bursty traffic but not sustained overload traffic.
- Pretty easy to implement and memory-efficient.

### 2. Leaky bucket

Run:

```bash
go run main.go run -l traffic.txt --engine leaky-bucket
```

Note that arrival time in the `traffic.txt` file is not encountered in this strategy. It only considers the number of requests received.

Think of this strategy as a bucket with a hole at the bottom. The bucket can hold a limited number of requests (`capacity`). When a request comes in, it will be added to the bucket (`queue`). If the bucket is full, the request will be rejected. The bucket will also be drained at a constant `drainRate`.

For example, you want to handle 300 requests every minute:

- capacity=300, drainRate=0.2 (1s/5req)

In this configuration, first 300 requests will be enqueued and processed at a rate of 1 req every 0.2s. Other requests will be rejected after the bucket is full.

Key points:

- This strategy is effective if you want a smooth and steady rate of requests.
- Requires a queue to store incoming requests. Queue might be full with old requests and cause starvation for more recent requests.

### 3. Fixed window

TODO

### 4. Sliding window

TODO
