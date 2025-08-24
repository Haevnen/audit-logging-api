## 1. System Architecture Overview
- This diagram shows the complete system architecture and how all components
interact with each other.

```mermaid
graph TB
    %% ========== CLIENT LAYER ==========
    subgraph "Client Layer"
        Client["Client Application"]
        Browser["Web Browser"]
        Mobile["Mobile App"]
    end

    %% ========== API GATEWAY ==========
    subgraph "API Gateway"
        LB["Load Balancer / ALB"]
        Auth["JWT Authentication"]
        RateLimit["Rate Limiting"]
        Validation["Request Validation"]
    end

    %% ========== API LAYER ==========
    subgraph "API Layer"
        LogAPI["Log API<br/>(POST /logs)"]
        BulkAPI["Bulk Log API<br/>(POST /logs/bulk)"]
        SearchAPI["Search API<br/>(GET /logs?filters)"]
        StatAPI["Stats API<br/>(GET /logs/stats)"]
        ExportAPI["Export Logs API<br/>(GET /logs/export)"]
        StreamAPI["Log Stream<br/>(WS /logs/stream)"]
        TenantAPI["Tenant API<br/>(/tenants)"]
    end

    %% ========== USECASE / SERVICE LAYER ==========
    subgraph "Usecase & Service Layer"
        LogUC["Log Usecases<br/>(Create - Get - Search - Stats - Export)"]
        TenantUC["Tenant Usecases<br/>(Create - List)"]
        AuthSvc["Auth Service<br/>(JWT Manager)"]
        PubSubSvc["PubSub Service<br/>Redis"]
    end

    %% ========== REPOSITORY LAYER ==========
    subgraph "Repository Layer"
        LogRepo["Log Repository"]
        TenantRepo["Tenant Repository"]
        TaskRepo["Async Task Repo"]
        OpenSearchRepo["OpenSearch Repo"]
    end

    %% ========== MESSAGE QUEUE ==========
    subgraph "Message Queue"
        SQS["AWS SQS"]
        ArchivalQueue["Archival Queue"]
        CleanupQueue["Cleanup Queue"]
        IndexQueue["Index Queue"]
    end

    %% ========== WORKERS ==========
    subgraph "Background Workers"
        ArchiveWorker["Archive Worker<br/>(S3 Upload + Cleanup trigger)"]
        CleanupWorker["Cleanup Worker<br/>(DB + OpenSearch Cleanup)"]
        IndexWorker["Index Worker<br/>(Sync to OpenSearch)"]
    end

    %% ========== DATA STORAGE ==========
    subgraph "Data Storage"
        Postgres["PostgreSQL + TimescaleDB"]
        S3["S3 Bucket<br/>(Archived Logs)"]
        OpenSearch["OpenSearch Cluster"]
        Redis["Redis<br/>(Pub/Sub)"]
    end

    %% ========== FLOWS ==========
    Client --> LB
    Browser --> LB
    Mobile --> LB

    LB --> Auth
    Auth --> RateLimit
    RateLimit --> Validation

    Validation --> LogAPI
    Validation --> BulkAPI
    Validation --> SearchAPI
    Validation --> StatAPI
    Validation --> ExportAPI
    Validation --> StreamAPI
    Validation --> TenantAPI

    LogAPI --> LogUC
    BulkAPI --> LogUC
    SearchAPI --> LogUC
    StatAPI --> LogUC
    ExportAPI --> LogUC
    StreamAPI --> PubSubSvc
    TenantAPI --> TenantUC

    LogUC --> LogRepo
    TenantUC --> TenantRepo
    LogUC --> TaskRepo
    LogUC --> OpenSearchRepo

    LogRepo --> Postgres
    TenantRepo --> Postgres
    TaskRepo --> Postgres
    OpenSearchRepo --> OpenSearch

    PubSubSvc --> Redis
    Redis --> StreamAPI

    %% Async flows
    LogUC -.-> SQS
    SQS -.-> ArchivalQueue
    SQS -.-> CleanupQueue
    LogUC -.-> IndexQueue

    ArchivalQueue -.-> ArchiveWorker
    CleanupQueue -.-> CleanupWorker
    IndexQueue -.-> IndexWorker

    ArchiveWorker --> S3
    ArchiveWorker -.-> CleanupQueue
    CleanupWorker --> Postgres
    CleanupWorker --> OpenSearch
    IndexWorker --> OpenSearch

    %% ========== STYLE ==========
    classDef client fill:#e1f5fe,stroke:#0288d1,stroke-width:1px
    classDef gateway fill:#e8eaf6,stroke:#3f51b5,stroke-width:1px
    classDef api fill:#ede7f6,stroke:#673ab7,stroke-width:1px
    classDef service fill:#e8f5e9,stroke:#388e3c,stroke-width:1px
    classDef repo fill:#fff3e0,stroke:#ef6c00,stroke-width:1px
    classDef mq fill:#fce4ec,stroke:#c2185b,stroke-width:1px
    classDef worker fill:#f1f8e9,stroke:#689f38,stroke-width:1px
    classDef storage fill:#f3e5f5,stroke:#8e24aa,stroke-width:1px

    class Client,Browser,Mobile client
    class LB,Auth,RateLimit,Validation gateway
    class LogAPI,BulkAPI,SearchAPI,StatAPI,ExportAPI,StreamAPI,TenantAPI api
    class LogUC,TenantUC,AuthSvc,PubSubSvc service
    class LogRepo,TenantRepo,TaskRepo,OpenSearchRepo repo
    class SQS,ArchivalQueue,CleanupQueue,IndexQueue mq
    class ArchiveWorker,CleanupWorker,IndexWorker worker
    class Postgres,S3,OpenSearch,Redis storage
```

