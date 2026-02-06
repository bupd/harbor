# Harbor Repository Analysis for AI Agents

## 1. Repository Overview

- **Purpose**: Harbor is an open-source trusted cloud-native registry that stores, signs, and scans container images and Helm charts. It provides a secure and manageable environment for your artifacts.
- **Ownership**: Harbor is a graduated project of the Cloud Native Computing Foundation (CNCF).
- **Homepage**: [goharbor.io](https://goharbor.io/)
- **GitHub**: [github.com/goharbor/harbor](https://github.com/goharbor/harbor)
- **License**: Apache 2.0

## 2. Technology Stack

### Backend
- **Language**: Go (see `GO_VERSION` in `versions.env`)
- **Database**: PostgreSQL
- **Caching/Job Queue**: Redis
- **Core Frameworks**:
    - **Web Framework**: Beego v2 (`github.com/beego/beego/v2`) - MVC pattern with controllers and routing
    - **ORM**: Custom Harbor ORM (`lib/orm`) - Database abstraction with transaction support
    - **API Framework**: go-swagger (`github.com/go-openapi/runtime`) - Auto-generates REST API from OpenAPI spec
    - **Database Migrations**: golang-migrate (`github.com/golang-migrate/migrate/v4`) - Version-controlled schema changes
    - **Authorization**: Casbin (`github.com/casbin/casbin`) - RBAC and policy-based permissions
    - **Caching**: Multi-backend system (`lib/cache/memory`, `lib/cache/redis`)
    - **Configuration**: Multi-source config (`pkg/config/db`, `pkg/config/rest`, `pkg/config/inmemory`)
    - **Logging**: Structured logging (`lib/log`) - Context-aware with configurable levels
    - **Task Queue**: gocraft/work fork (`github.com/goharbor/work v0.5.1-patch`) - Redis-backed job queue with Harbor enhancements
- **Key Dependencies**:
    - `docker/distribution`: Core Docker registry functionality
    - `helm.sh/helm/v3`: Helm chart support
    - `github.com/lib/pq`: PostgreSQL driver
    - `github.com/gomodule/redigo`: Redis client library

### Frontend
- **Language**: TypeScript
- **Core Frameworks**:
    - **Web Framework**: Angular 16 - Component-based SPA framework
    - **UI Kit**: Clarity Design System - VMware's enterprise UI components
    - **Styling**: SCSS - CSS preprocessor with variables and mixins
    - **Testing**: Karma/Jasmine - Unit testing framework
    - **Build**: Angular CLI - Development and build toolchain

### Build & Deployment
- **Build System**: Taskfile (for dev & build), IMPORTANT Makefiles should not be used.
- **Development**: Hybrid approach - native services with hot reload + Docker infrastructure
- **Deployment**: Docker Compose, Helm Chart

## 3. Directory Structure

- `api/v2.0/`: Contains the OpenAPI (Swagger) specification for the Harbor REST API.
- `src/`: The main source code directory (Go module: `github.com/goharbor/harbor/src`).
    - `cmd/`: Main applications for the various Harbor services.
        - `exporter/`: Metrics exporter service
        - `standalone-db-migrator/`: Database migration tool
        - `swagger/`: API documentation generator
    - `common/`: Shared code used across different services.
    - `controller/`: Business logic controllers (artifact, project, user, etc.)
    - `core/`: The core Harbor service, handling most of the logic.
        - `main.go`: Main entry point for core service
        - `controllers/`: Beego controllers for web endpoints
        - `auth/`: Authentication providers (LDAP, OIDC, UAA, etc.)
    - `jobservice/`: Manages background jobs like scanning and replication.
        - `main.go`: Main entry point for job service
        - `job/impl/`: Job implementations (GC, replication, scanning)
    - `lib/`: Core libraries for logging, configuration, ORM, caching, etc.
    - `migration/`: Database migration logic.
    - `pkg/`: Shared packages and utilities (data access layer).
    - `portal/`: The Angular-based frontend UI.
    - `registryctl/`: Controller for interacting with the Docker registry.
        - `main.go`: Registry controller service entry point
    - `server/`: API server implementation and routing.
        - `v2.0/restapi/`: Auto-generated OpenAPI server code
        - `middleware/`: HTTP middlewares
    - `testing/`: Mock implementations and test utilities.
- `make/`: Contains Makefiles and scripts for building and packaging Harbor. Do not use Makefiles anymore. ONLY `make/migrations/postgresql/` is still used.
- `tests/`: Integration and end-to-end tests.
- `taskfile/`: Taskfile configuration for development and build tasks.
- `dockerfile/`: Custom Dockerfiles for building Harbor images.

## 4. Development Workflow

### Essential Commands

**Building:**
- **Build all binaries**: `task build` or `task build:all-binaries`
- **Build specific binary**: `task build:binary:core:linux-amd64`
- **Build all Docker images**: `task images` or `task image:all-images`
- **Build specific image**: `task image:core`
- **Generate API server code**: `task build:gen-apis`
- **Generate frontend API code**: `task build:gen-apis:frontend`

**Development (Hybrid Approach):**
- **Start full environment**: `task dev:up` (Core, JobService, RegistryCtl, Trivy, Portal)
- **Start infrastructure only**: `task dev:infra:up` (PostgreSQL, Valkey, Registry)
- **Start backend containers**: `task dev:backend:up` (~3 sec rebuild with hot reload)
- **Start frontend with HMR**: `task dev:frontend:native` (< 1 sec updates)
- **Install dev tools**: `task dev:setup` (golangci-lint, delve, etc.)

**Testing:**
- **Run all tests**: `task test`
- **Run quick checks**: `task test:quick` (fast validation)
- **Run Go linting**: `task test:lint`
- **Run Go vulnerability check**: `task test:vuln-check`
- **Run Go unit tests**: `task test:unit` or `cd src/ && go test ./...`
- **Lint OpenAPI spec**: `task test:lint:api`
- **Run frontend lint**: `task test:frontend:lint`
- **Run frontend tests**: `task test:frontend:test` or `cd src/portal/ && npm run test`

**Database:**
- **Run migrations**: `task dev:db:migrate`
- **Reset database**: `task dev:db:reset`
- **Open DB shell**: `task dev:db:shell`

**Utilities:**
- **Clean artifacts**: `task clean`
- **Show build info**: `task info`
- **Show dev port info**: `task dev:info`
- **List all tasks**: `task --list-all`

For detailed instructions, see [QUICKSTART.md](QUICKSTART.md).

### Critical Development Guidelines

- **API Changes**: All changes to the REST API *must* be reflected in `api/v2.0/swagger.yaml`. Use `task build:gen-apis` to regenerate the server code after making changes.
- **Database Migrations**: Database schema changes require a new migration file in `make/migrations/postgresql/`.
- **Code Style**: Follow the existing code style. Use `task test:lint` to run linters and formatters.
- **Logging**: Always use Harbor's structured logger (`github.com/goharbor/harbor/src/lib/log`), never Go's standard `log` package. The structured logger produces consistent output with timestamps, levels, and source locations (e.g. `2026-02-06T12:18:16Z [INFO] [/app/src/jobservice/main.go:59]: message`). It works with its default configuration before any explicit initialization, so it is safe to use in early startup code. Use `log.Infof`, `log.Warningf`, `log.Errorf`, `log.Debugf` — not `log.Printf` or `fmt.Printf`.

## 5. Architecture & Patterns

### Service Architecture Overview

Harbor consists of multiple services that communicate via HTTP APIs:

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                    CLIENTS                                          │
├─────────────────────────────┬───────────────────────────────────────────────────────┤
│       Web Browser           │              Docker/OCI Client                        │
│      (Angular UI)           │         (docker push/pull, helm, etc.)                │
└──────────────┬──────────────┴────────────────────────┬──────────────────────────────┘
               │ HTTP :4200 (dev) / :8080 (prod)       │ Docker V2 API :8080
               ▼                                       ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              Harbor Core (port 8080)                                │
│                                                                                     │
│  Entry: src/core/main.go                                                            │
│                                                                                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌────────────────┐  │
│  │  REST API       │  │  Docker V2 API  │  │  Auth Handlers  │  │  Static Files  │  │
│  │  /api/v2.0/*    │  │  /v2/*          │  │  /c/login       │  │  /portal/*     │  │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘  └────────────────┘  │
│           │                    │                    │                               │
│  ┌────────▼────────────────────▼────────────────────▼────────┐                      │
│  │                    Controller Layer                       │                      │
│  │  src/controller/* (artifact, project, user, scan, etc.)   │                      │
│  └────────┬──────────────────────────────────────────────────┘                      │
│           │                                                                         │
│  ┌────────▼──────────────────────────────────────────────────┐                      │
│  │                    Package Layer (DAO)                    │                      │
│  │  src/pkg/* (artifact, project, user, blob, quota, etc.)   │                      │
│  └───────────────────────────────────────────────────────────┘                      │
└───────┬─────────────────┬─────────────────┬─────────────────┬───────────────────────┘
        │                 │                 │                 │
        │ HTTP            │ HTTP            │ TCP             │ TCP
        ▼                 ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────────────────────┐
│  JobService   │ │  RegistryCtl  │ │ Redis/Valkey  │ │         PostgreSQL            │
│  (port 8888)  │ │  (port 8080)  │ │  (port 6379)  │ │        (port 5432)            │
└───────┬───────┘ └───────┬───────┘ └───────────────┘ └───────────────────────────────┘
        │                 │
        │ HTTP            │ HTTP
        ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                         Docker Registry (port 5000)                                 │
│                                                                                     │
│  • Docker Distribution (github.com/distribution/distribution)                       │
│  • Implements Docker Registry V2 API                                                │
│  • Stores blobs and manifests in configurable storage backend                       │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              Storage Backend                                        │
│           (Filesystem, S3, GCS, Azure Blob, OpenStack Swift, Alibaba OSS)           │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Component Details

#### Harbor Core
| Attribute | Value |
|-----------|-------|
| **Port** | 8080 |
| **Entry Point** | `src/core/main.go` |
| **Framework** | Beego v2 web framework |
| **Role** | Central orchestrator and API gateway |

**Responsibilities:**
- REST API for Harbor operations (`/api/v2.0/*`) - auto-generated from OpenAPI spec
- Docker Registry V2 API proxy (`/v2/*`) - handles auth, then proxies to Registry
- User authentication (DB, LDAP, OIDC, UAA, AuthProxy)
- Authorization via RBAC (Casbin-based)
- Project and repository management
- Artifact metadata management (tags, labels, vulnerabilities)
- Quota enforcement
- Webhook notifications
- Serves Portal static files in production

**Key Code Paths:**
- Controllers: `src/controller/` (business logic)
- API handlers: `src/server/v2.0/handler/` (auto-generated)
- Middleware: `src/server/middleware/` (auth, quota, readonly)
- Registry client: `src/pkg/registry/client.go`

---

#### Portal (Web UI)
| Attribute | Value |
|-----------|-------|
| **Port** | 4200 (dev) / served by Core in production |
| **Entry Point** | `src/portal/` |
| **Framework** | Angular 16 + Clarity Design System |
| **Role** | Web-based user interface |

**Responsibilities:**
- Project and repository browsing
- Artifact management (tags, labels, scan results)
- User and robot account management
- System configuration
- Replication and retention policy management
- Audit log viewing

**Key Code Paths:**
- Components: `src/portal/src/app/` (Angular components)
- Services: `src/portal/src/app/shared/services/` (API clients)
- Models: `src/portal/src/app/shared/entities/` (TypeScript interfaces)

**Development:**
```bash
task dev:frontend:native       # Start with hot module reload on :4200
task dev:frontend:test         # Run tests in watch mode
task dev:frontend:test:once    # Run tests once
task test:frontend:lint        # Lint TypeScript/SCSS
```

---

#### JobService
| Attribute | Value |
|-----------|-------|
| **Port** | 8888 |
| **Entry Point** | `src/jobservice/main.go` |
| **Framework** | gocraft/work fork (`github.com/goharbor/work`) |
| **Role** | Asynchronous job execution |

**Responsibilities:**
- Execute background jobs triggered by Core
- Job queue management (Redis-backed)
- Job status tracking and reporting
- Webhook callbacks on job completion
- Supports job types: Generic, Scheduled, Periodic

**Job Types Implemented:**
| Job | Description | Location |
|-----|-------------|----------|
| **Garbage Collection** | Clean up deleted blobs/manifests | `src/jobservice/job/impl/gc/` |
| **Replication** | Sync artifacts between registries | `src/jobservice/job/impl/replication/` |
| **Scan** | Trigger vulnerability scanning | `src/jobservice/job/impl/scan/` |
| **Retention** | Apply retention policies | `src/jobservice/job/impl/retention/` |
| **P2P Preheat** | Warm distributed caches | `src/jobservice/job/impl/preheat/` |
| **Purge Audit** | Clean old audit logs | `src/jobservice/job/impl/purge/` |
| **System Artifact Cleanup** | Remove temporary artifacts | `src/jobservice/job/impl/systemartifact/` |

**Key Code Paths:**
- Job interface: `src/jobservice/job/interface.go`
- Job implementations: `src/jobservice/job/impl/`
- Worker pool: `src/jobservice/worker/`
- API server: `src/jobservice/api/`

**Communication:**
- Core → JobService: HTTP API to submit/query jobs (`src/pkg/task/`)
- Auth: `CORE_SECRET` shared secret

---

#### RegistryCtl
| Attribute | Value |
|-----------|-------|
| **Port** | 8080 (internal only) |
| **Entry Point** | `src/registryctl/main.go` |
| **Role** | Low-level storage operations |

**Responsibilities:**
- Delete manifests from storage (bypasses Registry V2 API)
- Delete blobs from storage
- Health check endpoint
- Direct access to `storage.Vacuum` from docker/distribution

**Why RegistryCtl Exists:**
The Docker Registry V2 API doesn't provide clean deletion APIs suitable for garbage collection. RegistryCtl wraps `storage.Vacuum` from docker/distribution to directly remove files from the storage backend.

**API Endpoints:**
```
DELETE /api/registry/blob/{digest}              # Delete blob
DELETE /api/registry/{repo}/manifests/{ref}     # Delete manifest
GET    /api/health                              # Health check (no auth)
```

**Key Code Paths:**
- Router: `src/registryctl/handlers/router.go`
- Blob handler: `src/registryctl/api/registry/blob/blob.go`
- Manifest handler: `src/registryctl/api/registry/manifest/manifest.go`

**Communication:**
- Core/JobService → RegistryCtl: `src/common/registryctl/client.go`
- Auth: `JOBSERVICE_SECRET` shared secret

---

#### Docker Registry
| Attribute | Value |
|-----------|-------|
| **Port** | 5000 |
| **Image** | `goharbor/registry-photon` (distribution/distribution) |
| **Role** | Artifact storage engine |

**Responsibilities:**
- Implement Docker Registry V2 API specification
- Store and serve blobs (image layers)
- Store and serve manifests (image metadata)
- Content-addressable storage by digest
- Support multiple storage backends

**Storage Backends Supported:**
- Filesystem (default for development)
- Amazon S3
- Google Cloud Storage
- Azure Blob Storage
- OpenStack Swift
- Alibaba OSS

**Communication:**
- Core → Registry: `src/pkg/registry/client.go` (Docker V2 API)
- RegistryCtl → Registry Storage: Direct via `storage.Vacuum`

---

#### PostgreSQL
| Attribute | Value |
|-----------|-------|
| **Port** | 5432 |
| **Role** | Persistent metadata storage |

**Data Stored:**
| Table Category | Examples |
|----------------|----------|
| **Users & Auth** | `harbor_user`, `oidc_user`, `robot`, `user_group` |
| **Projects** | `project`, `project_member`, `project_metadata` |
| **Artifacts** | `artifact`, `artifact_blob`, `tag`, `artifact_reference` |
| **Repositories** | `repository` |
| **Jobs** | `execution`, `task`, `schedule` |
| **Scanning** | `scan_report`, `vulnerability_record` |
| **Policies** | `retention_policy`, `replication_policy`, `immutable_tag_rule` |
| **Quotas** | `quota`, `quota_usage` |
| **Audit** | `audit_log` |
| **System** | `properties`, `job_log`, `notification_policy` |

**Key Code Paths:**
- Migrations: `make/migrations/postgresql/`
- ORM: `src/lib/orm/`
- DAO layer: `src/pkg/*/dao/`

**Development:**
```bash
task dev:db:migrate  # Run migrations
task dev:db:reset    # Reset database
task dev:db:shell    # Open psql shell
```

---

#### Redis / Valkey
| Attribute | Value |
|-----------|-------|
| **Port** | 6379 |
| **Role** | Caching and job queue |

**Data Stored:**
| Purpose | Key Pattern | Description |
|---------|-------------|-------------|
| **Job Queue** | `jobs:*` | Pending/running job data (gocraft/work) |
| **Job Status** | `job_stats:*` | Job execution status |
| **Cache** | `cache:*` | Manifest cache, config cache |
| **Session** | `session:*` | User session data |
| **Idempotency** | `idempotent:*` | Prevent duplicate operations |

**Key Code Paths:**
- Cache abstraction: `src/lib/cache/`
- Redis client: `src/lib/cache/redis/`
- Job queue: Uses gocraft/work (connects directly)

---

### Inter-Service Communication Matrix

| From | To | Protocol | Client Code | Auth | Purpose |
|------|----|----------|-------------|------|---------|
| Core | Registry | HTTP (V2 API) | `src/pkg/registry/client.go` | Basic/Token | Push/pull artifacts |
| Core | RegistryCtl | HTTP | `src/common/registryctl/client.go` | `JOBSERVICE_SECRET` | Delete blobs/manifests |
| Core | JobService | HTTP | `src/pkg/task/` | `CORE_SECRET` | Submit/query jobs |
| Core | PostgreSQL | TCP | `src/lib/orm/` | Password | Metadata CRUD |
| Core | Redis | TCP | `src/lib/cache/redis/` | Password (optional) | Caching |
| JobService | PostgreSQL | TCP | `src/lib/orm/` | Password | Job state |
| JobService | Redis | TCP | gocraft/work | Password (optional) | Job queue |
| JobService | Registry | HTTP (V2 API) | `src/pkg/registry/client.go` | Basic/Token | Artifact operations |
| JobService | RegistryCtl | HTTP | `src/common/registryctl/client.go` | `JOBSERVICE_SECRET` | GC cleanup |
| Portal | Core | HTTP | Angular HttpClient | Session/Token | All operations |

---

### Data Flow Examples

#### Image Push
```
1. Client → Core: POST /v2/repo/blobs/uploads/
   └─ Core authenticates user, checks project permissions
   └─ Core checks quota

2. Core → Registry: POST /v2/repo/blobs/uploads/
   └─ Registry returns upload URL

3. Client → Core → Registry: PATCH/PUT blob data
   └─ Blob stored in storage backend

4. Client → Core: PUT /v2/repo/manifests/tag
   └─ Core validates manifest

5. Core → Registry: PUT /v2/repo/manifests/tag
   └─ Manifest stored

6. Core → PostgreSQL:
   └─ INSERT artifact record
   └─ INSERT tag record
   └─ UPDATE quota usage

7. Core → JobService: Submit scan job (if auto-scan enabled)
```

#### Image Pull
```
1. Client → Core: GET /v2/repo/manifests/tag
   └─ Core authenticates (or allows anonymous if configured)
   └─ Core checks project permissions

2. Core → Redis: Check manifest cache
   └─ If cached, return immediately

3. Core → Registry: GET /v2/repo/manifests/tag
   └─ Registry returns manifest

4. Core → Redis: Cache manifest

5. Core → Client: Return manifest

6. Client → Core → Registry: GET /v2/repo/blobs/{digest}
   └─ Blob streamed to client
```

#### Garbage Collection
```
1. User/Schedule → Core: Trigger GC

2. Core → JobService: Submit GC job
   └─ POST /api/v1/jobs

3. JobService executes GC job:
   a. MARK phase:
      └─ Query PostgreSQL for deleted artifacts
      └─ Query PostgreSQL for unreferenced blobs
      └─ Mark blobs for deletion in PostgreSQL

   b. SWEEP phase (parallel workers):
      └─ For each manifest to delete:
          └─ RegistryCtl.DeleteManifest(repo, digest)
          └─ RegistryCtl → storage.Vacuum.RemoveManifest()

      └─ For each blob to delete:
          └─ RegistryCtl.DeleteBlob(digest)
          └─ RegistryCtl → storage.Vacuum.RemoveBlob()

   c. Update PostgreSQL: Mark as deleted

4. JobService → Core: Webhook callback (job complete)
```

#### Vulnerability Scan
```
1. User/Auto → Core: Trigger scan for artifact

2. Core → JobService: Submit scan job

3. JobService → Scanner Adapter:
   └─ Send artifact info to external scanner (Trivy, Clair, etc.)

4. Scanner → JobService: Return vulnerability report

5. JobService → PostgreSQL:
   └─ Store scan report
   └─ Store vulnerability records

6. JobService → Core: Webhook callback

7. Core → PostgreSQL: Update artifact scan status
```

---

### Layered Code Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Layer                               │
│  src/server/v2.0/handler/    Auto-generated from OpenAPI spec   │
│  src/server/v2.0/restapi/    go-swagger runtime                 │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
┌─────────────────────────────────▼───────────────────────────────┐
│                      Controller Layer                           │
│  src/controller/artifact/      Artifact business logic          │
│  src/controller/project/       Project business logic           │
│  src/controller/scan/          Scanning orchestration           │
│  src/controller/replication/   Replication orchestration        │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
┌─────────────────────────────────▼───────────────────────────────┐
│                      Package Layer (DAO)                        │
│  src/pkg/artifact/             Artifact CRUD + queries          │
│  src/pkg/project/              Project CRUD + queries           │
│  src/pkg/user/                 User management                  │
│  src/pkg/blob/                 Blob tracking                    │
└─────────────────────────────────┬───────────────────────────────┘
                                  │
┌─────────────────────────────────▼───────────────────────────────┐
│                       Library Layer                             │
│  src/lib/orm/                  Database abstraction             │
│  src/lib/cache/                Multi-backend caching            │
│  src/lib/log/                  Structured logging               │
│  src/lib/config/               Configuration management         │
│  src/lib/errors/               Error handling                   │
└─────────────────────────────────────────────────────────────────┘
```

**Design Patterns:**
- **Repository Pattern**: DAO interfaces in `src/pkg/*/dao/`
- **Controller Pattern**: Business logic in `src/controller/`
- **Middleware Chain**: Request processing in `src/server/middleware/`
- **Plugin Architecture**: Scanners, auth providers, storage drivers

## 6. Job Service Framework

Harbor's job service is built on top of a **forked version of gocraft/work** (`github.com/goharbor/work v0.5.1-patch`) with Harbor-specific enhancements:

### Core Features
- **Redis-backed queue**: Reliable job persistence and distribution
- **Job types**: Generic (immediate), Scheduled (delayed), Periodic (cron-like)
- **Enhanced status tracking**: `error`, `success`, `stopped`, `cancelled`, `scheduled`
- **Job control**: Stop, cancel, and retry operations
- **Execution context**: Provides logger, system context, and operation signals
- **Unique jobs**: Prevents duplicate jobs in queue
- **Job executions**: Parent-child job relationships and tracking

### Job Implementation Interface
Jobs must implement the `job.Interface` with methods:
- `MaxFails()`: Retry limit (default: 4)
- `MaxCurrency()`: Concurrency limit per Redis instance
- `ShouldRetry()`: Retry policy for failed jobs
- `Validate(params)`: Parameter validation
- `Run(ctx, params)`: Main job logic

### Key Components
- **API Server**: REST API for job management
- **Controller**: Core coordination and flow control
- **Job Launcher**: Handles Generic and Scheduled jobs
- **Scheduler**: Manages Periodic jobs with cron scheduling
- **Stats Manager**: Job status and webhook notifications
- **Logger Framework**: Multi-backend logging (STD_OUTPUT, FILE, DB)

### Job Types and Use Cases
- **Vulnerability scanning**: Background image analysis
- **Garbage collection**: Cleanup of unused artifacts
- **Replication**: Cross-registry artifact synchronization
- **Retention policies**: Automated artifact lifecycle management
- **P2P preheat**: Distributed cache warming

## 7. Common Development Tasks

### Adding a new API endpoint

1.  Modify `api/v2.0/swagger.yaml` to define the new endpoint.
2.  Run `task build:gen-apis` to generate the server-side code.
3.  Implement the business logic for the new endpoint in the appropriate controller in the `src/controller/` directory.
4.  Add integration tests for the new endpoint in the `tests/` directory.

### Adding a new background job

1.  Define a new job type in the `src/jobservice/` directory.
2.  Implement the job logic.
3.  Trigger the job from the appropriate service (usually `core`).

### Making a frontend change

1.  Navigate to the `src/portal/` directory.
2.  Make your changes to the Angular components.
3.  Run `npm run test` and `npm run lint` to verify your changes.
4.  Build the frontend with `npm run build`.

## 7. Backend Testing & Code Quality

### Linting & Formatting
- `task test:lint` - Run Go linting, format checks, and validation
- `task test:lint-report` - Generate lint report for CI (GitHub Actions format)
- **Linter Configuration**: `.golangci.yaml` with enabled linters:
    - `bodyclose`, `errcheck`, `goheader`, `govet`
    - `ineffassign`, `misspell`, `revive`, `staticcheck`, `whitespace`
- **Formatters**: `gofmt` (no simplify) and `goimports` with local prefixes

### Testing
- `task test:unit` - Run Go unit tests with race detection
- `task test:vuln-check` - Check for Go vulnerabilities
- `task test:lint-api` - Lint OpenAPI spec with Spectral
- From `src/`: `go test ./...` - Run Go unit tests directly
- **Test Structure**: Separate `testing/` package with mocks and test utilities
- **Mock Generation**: Uses `.mockery.yaml` for generating mocks

### Code Generation
- `task build:gen-apis` - Generate API server code from `api/v2.0/swagger.yaml`
- Models and handlers auto-generated using go-swagger
- Copyright header template in `copyright.tmpl`

## 8. Integration Testing
- `tests/ci/api_run.sh` - Run Python API tests
- Robot Framework tests in `tests/robot-cases/`

## 9. Configuration

Harbor uses a YAML configuration file (`make/harbor.yml`) that gets processed by the prepare script. Key configuration areas:
- Database settings (PostgreSQL)
- External services (Redis, registry storage)
- Authentication (LDAP, OIDC)
- Certificate management

## 10. Database

Harbor uses PostgreSQL with migration files in `make/migrations/postgresql/`. Migrations follow a naming pattern `XXXX_version_description.up.sql`.

## 11. Go Development Specifics

### Module Structure
- **Root Module**: `github.com/goharbor/harbor/src`
- **Go Version**: See `GO_VERSION` in `versions.env` (with `godebug x509negativeserial=1`)
- **Package Count**: 280+ packages organized by domain

### Key Entry Points
- `src/core/main.go`: Main Harbor service (Beego web server)
- `src/jobservice/main.go`: Background job processing service
- `src/registryctl/main.go`: Registry controller service
- `src/cmd/exporter/main.go`: Metrics exporter

### Controller Layer Pattern
Controllers in `src/controller/` handle business logic for:
- Artifacts, Projects, Users, Repositories
- Scanning, Replication, Retention policies
- Security, Quotas, Webhooks, P2P preheat

### Data Access Layer
Packages in `src/pkg/` provide:
- DAO interfaces and implementations
- Domain models and DTOs
- Database migration utilities
- Caching abstractions

### Authentication & Authorization
- Multiple auth providers: DB, LDAP, OIDC, UAA, AuthProxy
- RBAC system with project-level and system-level permissions
- Token-based authentication for API access

## 12. Important Notes

- The main branch may be unstable - use releases for stable builds
- Harbor requires Docker 20.10.10+ and docker-compose 1.18.0+
- All API changes require updating the OpenAPI spec in `api/v2.0/swagger.yaml`
- Job Service handles asynchronous operations (vulnerability scanning, garbage collection, replication)
- The system supports multiple container artifact formats (OCI, Helm charts, CNAB bundles)
- Code is auto-generated from OpenAPI - edit spec first, then regenerate



## 12. gopls-mcp — Semantic Go Code Intelligence

[gopls-mcp](https://github.com/xieyuschen/gopls-mcp) is a hard fork of gopls that exposes compiler-grade semantic analysis as MCP tools. It gives Claude type-aware code navigation instead of text search. [Documentation](https://gopls-mcp.org).

### Workspace Setup

The MCP server creates a gopls view from its `-workdir` flag (configured in `.mcp.json`). This must point to the directory containing `go.mod` (in this repo: `src/`). Most tools accept a `Cwd` parameter — if omitted, they use the primary view. If `Cwd` doesn't match any existing view, the call fails.

### Cache Behavior

The first tool call may be slow (~5-10s) as gopls loads and type-checks the workspace. Subsequent calls are fast (~50-150ms). Use `go_build_check` as a good first call — it verifies correctness and warms the cache for all other tools.

### SymbolLocator — Core Input Pattern

Most navigation tools (`go_definition`, `go_symbol_references`, `go_implementation`, `go_get_call_hierarchy`, `go_dryrun_rename_symbol`) use a `SymbolLocator` instead of file:line:column coordinates:

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `symbol_name` | Yes | Exact identifier name | `"EnsureScanners"` |
| `context_file` | Yes | Absolute path to anchor file | `"/path/to/src/core/main.go"` |
| `package_identifier` | No | Import alias if cross-package | `"scan"` |
| `parent_scope` | No | Receiver for methods, function for locals | `"*Server"` |
| `kind` | No | Symbol type filter | `"function"`, `"method"`, `"struct"`, `"interface"`, `"variable"`, `"const"` |
| `line_hint` | No | Approximate line for disambiguation | `356` |
| `signature_snippet` | No | Code snippet to verify match | `"func EnsureScanners("` |

### Key Parameters

- **`include_body: true`** on `go_definition` and `go_get_package_symbol_detail` — returns full function implementation, avoids needing `go_read_file`
- **`include_docs: true`** on `go_list_package_symbols` and `go_get_package_symbol_detail` — includes doc comments
- **`direction`** on `go_get_call_hierarchy` — `"incoming"` (who calls this), `"outgoing"` (what this calls), or `"both"`

### Tool Quick Reference

| Category | Tool | Purpose |
|----------|------|---------|
| **Discovery** | `go_get_started` | Beginner-friendly project overview |
| | `go_analyze_workspace` | Discover packages, entry points, dependencies |
| | `go_list_modules` | List module and direct dependencies |
| | `go_list_module_packages` | List packages in a module |
| | `go_list_package_symbols` | List exported symbols with optional docs/bodies |
| | `go_search` | Fuzzy symbol name search (single token only) |
| **Navigation** | `go_definition` | Jump to symbol definition with optional body |
| | `go_symbol_references` | Find all usages across codebase |
| | `go_implementation` | Find interface implementations or implemented interfaces |
| | `go_get_call_hierarchy` | Bidirectional call graph (incoming + outgoing) |
| | `go_get_package_symbol_detail` | Detailed info for specific symbols |
| | `go_read_file` | Read file content through gopls |
| **Refactoring** | `go_dryrun_rename_symbol` | Preview rename as unified diff (no changes applied) |
| | `go_get_dependency_graph` | Package dependency analysis (imports + dependents) |
| **Verification** | `go_build_check` | Incremental type checking (~500x faster than `go build`) |

### Common Workflows

| Goal | Tool Chain |
|------|------------|
| What does function X do? | `go_search("X")` then `go_definition(include_body=true)` |
| What calls function X? | `go_get_call_hierarchy(direction="incoming")` |
| What does X call? | `go_get_call_hierarchy(direction="outgoing")` |
| What implements interface Y? | `go_implementation(kind="interface")` |
| Explore unfamiliar package | `go_list_package_symbols(include_docs=true)` then `go_get_package_symbol_detail` for specifics |
| Safe rename | `go_symbol_references` (assess impact) then `go_dryrun_rename_symbol` (preview diff) |
| Are there build errors? | `go_build_check` (also warms cache) |
| Location unknown | `go_search("SymbolName")` to find file, then targeted tool |

### go_search Limitations

`go_search` is a **symbol name index**, not a text search engine. The query must be a single identifier token.

**What it searches:** top-level functions, types, variables, constants, methods, struct fields, function parameters, interface methods, named return values.

**What it cannot search:** string literals, comments, doc content, code patterns, local variables, natural language descriptions. For text content, use the Grep tool instead.

**Tip:** If `go_search("LongSpecificName")` returns nothing, try a shorter suffix (e.g., `"SpecificName"`). Never use spaces or sentences.

### When to Use gopls-mcp vs Grep

| Use gopls-mcp | Use Grep |
|---------------|----------|
| Find symbol definitions | Search string literals or comments |
| Find interface implementations (discovers unexported types and mocks automatically) | Search for TODO/FIXME/HACK markers |
| Trace call hierarchies | Find log messages or error strings |
| Assess rename/refactor impact | Quick filename-based search |
| Build error checking | Search across non-Go files |
