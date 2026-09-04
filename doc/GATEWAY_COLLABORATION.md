# Hybrid Gateway Layer and Real-Time Collaboration

[cite_start]To maximize network efficiency, the BFF (Backend For Frontend) pattern was split into two distinct technologies by protocol type and use case[cite: 23].

## 1. API Gateway in Go (Golang)
[cite_start]Responsible for managing the entire traditional synchronous *Request-Response* flow (HTTP/REST)[cite: 24].
* [cite_start]**Responsibilities:** Acts as an ultra-fast, low-latency proxy[cite: 25].
* [cite_start]**Security and Traffic:** Validates the JWT token in the `Authorization` header and applies *Rate Limiting* policies[cite: 25].
* [cite_start]**Protocol Translation:** Translates HTTP requests received from the browser into internal gRPC calls directed at the microservices[cite: 25].

## 2. Elixir Collaboration Engine (Phoenix Channels)
[cite_start]Responsible strictly for real-time, bidirectional persistent connections via WebSockets[cite: 26, 77].

### In-Memory State Management (BEAM)
* [cite_start]**Isolated Concurrency:** For each screen under active edit in the builder panel, the Erlang VM (BEAM) instantiates an isolated lightweight process (`GenServer`), keeping the component tree alive in memory[cite: 27, 78].
* [cite_start]**Replication and Synchronization:** The lightweight mutations sent by a user are processed by the `GenServer`, replicated into safety *snapshots* on a global Redis instance, and instantly propagated via *broadcast* to the other co-creators[cite: 79].

### Debounce-Based Persistence Strategy (Write-Behind)
[cite_start]To shield the relational database from write exhaustion induced by micro-movements in the UI, the following flow is adopted[cite: 81]:
1. [cite_start]Changes accumulate temporarily only in the BEAM process's memory and in Redis[cite: 82].
2. [cite_start]When the `GenServer` detects a period of network inactivity (e.g., 5 seconds of silence), it consolidates the recursive tree into a single stable payload[cite: 83].
3. [cite_start]The Elixir process fires off a **single optimized batch gRPC call** to the *Design Engine*, persisting the data into the relational database's `JSONB` column[cite: 84].

### Presence and Conflict Control
* [cite_start]**Phoenix Presence:** Tracking of online users and rendering of simultaneous cursors operates in a decentralized manner via CRDTs (*Conflict-Free Replicated Data Types*)[cite: 86].
* [cite_start]**Optimistic Locking:** Direct editing conflicts on the same element are prevented through temporary locks by *Blind Index*, dynamically disabling the inputs on competing browsers[cite: 87].