## 2. Sequence flow
### 2.1 Create log flow
```mermaid
sequenceDiagram
    autonumber
    participant Client as Client (App / Browser)
    participant Handler as LogHandler (POST /logs)
    participant UC as CreateLogUseCase
    participant Repo as LogRepository
    participant TaskRepo as AsyncTaskRepository
    participant Tx as TxManager
    participant DB as PostgreSQL/TimescaleDB
    participant SQS as AWS SQS (Index Queue)
    participant PubSub as Redis PubSub
    participant Worker as IndexWorker
    participant OS as OpenSearch

    Client->>Handler: POST /logs (JSON + JWT)
    Handler->>UC: Execute(tenantId, userId, log)

    UC->>Tx: TransactionExec()
    Tx->>Repo: CreateBulk(log)
    Repo->>DB: INSERT log
    DB-->>Repo: OK

    Tx->>TaskRepo: Create(async_task TaskReindex)
    TaskRepo->>DB: INSERT async_task
    DB-->>TaskRepo: OK

    Tx->>SQS: PublishIndexMessage(taskID, log)
    Tx-->>UC: Transaction Commit

    UC->>PubSub: BroadcastLog(log)
    UC-->>Handler: return created log
    Handler-->>Client: 201 Created (logId, timestamp)

    %% Background Worker
    SQS-->>Worker: Deliver message (taskID, logs)
    Worker->>TaskRepo: GetByID(taskID)
    TaskRepo-->>Worker: Pending task
    Worker->>TaskRepo: UpdateStatus(RUNNING)

    Worker->>OS: IndexLogsBulk(logs)
    OS-->>Worker: OK

    Worker->>TaskRepo: UpdateStatus(SUCCEEDED)
    Worker-->>SQS: DeleteMessage
```

