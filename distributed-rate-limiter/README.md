## Distributed Rate Limiter
Let's say that this RL will run inside an API gateway of some sort.
So, the API gateway will pass the request through the rate limiter. The API gateway will pass the request object to the rate limiter.
The rate limiter will use the config to parse the rules.

### Functional Requirements
- Identify clients by
    - IP address
    - API Key
    - User Id
- The rate limits should be configurable.

#### Non-functional requirements
- Latency (< 10 ms)
- Availability > Consistency

### Algorithms
- Fixed window counter
  - Con: Burstiness at the edges
- Sliding window log
  - Log timestamp of each request for each user. When request comes in, discard all timestamps outside of the window then check the count.
  - Use a MinHeap to store timestamps
  - Con: Storage
- Sliding window counter
  - Approximate algorithm.
  - Two counters per client. Previous window count and current window count
  - If 70% through current window, take 30% of the previous window count and add current count. If this is greater than limit, then deny the request
- Token Bucket
  - Store for each client the last timestamp and the current token count

## Configuration
How to load the configuration?
- For per API configuration or per userId configuration, rely on Redis. The API Key Gen service can write the RPS per API key to Redis.
- For IP-based configuration, either rely on Redis or use something like etcd for dynamic configuration management.
But, it is probably better to manage all configuration one way only. So, let's use something like etcd for API, userId, and IP-address related configuration.
When any of these change, the rate limiter can either update it's internal config via pull.

What happens when the configuration needs to be changed?
How to introduce a new rule?

