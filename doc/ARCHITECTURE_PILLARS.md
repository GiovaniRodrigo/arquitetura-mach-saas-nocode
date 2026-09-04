# Mapping of the MACH Pillars & Core Components

[cite_start]The system architecture is entirely based on the acronym **MACH**: *Microservices, API-first, Cloud-native SaaS and Headless*[cite: 4].

## 1. 🧩 M – Microservices
[cite_start]The platform engine is divided into specialized services, independent and with isolated lifecycles[cite: 16]:

* [cite_start]**Design Engine (UI):** Responsible for the CRUD that stores the definitions and metadata of the interface created by the user, structured as a recursive tree[cite: 17].
* [cite_start]**Logic Engine (Rules):** Stores and interprets business rules based on logical nodes (decision trees)[cite: 18].
* [cite_start]**IAM Service (Identity & Access Management):** Centralizes access control, authentication, permission levels, and Tenant isolation[cite: 19].
* [cite_start]**Deploy Engine:** Controls the publication states of customer systems, managing versioning via status flags and orchestrating the provisioning of dynamic resources[cite: 20].
* [cite_start]**Export Engine:** Isolated service responsible for coordinating the collection of large volumes of data for asynchronous export of complete packages[cite: 21].

## 2. 🔌 A – API-first & Internal Communication
* [cite_start]**Contract-First Approach:** All internal communication is defined before the business code is developed[cite: 148].
* [cite_start]**gRPC Protocol:** Communication between internal microservices occurs via gRPC over HTTP/2 with Protocol Buffers, ensuring high performance and strong typing[cite: 25, 198, 199].

## 3. ☁️ C – Cloud-native SaaS & Cost Efficiency
* [cite_start]**Cloud Elasticity:** The system leverages on-demand scalability of the cloud infrastructure to mitigate bottlenecks[cite: 30, 154].
* [cite_start]**Shared Database Multi-tenancy:** To optimize RAM and CPU usage, multiple customers share the same database instance[cite: 31]. [cite_start]Isolation is logically guaranteed by the `tenant_id` column[cite: 32].

## 4. 👤 H – Headless & Rendering Engine (Headless Player)
[cite_start]The graphical interface viewed by the end user is completely decoupled from the back-end[cite: 35]:
* [cite_start]**Recursive Tree Structure:** The front-end acts as a universal intelligent renderer that consumes raw structural definitions (JSON) based on the *Composite* pattern (`componente_filhos` property)[cite: 36, 37].
* [cite_start]**Dynamic Navigation:** Switching between screens occurs under the SPA (Single Page Application) model through `redirect` actions associated with dynamic routes[cite: 38].
