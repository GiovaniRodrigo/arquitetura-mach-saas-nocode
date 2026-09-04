# Data Management, Security, and IAM Isolation

[cite_start]Data security and integrity on the platform are enforced across multiple layers to support a secure multi-tenant ecosystem[cite: 65].

## 1. Dynamic Data Management (JSONB)
[cite_start]Since users create their own fields and tables, the database uses `JSONB`-type columns integrated into the same physical table[cite: 32].

## 2. Expanded Mapping via Blind Index
[cite_start]To perform validations and indexing without compromising privacy or exposing raw data[cite: 33]:
* [cite_start]Dynamic field definitions use a model based on the components' **Blind Index** (cryptographic blind index) as the key[cite: 33].
* [cite_start]This structure defines type, requiredness, and limits (such as minimum value or maximum string length), ensuring safe and anonymous validations[cite: 33, 319].

## 3. Security and Identity Propagation (IAM)
* [cite_start]**JWT Tokens and gRPC Metadata:** The JWT token intercepted by the Go API Gateway is extracted and natively injected as **binary gRPC Metadata**[cite: 66, 67]. [cite_start]This securely propagates the identity context and `tenant_id` to all internal microservices[cite: 67].
* [cite_start]**Server-Side Condition Evaluation:** To prevent visual fraud that bypasses client-side code, the *IAM Service* processes all dynamic conditions strictly on the back-end[cite: 68]. [cite_start]The server evaluates the complex rules and returns to the front-end only a simple boolean map indexed by the *Blind Index* of the affected components[cite: 69].

### Example Permissions Payload Sent to the Headless Player
```json
{  
  "permissions": {    
    "8f3b2a1...": { "view": true, "click": false },    
    "4a9e2d3...": { "view": false, "click": false }  
  }
}
```

## 4. Data Isolation at the Database Layer
[cite_start]The database microservice applies automatic filtering clauses (`WHERE tenant_id = :id`) based on the context extracted from gRPC, natively preventing data leakage between customers[cite: 69].
