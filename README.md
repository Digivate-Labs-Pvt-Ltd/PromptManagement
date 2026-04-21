# Prompt Management Service

A production-grade, versioned Prompt Management Service built in native Go (`net/http`) and PostgreSQL. This centralized registry allows development and engineering teams to store, version, and manage logic prompts, configurations (e.g., `temperature`, `top_k`), and operational scopes without relying on external UI blockers.

> **Note**: This service handles **configuration storage and history tracking only**. It does NOT call external LLM providers or manage Vector Similarity queries natively. Vector prompt instructions are stored as reference `TEXT` objects for execution by upstream services.

---

## 🚀 Features
- **Stateless Authentication**: JWT (`HS256`) access with Secure Password Hashing (`bcrypt`).
- **Immutable History**: Prompt Updates are natively logged as semantic-versioned, immutable rows (e.g. `v1.0.0` -> `v1.0.1`).
- **Transactional Promotions**: Push specific prompt versions from `draft` to `active` via sub-second, atomic PostgreSQL bounds.
- **Strict Method Scopes**: Completely `POST`-driven API routes encapsulating complex state transformations.
- **Dependency-Free HTTP**: Native Go 1.22+ `ServeMux` implementation without external heavy-weight frameworks like Gin or Fiber.

---

## 🛠️ Prerequisites

To run this application locally, you will need the following installed:
- [Go (1.21 or later)](https://go.dev/doc/install)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/) (For the PostgreSQL Database)
- [Make](https://www.gnu.org/software/make/)
- [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (For running database migrations)

---

## 🏁 Getting Started

Follow these steps to configure, build, and start the Prompt Management environment.

### 1. Environment Configuration
Copy the sample environment variable file and generate your actual configuration points (especially useful for altering database connection settings if hitting external servers).
```bash
cp .env.example .env
```
*Note: Make sure your `.env` contains the required `DATABASE_URL` pointing to the internal Docker container (see below).*

### 2. Infrastructure Boot Initialization
Start the detached PostgreSQL database container via Compose.
```bash
make docker-up
```
> **Tip:** If the port `5432` is already in use by a host installation of Postgres, update your `docker-compose.yml` to export `5433:5432` and update the `DATABASE_URL` in `.env` accordingly.

### 3. Run Database Migrations
Run the initial SQL schema migrations utilizing the explicit `make` command wrapper to populate the relational schemas (Users, Prompt Groups, Items).
```bash
make migrate-up
```

### 4. Start the Application
Compile and execute the Go API server using the built-in watcher:
```bash
make run
```
The server will now be live on port `8080` (or whichever port is defined in `.env`).

---

## 💻 Available `Makefile` Commands

| Command | Description |
|---|---|
| `make build` | Compiles the binary into `tmp/api` |
| `make run` | Compiles and starts the server via the newly created binary |
| `make docker-up` | Boots the `docker-compose.yml` local infrastructure stack |
| `make docker-down`| Tears down the dockerized architecture stack |
| `make migrate-up` | Applies the PostgreSQL schemas mapping incrementally |
| `make migrate-down`| Rolls back the most recent iteration of schema modifications |
| `make test` | Runs the test files recursively across all isolated packages |
| `make lint` | Validates styles universally against `golangci-lint` |

---

## 🌐 API Overview

All authenticated endpoints require a header passed as: 
`Authorization: Bearer <your-jwt-token>`

### Auth
- **`POST /auth/register`** - Registers a new user internally (`email`, `username`, `full_name`, `password`).
  ```bash
  curl -X POST http://localhost:8080/auth/register \
       -H "Content-Type: application/json" \
       -d '{"email":"test@example.com", "username":"testuser", "full_name":"Test User", "password":"password123"}'
  ```
- **`POST /auth/login`** - Authenticates the user and dispatches a JWT token string matching internal scopes.
  ```bash
  curl -X POST http://localhost:8080/auth/login \
       -H "Content-Type: application/json" \
       -d '{"username":"testuser", "password":"password123"}'
  ```

### Prompt Management (Groups)
The `Management Groups` serve as the absolute parent logic directories.
- **`POST /prompts/create`** - Create a management group parameters (`client`, `use_case`).
  ```bash
  curl -X POST http://localhost:8080/prompts/create \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer <your-jwt-token>" \
       -d '{"client":"Acme", "use_case":"Support", "document_type":"Email", "category":"Sales", "stage_name":"Initial"}'
  ```
- **`POST /prompts/update`** - Update overarching details excluding direct query prompts.
- **`POST /prompts/list`** - Fetches paginated directories of active mappings merging User interactions.
- **`POST /prompts/get`** - Retrieves detailed group metadata plus all associated prompt versions (Aggregated Fetch).
  ```bash
  curl -X POST http://localhost:8080/prompts/get \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer <your-jwt-token>" \
       -d '{"id":"<uuid>"}'
  ```

### Bulk Operations
- **`POST /prompts/create-full`** - Atomic creation of a management group and its initial prompt items.
  ```bash
  curl -X POST http://localhost:8080/prompts/create-full \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer <your-jwt-token>" \
       -d '{
         "client": "Acme",
         "use_case": "Support",
         "document_type": "JSON",
         "category": "Sales",
         "stage_name": "Initial",
         "items": [
           {
             "question_key": "Q1",
             "prompt_text": "Analyze text",
             "vector_prompt": "Optional vector instruction",
             "generation_config": {"temperature": 0.7},
             "top_k": 5,
             "change_log": "Initial seed"
           }
         ]
       }'
  ```
  *Note: Prompts created via this endpoint are automatically promoted to `active` status (`v1.0.0`).*

### Prompt Items (Versions)
Prompt `Items` operate within a given Management Group, housing infinite immutable execution iterations.
- **`POST /prompts/items/add`** - Add a target representation triggering automated `Version Auto-bumping` (`v.1.0.X`).
  ```bash
  curl -X POST http://localhost:8080/prompts/items/add \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer <your-jwt-token>" \
       -d '{"management_id":"<uuid>", "question_key":"Q1", "prompt_text":"Analyze this text"}'
  ```
- **`POST /prompts/items/list`** - Return arrays encompassing available configurations sorted against parent parameters.
- **`POST /prompts/items/promote`** - Transactionally demotes existing structures and executes Live State bindings linking backwards to origin Group properties.
  ```bash
  curl -X POST http://localhost:8080/prompts/items/promote \
       -H "Content-Type: application/json" \
       -H "Authorization: Bearer <your-jwt-token>" \
       -d '{"management_id":"<uuid>", "item_id":"<uuid>"}'
  ```
