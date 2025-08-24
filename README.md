# Audit Logging API  

## 1. Overview  
- Audit Logging API is a multi-tenant, high-performance logging system designed
  to track and manage user actions across different applications.
- Built with **Go**, **PostgreSQL/Timescaledb**, **AWS SQS/S3**, **Redis** and **OpenSearch**, this system ensures **data integrity, tenant isolation, and scalability** for enterprise-grade workloads.  

---

## 2. Features  
- **Log Management**  
  - Create single or bulk log entries with metadata  
  - Structured schema: user, tenant, action, resource, before/after state, severity, timestamp  

- **Search & Retrieval**  
  - Filter logs by date, user, action type, severity, tenant  
  - Full-text search in messages & metadata & before/after state (via OpenSearch)  
  - Pagination for large datasets  

- **Export & Streaming**  
  - Export logs in JSON or CSV (can support large amount of logs)
  - Real-time WebSocket streaming  

- **Tenant Management**  
  - Strict tenant isolation  
  - Admin-only tenant creation & listing  

- **Data Management**  
  - Configurable retention (through cleanup API)
  - Data compression (database level)
  - Archival policies by backing up and store in S3 through background workers
  - Cleanup via async tasks  

- **Security & Performance**  
  - JWT-based authentication  
  - Role-based authorization (`Admin`, `Auditor`, `User`)  
  - Rate limiting per tenant (the same threshold value for each tenant)
  - 1000+ logs/sec throughput  

---
## 3. Project Structure

```
├── Makefile                    # Automation command
├── README.md                   # Project overview
├── api                         # OpenAPI specifications
│   ├── api_service.v1.yaml     # API definition
│   └── gen
│       └── specs               # Generated OpenAPI spec
├── cmd                         # Application entry points
│   ├── async-task              # Background async tasks (archival, cleanup, indexing)
│   └── audit-logging-api       # Main API server entrypoint
├── docker-compose.yml          # Docker service
├── internal                    
│   ├── adapter
│   │   └── http                # Handler code
│   │       ├── gen
│   │       │   └── api         # Generated server code
│   ├── apperror                # Centralized application error handling
│   ├── auth                    # JWT authentication
│   ├── config                  # Environment config loader
│   ├── constant                # Constants used across modules
│   ├── entity                  # Domain entities
│   │   ├── async_task
│   │   ├── log
│   │   └── tenant
│   ├── infra                   # HTTP middleware (auth, ratelimit)
│   │   └── middleware
│   ├── interactor              # Transaction manager
│   ├── registry                # Dependency injection
│   ├── repository              # Database persistent layer
│   ├── service                 # Infrastructure service (S3, SQS, Redis)
│   ├── usecase                 # Business logic
│   │   ├── log
│   │   └── tenant
│   └── worker                  # Background worker (archival, cleanup, indexing)
├── localstack_bootstrap        # Init script for Localstack
├── migrations                  # Database migrations
│   ├── files
├── opensearch_bootstrap        # Init script for Opensearch
├── pkg                         # Shared utilities
```
---
## 4. API Endpoints

| Method | Endpoint               | Roles Allowed        | Description             |
| ------ | ---------------------- | -------------------- | ----------------------- |
| POST   | `/api/v1/logs`         | Admin, User          | Create a log entry      |
| POST   | `/api/v1/logs/bulk`    | Admin, User          | Create logs in bulk     |
| GET    | `/api/v1/logs`         | Admin, Auditor, User | Search / filter logs    |
| GET    | `/api/v1/logs/{id}`    | Admin, Auditor, User | Get single log entry    |
| GET    | `/api/v1/logs/stats`   | Admin, Auditor, User | Log statistics          |
| GET    | `/api/v1/logs/export`  | Admin, Auditor       | Export logs (JSON/CSV)  |
| DELETE | `/api/v1/logs/cleanup` | Admin, User          | Cleanup old logs        |
| WS     | `/api/v1/logs/stream`  | Admin, Auditor, User | Real-time log streaming |
| GET    | `/api/v1/tenants`      | Admin                | List tenants            |
| POST   | `/api/v1/tenants`      | Admin                | Create new tenant       |

- Details: http://localhost:8080/ (Swagger UI)

---
## 5. Installation & Setup

### 5.1 Prerequisites
- **Go 1.23+** installed
- **Docker** and **Docker Compose** installed
- **Make** utility (for running Makefile commands)

### 5.2 Clone repository
```bash
git clone https://github.com/Haevnen/audit-logging-api.git
cd audit-logging-api
```

### 5.3 Build and start dependencies services
```bash
make run
make codegen
make migrate
```

### 5.4 Start API server and worker
```bash
make server
make worker # using another terminal
```
---

## 6 Testing
- Unittest: 54%
```bash
make test
ok      github.com/Haevnen/audit-logging-api/internal/adapter/http      1.967s  coverage: 71.6% of statements
ok      github.com/Haevnen/audit-logging-api/internal/auth      1.682s  coverage: 89.5% of statements
ok      github.com/Haevnen/audit-logging-api/internal/config    1.445s  coverage: 55.6% of statements
ok      github.com/Haevnen/audit-logging-api/internal/infra/middleware  2.103s  coverage: 93.9% of statements
        github.com/Haevnen/audit-logging-api/internal/interactor                coverage: 0.0% of statements
        github.com/Haevnen/audit-logging-api/internal/repository                coverage: 0.0% of statements
        github.com/Haevnen/audit-logging-api/internal/service           coverage: 0.0% of statements
ok      github.com/Haevnen/audit-logging-api/internal/usecase/log       2.365s  coverage: 93.9% of statements
ok      github.com/Haevnen/audit-logging-api/internal/usecase/tenant    2.605s  coverage: 100.0% of statements
ok      github.com/Haevnen/audit-logging-api/internal/worker    3.222s  coverage: 86.6% of statements
>>> Overall Coverage:
total:                                                                                (statements)                     54.1%
```
- Integration test: API endpoint tested with [Postman
  collection](docs/Audit%20Logging%20API.postman_collection.json)

## 7 Documentation
- [System Architecture](docs/System_architecture.md)
- [Design decision](docs/Design_decision.md)
- [Database design](docs/Database_design.md)
- [OpenAPI 3.0 spec](http://localhost:8080/)
- [Postman collection](docs/Audit%20Logging%20API.postman_collection.json)