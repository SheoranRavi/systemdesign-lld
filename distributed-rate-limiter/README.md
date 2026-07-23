# Distributed Rate Limiter

A distributed token-bucket rate limiter intended to run in an API gateway.
Rules are stored in etcd and per-client token-bucket state is stored in Redis.
Redis Lua scripts update the bucket atomically across server instances.

## Features

- Client identification by IP address, API key, or user ID.
- Dynamically loaded rules from etcd.
- Atomic Redis Lua token-bucket implementation.
- Horizontally scalable server and test-client containers.
- Nginx load balancing across server replicas.

## Endpoint restrictions

The test setup enforces these client-key combinations:

| Endpoint | Required client key |
| --- | --- |
| `/login` | `IpAddressKey` |
| `/signup` | `IpAddressKey` |
| `/getgeo` | `ApiKey` |

Invalid combinations return `400 Bad Request`.

## Local build

Build the server executable into `bin/drl`:

```bash
make build
```

Remove the generated executable:

```bash
make clean
```

## Docker architecture

```text
test clients -> Nginx load balancer -> server replicas
                                      |-> Redis
                                      |-> etcd
```

The server uses these environment variables:

| Variable | Default | Description |
| --- | --- | --- |
| `ETCD_ADDR` | `localhost:2379` | etcd endpoint |
| `REDIS_ADDR` | `localhost:6379` | Redis endpoint |
| `PORT` | `8080` | HTTP listening port |

The test client uses:

| Variable | Default | Description |
| --- | --- | --- |
| `TARGET_URL` | `http://localhost:8080/checkRateLimit` | Rate-limiter URL |
| `TOTAL_REQUESTS` | `100000` | Requests per client instance |
| `CONCURRENCY` | `400` | Workers per client instance |
| `UNIQUE_CLIENTS` | `100` | Generated IP/API-key count |
| `CPU_PROFILE` | `cpu.prof` | CPU profile output path |

## Run with Docker Compose

Build the images:

```bash
docker compose build
```

Start a two-server, two-client test:

```bash
docker compose up -d \
  --scale server=2 \
  --scale client=2 \
  etcd redis server loadbalancer client
```

The load balancer is exposed at:

```text
http://localhost:8080/checkRateLimit
```

To change the number of instances:

```bash
docker compose up -d \
  --scale server=5 \
  --scale client=8 \
  server loadbalancer client
```

Always specify the desired `server` scale when starting clients or the load
balancer. Otherwise Compose may reconcile the dependency graph using the
default server scale and remove extra server containers.

Nginx resolves the Docker `server` service dynamically, so newly created or
removed server replicas are picked up without retaining stale container IPs.

## Performance optimizations

The test client and rate limiter include several throughput-oriented
optimizations:

- The client drains each response body before closing it:

  ```go
  io.Copy(io.Discard, resp.Body)
  resp.Body.Close()
  ```

  Draining the body allows Go's HTTP transport to reuse persistent TCP
  connections instead of repeatedly establishing new connections.

- The client uses one shared `http.Client` and a tuned `http.Transport` with
  connection pooling and high `MaxIdleConnsPerHost` settings.
- Request JSON payloads are pre-generated before the benchmark starts, keeping
  request generation and JSON encoding out of the measured request loop.
- The rate limiter uses an atomic Redis Lua script, so token refill and token
  consumption happen in one Redis operation without a read-modify-write race.
- Server replicas share Redis state, allowing any server instance to process a
  request for any client bucket.

Inspect containers and output:

```bash
docker compose ps -a
docker compose logs -f client
docker compose logs -f loadbalancer
docker stats
```

The client prints its summary only after all requests for that instance finish.

Stop containers while preserving Redis and etcd data:

```bash
docker compose down
```

Reset Redis and etcd data as well:

```bash
docker compose down -v
```

## Throughput measurement

Each client reports its own completed-request rate. For aggregate throughput,
use all clients' total request count divided by the wall-clock interval from
the earliest client start to the latest client finish:

```text
aggregate RPS = total responses from all clients / test duration
```

Track response categories separately:

```text
total RPS   = all HTTP responses / duration
allowed RPS = HTTP 200 / duration
rejected    = HTTP 429 / duration
failures    = HTTP 5xx or 502 / duration
```

`429` is a valid rate-limit decision. `502` indicates a load-balancer or
upstream availability problem and should be eliminated before comparing runs.

Run tests with a warm-up period, then repeat fixed-size or fixed-duration runs.
Sweep client concurrency and replica counts independently. Monitor CPU and
memory while the test is running, not after it finishes:

```bash
docker stats
```

Server replicas share one Redis instance. Since the token-bucket Lua script is
executed atomically by Redis, Redis can become the scaling bottleneck even when
additional server replicas are added. Nginx, the Docker host, and the client
containers can also become shared bottlenecks.

For comparable tests, record:

- server replica count
- client replica count
- client concurrency
- aggregate RPS
- HTTP 200, 429, 502, and 5xx rates
- latency percentiles, if available
- CPU and memory for clients, Nginx, servers, and Redis
