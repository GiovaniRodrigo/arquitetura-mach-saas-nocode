# Asynchronous Messaging (KEDA) and Observability (OpenTelemetry)

## 1. Asynchronous Event Layer and Dynamic Autoscaling (KEDA)
To ensure that heavy background tasks (notifications, webhooks, and third-party integrations configured by the user) do not block the system's synchronous threads, an event-driven asynchronous architecture is adopted via **RabbitMQ**, orchestrated by **KEDA (Kubernetes Event-driven Autoscaling)**.

### Messaging Topology and Flow
1. When a business rule processed by the Logic Engine determines that a background task should be triggered, the service publishes a JSON-structured event containing the `tenant_id` and the corresponding `component_blind_index` to a RabbitMQ *Exchange*.
2. RabbitMQ performs dynamic routing to the dedicated queues (e.g., `webhooks.disparo`, `notificacoes.envio`), immediately freeing up the gRPC flow.

### Elastic Scale-to-Zero Scalability
The *workers* responsible for consuming the messages are controlled by KEDA in the Kubernetes cluster using a custom `ScaledObject` resource:
* **Scaling Metric:** KEDA monitors queue size (`QueueLength`) directly via the RabbitMQ API, avoiding the inefficient reactivity of CPU/Memory metrics.
* **Scale-to-Zero:** If the queue is completely empty (e.g., during low-activity overnight periods), KEDA reduces the number of active Pods to **0**, eliminating idle compute costs.
* **Behavior Under Load:** Upon detecting a message backlog, KEDA scales out aggressively up to the stipulated limit (e.g., `maxReplicaCount: 50`), distributing the load fairly.

### Isolation Against the "Noisy Neighbor"
To mitigate the risk of a single customer flooding the platform with millions of webhooks and monopolizing the *workers*, the Deploy Engine provisions dynamic routing keys and RabbitMQ enforces *Fair Queuing* policies. Repeated third-party integration failures are immediately diverted to error queues (*Dead Letter Queues* - DLQ), triggering alerts on the respective *tenant*'s dashboard without impacting other users.

## 2. Observability and Distributed Tracing (OpenTelemetry)
End-to-end traceability across hybrid network flows (HTTP, gRPC, and AMQP/RabbitMQ) is ensured using the **OpenTelemetry** standard, with storage and visualization via **Jaeger**.

### Cross-Transport Context Propagation
To unify the journey of a distributed request, the system injects and extracts the global trace identifier (**Trace ID**) following the *W3C Trace Context* specification via the `traceparent` header:
1. **Origin (API Gateway):** The Go middleware generates the root Trace ID upon receiving the HTTP request.
2. **Synchronous Transit (gRPC):** The OpenTelemetry gRPC interceptor natively injects the Trace ID as binary Metadata in the call to the Logic Engine.
3. **Asynchronous Transit (AMQP):** Before publishing the background message to RabbitMQ, the Logic Engine writes the current trace context into the AMQP message's *Headers* map.
4. **Destination (Workers):** The executing worker extracts the `traceparent` from the message headers, opens a time sub-block (*Span*), and attaches the execution logs or errors directly to the transaction's unified history.

### Secure Multi-Tenant Context Tags
Each monitoring *Span* is enriched with anonymized structural attributes to enable internal auditing and agile debugging by the engineering team:
* `platform.tenant_id`: Allows instant filtering of all infrastructure *traces* associated with a specific customer.
* `platform.component.blind_index`: Identifies the exact component generating the bottleneck without ever exposing sensitive business data or metadata in Jaeger.
