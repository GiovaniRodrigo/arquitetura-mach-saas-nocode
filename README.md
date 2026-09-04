# MACH V4 System Builder — Low-Code / No-Code Platform

## 1. Overview

This repository contains the unified technical architecture specification for the **MACH V4** platform. The system was designed under modern distributed architecture principles to allow users to build their own digital applications through an entirely visual interface.

Version V4 consolidates gRPC contracts via Protocol Buffers, dynamic payload mapping using key-value maps, fine-grained component-level access control (IAM), and integrates scalable asynchronous messaging layers with real-time synchronization.

## 2. Core Platform Requirements

* **UI CRUD:** Creation and saving of metadata for building visual user interfaces.
* **Business Rules CRUD:** Registration of operational rules, access control, and functional requirements.
* **Instant Publishing:** Immediate deployment mechanism based on an interpreted approach.
* **Hierarchical Multi-Tenancy:** Structured logical isolation for Owners, Partners, and End Customers.
* **Asynchronous Export:** Complete background extraction of operational data and metadata.
* **Real-Time Collaboration:** Simultaneous multi-user editing within the builder canvas.

## 3. Documentation Reading Guide

To gain a detailed understanding of each infrastructure component, refer to the specialized files:

* [`ARCHITECTURE_PILLARS.md`](doc/ARCHITECTURE_PILLARS.md): Mapping of the 4 MACH pillars and microservice breakdown.
* [`GATEWAY_COLLABORATION.md`](doc/GATEWAY_COLLABORATION.md): Details on the hybrid Gateway (Go & Elixir) and WebSocket synchronization.
* [`DATA_SECURITY.md`](doc/DATA_SECURITY.md): Dynamic data management, Blind Index, and component-level IAM policies.
* [`ASYNC_OBSERVABILITY.md`](doc/ASYNC_OBSERVABILITY.md): Distributed messaging with KEDA and OpenTelemetry tracing.
* [`CONTRACTS_PERFORMANCE.md`](doc/CONTRACTS_PERFORMANCE.md): Official `.proto` contracts, rendering strategies, and technical evolution roadmap.
