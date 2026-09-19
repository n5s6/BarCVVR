# BarCVVR - Go Migration

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://go.dev)
[![Gin Framework](https://img.shields.io/badge/gin-v1.11-brightgreen.svg)](https://gin-gonic.com/)
[![GORM Database](https://img.shields.io/badge/gorm-v1.31-red.svg)](https://gorm.io/)

This project is a complete and robust migration of the legacy **BarCVVR Ruby on Rails** application into a modern **Go (Golang)** backend. 

The primary goal of this migration was to retain 100% compatibility with the original frontend without altering the existing PostgreSQL database schema. The frontend HTML, CSS, JavaScript (Materialize CSS), and assets have been seamlessly ported over to Go's standard `html/template` system.

---

## 🎯 Features

*   **1:1 Frontend Replication**: The Go application perfectly mimics the old Rails UI, with no stylistic or functional changes for the end-users.
*   **Zero-Migration Database**: The backend utilizes GORM to attach securely to the existing PostgreSQL schema (`users`, `drinks`, `kegs`, `operations`, `beerflows`), expecting the historical data natively.
*   **High Performance**: Migrated from Ruby's interpreted runtime to Go's compiled binary execution, radically improving request throughput and decreasing memory constraints.
*   **Docker Containerized**: Configured with a multi-stage `Dockerfile` exporting a minimal Alpine image under 20MB, far lighter than the previous Rails dependencies.
*   **Test Covered**: Protected by a native `httptest` Go testing strategy running on transient in-memory SQLite tables mirroring production structures.

---

## 🏗 Directory Architecture

```text
GO/
│
├── main.go               # App entrypoint, Gin Router configuration, Static server setup
├── go.mod / go.sum       # Module dependencies tracking
├── Dockerfile            # Multi-stage container instructions
│
├── models/               # GORM structural schema definitions mapped to legacy DB
│   └── models.go
│
├── handlers/             # Core controller logic (replacing Rails Controllers)
│   ├── users.go
│   ├── drinks.go
│   ├── kegs.go
│   ├── operations.go
│   ├── beerflows.go
│   └── *_test.go         # Comprehensive unit coverage suites
│
├── templates/            # Ported Rails ERB templates -> Go html/template
│   ├── layouts/
│   ├── users/
│   ├── drinks/
│   ├── kegs/
│   ├── operations/
│   └── beerflows/
│
└── assets/               # Unchanged legacy static frontend assets
    ├── stylesheets/
    ├── javascripts/
    └── images/
```

---

## 🚀 Environment Configuration

The application requires specific environment variables to function correctly. You can configure these globally in your system or place them in a `.env` file at the root of the `./GO/` directory. 

```env
# REQUIRED: PostgreSQL Connection URI
# Format: postgres://user:password@hostname:port/database
DATABASE_URL=postgres://user:password@localhost:5432/barcvvr

# REQUIRED: Password validating restricted deletion/edit routes
ADMIN_PASSWORD=your_super_secret_password

# OPTIONAL: Web server listening port (Defaults to 3000)
PORT=8080
```

---

## 💻 Local Development Setup

To run the application locally for development:

1. **Verify Go installation**: Ensure Go 1.25+ is installed on your machine (`go version`).
2. **Clone the repository**: Navigate into the newly created `GO` folder.
    ```bash
    cd GO
    ```
3. **Install Go Module Dependencies**:
    ```bash
    go mod tidy
    ```
4. **Boot the Gin Server**:
    Start the local application daemon. It will inherently load your `.env`.
    ```bash
    go run main.go db.go
    ```
    _Tip: If you do not have PostgreSQL running locally, you can start the application using a local SQLite fallback database by appending the `USE_SQLITE` environment variable:_
    ```bash
    USE_SQLITE=true go run main.go db.go
    ```
5. **Access the Application**:
    Navigate to `http://localhost:3000` in your web browser.

---

## 🐳 Docker Production Deployment

This project builds an ultra-lean Docker image by compiling the Go binary inside a `golang-alpine` builder before injecting the raw binary alongside the `templates` and `assets` into a tiny runtime layer.

1. **Build the Production Image**:
    ```bash
    docker build -t barcvvr-backend-go:latest .
    ```

2. **Run as a detached Container**:
    _Note: Substitute the `--env` flags with your production credentials._
    ```bash
    docker run -d \
      --name barcvvr-app \
      -p 3000:3000 \
      -e DATABASE_URL='postgres://user:pass@host/db' \
      -e ADMIN_PASSWORD='pass' \
      barcvvr-backend-go:latest
    ```

---

## ☸️ Kubernetes Deployment

A production-ready Kubernetes manifest is provided internally under the `k8s/` directory. It defines the deployment structure complete with environment variables grouped via a `Secret` and an internal `Service`.

1. **Review and Update Secrets:**
   Open `k8s/deployment.yaml` and replace the placeholder `DATABASE_URL` and `ADMIN_PASSWORD` credentials located under the `Secret` block.

2. **Apply Manifests to Cluster:**
   ```bash
   kubectl apply -f k8s/deployment.yaml
   ```

3. **Routing Configuration:**
   The service `barcvvr-service` exposes the application internally on `TCP 80`. You may then link it to your existing Ingress Controller (e.g., Traefik, Nginx) for exterior routing.

---

## 🧪 Testing and Validation

Comprehensive unit testing is available globally validating the core HTTP Handlers and DB structs. The test-suite builds a localized `:memory:` SQLite cache mimicking Postgres automatically, letting tests behave securely offline without database constraints.

Execute the suite aggressively using Go's built-in toolchain:

```bash
# Run tests recursively through the project
go test -v ./...
```

You should expect cleanly passing states (`PASS`) resolving from the respective handler coverage boundaries identifying successful routing paths and mathematical user amount adjustments natively.
