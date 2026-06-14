## Distributed Rate Limiter

### Functional Requirements
Identify clients by
- IP address
- API Key
- User Id

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