### 2.2 Log cleanup flow
```mermaid
sequenceDiagram
    participant Client as Client (API Caller)
    participant API as LogHandler (DELETE /logs/cleanup)
    participant UC as DeleteLogUseCase
    participant Tx as TxManager
    participant AsyncRepo as AsyncTaskRepository
    participant SQS as SQS (Archive Queue)
    participant ArchWorker as ArchiveWorker
    participant LogRepo as LogRepository
    participant S3 as S3 Bucket
    participant Tx2 as TxManager (ArchiveWorker)
    participant AsyncRepo2 as AsyncTaskRepository
    participant SQS2 as SQS (Cleanup Queue)
    participant CleanWorker as CleanUpWorker
    participant LogRepo2 as LogRepository
    participant OpenSearch as OpenSearch

    %% --- Delete request from client ---
    Client->>API: DELETE /logs/cleanup (with beforeDate, tenantId)
    API->>UC: Execute(tenantId, userId, beforeDate)

    %% --- Create async task & push to SQS ---
    UC->>Tx: TransactionExec
    Tx->>AsyncRepo: Create async_task (type=Archive, status=Pending)
    AsyncRepo-->>Tx: TaskID
    Tx->>SQS: PublishArchiveMessage(TaskID, beforeDate)
    Tx-->>UC: success
    UC-->>API: return 202 Accepted

    %% --- Archive Worker consumes SQS ---
    SQS->>ArchWorker: Deliver Archive Task (TaskID, beforeDate)
    ArchWorker->>AsyncRepo2: GetByID(TaskID)
    ArchWorker->>AsyncRepo2: UpdateStatus(RUNNING)
    ArchWorker->>LogRepo: FindLogsForArchival(tenantId, beforeDate)
    LogRepo-->>ArchWorker: [Logs]
    ArchWorker->>S3: UploadLogs(TaskID, Logs)
    ArchWorker->>Tx2: TransactionExec
    Tx2->>AsyncRepo2: UpdateStatus(SUCCEEDED)
    Tx2->>AsyncRepo2: Create async_task (type=Cleanup, status=Pending)
    Tx2->>SQS2: PublishCleanUpMessage(CleanupTaskID, beforeDate)
    Tx2-->>ArchWorker: success

    %% --- Cleanup Worker consumes SQS ---
    SQS2->>CleanWorker: Deliver Cleanup Task (CleanupTaskID, beforeDate)
    CleanWorker->>AsyncRepo2: GetByID(CleanupTaskID)
    CleanWorker->>AsyncRepo2: UpdateStatus(RUNNING)
    CleanWorker->>LogRepo2: CleanupLogsBefore(beforeDate)
    LogRepo2-->>CleanWorker: [DeletedLogIDs]
    CleanWorker->>OpenSearch: DeleteLogsBulk(DeletedLogIDs)
    CleanWorker->>AsyncRepo2: UpdateStatus(SUCCEEDED)
    CleanWorker-->>SQS2: Done
```
```bash
task_id                             |status   |task_type  |payload|created_at                   |updated_at                   |tenant_uid                          |user_id |error_msg|
------------------------------------+---------+-----------+-------+-----------------------------+-----------------------------+------------------------------------+--------+---------+
844ecbda-e848-4e77-b55a-880a2dc8a42b|succeeded|reindex    |       |2025-08-24 10:14:30.586 +0700|2025-08-24 10:14:31.029 +0700|                                    |123e4567|         |
c45fdacf-a1b7-4070-8713-6908ad0704e5|succeeded|reindex    |       |2025-08-24 10:17:11.755 +0700|2025-08-24 10:17:11.955 +0700|                                    |123e4567|         |
57187837-01d4-4965-b9ca-80db9ea48396|succeeded|reindex    |       |2025-08-24 10:23:34.447 +0700|2025-08-24 10:23:34.667 +0700|                                    |123e4567|         |
9e4109ad-7014-4918-9bd7-401099422ac9|succeeded|archive    |       |2025-08-24 10:32:16.931 +0700|2025-08-24 10:32:17.218 +0700|c2e217a3-fe51-444c-a41c-6956c4c81d51|123e4567|         |
d3381dd4-b3a6-4c1d-9025-b31449fe435b|succeeded|log_cleanup|       |2025-08-24 10:32:17.220 +0700|2025-08-24 10:32:17.405 +0700|c2e217a3-fe51-444c-a41c-6956c4c81d51|123e4567|         |
```
- I created a async task table to tracking the progress of all background job (archival, indexing, cleanup)
    - Status: Pending, Running, Succeeded, Failed
    - Task_type: Log_cleanup, reindex, archive

### 2.3 Search log flow (advanced search)
```mermaid
sequenceDiagram
    participant Client as Client (API Caller)
    participant API as LogHandler (GET /api/v1/logs?filters)
    participant UC as SearchLogsUseCase
    participant Repo as LogSearchRepository (OpenSearch)
    participant OS as OpenSearch Cluster

    %% --- Request from client ---
    Client->>API: GET /logs?tenant_id&user_id&action&severity&q&page&page_size

    %% --- Handler builds filters ---
    API->>API: Parse query params<br/>Build LogSearchFilters
    API->>UC: Execute(ctx, filters)

    %% --- Usecase calls repository ---
    UC->>Repo: Search(ctx, filters)

    %% --- Repository builds query ---
    Repo->>Repo: buildQuery(filters, from)
    Repo->>OS: POST /index/_search<br/>with ES query JSON

    %% --- OpenSearch responds ---
    OS-->>Repo: Hits {total, logs[]}
    Repo-->>UC: SearchResult {Total, Logs}
    UC-->>API: SearchResult

    %% --- Handler maps to API response ---
    API->>API: Convert logs to DTO
    API-->>Client: 200 OK + {total, items, pageNumber, pageSize}
```