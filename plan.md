
You are a senior backend architect and staff-level Go engineer with deep experience building production-grade fintech systems and core banking platforms. I need you to help me design and build a complete, production-oriented backend for a neobank (similar to Monobank, Monzo, or Revolut).

**Existing Foundation:**
I already have a production-ready double-entry ledger service:  
https://github.com/iho/goledger

This ledger is built in Go and includes:
- Proper double-entry accounting
- PostgreSQL as the primary database (using sqlc + pgx)
- Redis-backed idempotency
- Concurrency safety with deadlock prevention (sorted locking)
- Clean architecture (domain + use cases)
- Observability (Prometheus metrics + structured logging)
- CLI tool and Docker support
- Support for accounts, transfers, and holds

**Task:**
Design and plan a full neobank backend system, using the existing `goledger` as the core **Ledger Service**. Do not rewrite the ledger from scratch — extend or integrate with it where needed.

**High-level Goals:**
- Build a scalable, reliable, and secure neobank backend
- Focus on correctness of financial operations (never lose or double money)
- Mobile-first experience
- Event-driven architecture
- Production-grade practices from day one

**Mandatory Architectural Principles:**
- Event-driven architecture with Kafka (or NATS) as the central event bus
- Domain-Driven Design and Clean Architecture
- Prefer Go for performance-critical and high-concurrency services (Ledger, Payments, Notifications, etc.)
- Strong consistency for all monetary operations
- Idempotency Key pattern on every mutating operation
- Saga pattern for complex multi-step business processes
- Outbox pattern for reliable event publishing
- Observability-first (OpenTelemetry tracing + metrics + structured logs)
- Security and auditability by design

**Core Services for MVP (must include):**

1. **User Service** — registration, profile management, basic KYC flow
2. **Ledger Service** — based on the existing `goledger` repository (extend if necessary)
3. **Payment / Transfer Service** — internal P2P transfers and basic external payments
4. **Card Service** — virtual card issuing and management (with mock or placeholder integration to a card processor)
5. **Fraud & Risk Service** — basic transaction monitoring and risk scoring (can start with rules + placeholder for ML)
6. **Notification Service** — push notifications, email, and SMS
7. **API Gateway / BFF** — single entry point for the mobile application

**Technical Stack Preferences:**
- Language: Go 1.23+
- Database: PostgreSQL (primary) + Redis
- Messaging: Kafka (preferred) or NATS
- API: REST (chi or similar) + gRPC between internal services
- Observability: OpenTelemetry + Prometheus + Grafana
- Deployment: Docker + docker-compose for local development, Kubernetes-ready
- Testing: Table-driven tests, integration tests, contract tests

**Non-Functional Requirements:**
- High reliability and correctness for financial operations
- Horizontal scalability
- Idempotency and exactly-once processing where possible
- Comprehensive audit logging
- Graceful degradation and resilience patterns
- Clear separation between read and write paths where beneficial (CQRS style)

**Deliverables (structure your response like this):**

1. **High-Level Architecture**  
   - Mermaid diagram of the overall system  
   - Description of service boundaries and communication patterns

2. **MVP Scope & Phased Plan**  
   - What to build in MVP (Phase 1)  
   - Recommended order of implementation  
   - What to defer to later phases

3. **Integration Strategy with Existing goledger**  
   - How to use/extend the current ledger  
   - What changes or additions are recommended

4. **Detailed Service Design** (for each MVP service)  
   - Responsibilities  
   - Key APIs / gRPC contracts  
   - Data model highlights  
   - Important business rules

5. **Key Technical Patterns**  
   - Concrete implementation examples or pseudocode for:  
     - Idempotency Key  
     - Saga pattern (example flow, e.g. card issuance or transfer)  
     - Outbox pattern

6. **Database Schema**  
   - Core tables for Ledger + other critical services  
   - Recommendations for indexing and partitioning

7. **Event Schema**  
   - Important domain events that should be published

8. **Roadmap**  
   - From MVP → Production (what to add next: Auth, full KYC, reconciliation jobs, analytics, compliance features, etc.)

9. **Potential Risks & Trade-offs**  
   - Key decisions and their implications

**Additional Instructions:**
- Prioritize correctness and financial integrity over premature optimization.
- Make the design practical and implementable by a small team.
- Include concrete code examples or file structure suggestions where helpful.
- Assume we will start with a monorepo but design services so they can be extracted later.
- Focus first on a solid MVP that can handle real money movement safely.

Begin with the high-level architecture and MVP plan, then go deeper into the details as needed.

   - “Focus more on the Saga implementation examples.”
   - “Include gRPC definitions for inter-service communication.”
