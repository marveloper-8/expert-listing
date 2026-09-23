# Property Listings & Geospatial Search API

Here is my submission for the Property Listings API backend assessment.

The service is written in **Go 1.24** using **Gin** and **GORM**. It implements full CRUD operations for listings, a geospatial proximity search endpoint using spherical trigonometry (Haversine formula), pagination, input validation with field-level errors, and interactive Swagger documentation.

---

## Quick Start (Swagger in 30 Seconds)

To make evaluating this straightforward and frictionless:
- The app defaults to an embedded SQLite database (`expertlisting.db`).
- On first launch, it automatically seeds 6 realistic property listings across Lagos landmarks (Ikoyi, Victoria Island, Lekki, Ikeja).
- No external database setup, Docker download, or credentials required to test immediately.

```bash
# 1. Start the server
go run cmd/main.go

# 2. Open Swagger UI in your browser
# http://localhost:8080/swagger/index.html
```

---

## Data Structure & Hierarchy

Listings use a clean nested structure grouping location attributes (`address`, `city`, `latitude`, `longitude`) under a `location` object rather than polluting the root level.

### Create Listing Payload (`POST /api/listings`)
```json
{
  "title": "Luxury 3-Bedroom Penthouse in Ikoyi",
  "description": "Penthouse with waterfront views, private elevator, and 24/7 serviced power.",
  "price": 185000000.00,
  "type": "sale",
  "bedrooms": 3,
  "bathrooms": 3,
  "location": {
    "address": "12 Alexander Avenue, Ikoyi",
    "city": "Lagos",
    "latitude": 6.4520,
    "longitude": 3.4350
  },
  "agent_id": "agent_alpha_101"
}
```

### Response Object
```json
{
  "success": true,
  "status_code": 200,
  "message": "Listing retrieved successfully",
  "data": {
    "id": "11111111-1111-1111-1111-111111111111",
    "title": "Luxury 3-Bedroom Penthouse in Ikoyi",
    "description": "Penthouse with waterfront views, private elevator, and 24/7 serviced power.",
    "price": 185000000,
    "type": "sale",
    "bedrooms": 3,
    "bathrooms": 3,
    "location": {
      "address": "12 Alexander Avenue, Ikoyi",
      "city": "Lagos",
      "latitude": 6.452,
      "longitude": 3.435
    },
    "agent_id": "agent_alpha_101",
    "distance_km": 2.45,
    "created_at": "2026-03-23T10:00:00Z",
    "updated_at": "2026-03-23T10:00:00Z"
  }
}
```

---

## How to Test on Swagger UI

Once `go run cmd/main.go` is running, open:
👉 **`http://localhost:8080/swagger/index.html`**

### 1. View Seeded Listings
- Expand `GET /api/listings` $\rightarrow$ **Try it out** $\rightarrow$ **Execute**.
- Returns the 6 pre-seeded properties with pagination metadata.

### 2. Test Geospatial Proximity Search
- Expand `GET /api/listings/search` $\rightarrow$ **Try it out**.
- Fill in:
  - `lat`: `6.4281`
  - `lng`: `3.4219` (Victoria Island center)
  - `radius_km`: `5.0`
- Click **Execute**.
- **What happens**:
  - The API finds properties within 5 km (e.g. VI Flat at 0 km and Ikoyi Villa at ~3.01 km).
  - Farther listings (like Ikeja GRA at ~20 km) are excluded.
  - Results include a computed `distance_km` field and are sorted from nearest to farthest.

### 3. Test Filter Combinations
- On `GET /api/listings/search`:
  - Set `type` to `shortlet`.
  - Set `max_price` to `100000`.
- Click **Execute** to see matching shortlet suites within budget.

### 4. Test Validation Errors
- Expand `POST /api/listings` $\rightarrow$ **Try it out**.
- Try passing a negative price (`"price": -500`) or invalid type (`"type": "lease"`).
- Click **Execute** $\rightarrow$ returns a `422 Unprocessable Entity` with specific field error details:
  ```json
  {
    "success": false,
    "status_code": 422,
    "message": "Validation failed",
    "errors": [
      {
        "field": "price",
        "message": "price must be greater than or equal to 0"
      },
      {
        "field": "type",
        "message": "type must be one of: rent, sale, shortlet"
      }
    ]
  }
  ```

---

## Architecture & Code Structure

I structured the project using layered Clean Architecture:

```
expertlisting/
├── cmd/
│   └── main.go                 # App entrypoint, DI, route registration & graceful shutdown
├── internal/
│   ├── config/                 # Env loader, DB pool config (Postgres + SQLite fallback)
│   ├── database/               # Starter demo data seeder
│   ├── handlers/               # Gin controllers with Swagger tags
│   ├── middleware/             # CORS, request logging, panic recovery
│   ├── models/                 # Domain entities, DTOs & response envelopes
│   ├── repository/             # GORM queries, Haversine geospatial query builder
│   ├── services/               # Business logic, consistency checks & pagination math
│   └── utils/                  # Haversine math, bounding box calc & error formatters
├── tests/                      # Unit & integration test suites
├── docs/                       # Swagger specifications (swag init)
├── Dockerfile                  # Multi-stage Alpine container
└── docker-compose.yml          # Optional Postgres + API container stack
```

### Performance & Security Notes:
1. **Low Latency Read Cache**:
   - In-memory read cache (`internal/cache/cache.go`) backed by `sync.RWMutex` with auto-invalidation on writes.
   - Database PRAGMAs: Write-Ahead Logging (`WAL`), `synchronous = NORMAL`, 64MB page cache, and 256MB memory-mapped I/O (`mmap_size`).
   - GORM configured with `PrepareStmt: true` and `SkipDefaultTransaction: true`.
2. **Security Controls**:
   - Security headers: `Strict-Transport-Security`, `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`, `Content-Security-Policy`.
   - Max body size enforcement (`1MB`) via `http.MaxBytesReader`.
   - Token bucket rate limiter (120 req/min per IP) returning `429 Too Many Requests`.
3. **Repository Pattern with Interfaces**: Decouples business logic from storage, making testing and swapping DB drivers straightforward.
4. **Dual Database Engine**:
   - **Embedded SQLite (`glebarez/sqlite`)**: Default for development and automated tests. Pure Go, no CGO.
   - **PostgreSQL (`gorm.io/driver/postgres`)**: For production / Docker. Connects via connection pooling (`MaxOpen: 50`, `MaxIdle: 25`).
5. **Graceful Shutdown**: The HTTP server catches `SIGINT` and `SIGTERM` signals, giving in-flight requests 5 seconds to complete.

---

## Performance Benchmarks

Run benchmarks locally:
```bash
cd tests && go test -bench=. -benchmem -run=^$
```

Benchmark results on consumer hardware:
```text
BenchmarkGetListings-4        	  123540	      8683 ns/op (0.008 ms)	    2107 B/op	      21 allocs/op
BenchmarkGeospatialSearch-4   	   58843	     18517 ns/op (0.018 ms)	    2663 B/op	      43 allocs/op
BenchmarkHaversineMath-4      	1000000000	         0.56 ns/op         	       0 B/op	       0 allocs/op
```

---

## Geospatial Proximity Approach

To calculate distance between two coordinates, I used the **Haversine formula**:

$$d = 2R \arcsin\left(\sqrt{\sin^2\left(\frac{\Delta \varphi}{2}\right) + \cos(\varphi_1)\cos(\varphi_2)\sin^2\left(\frac{\Delta \lambda}{2}\right)}\right)$$

Where $R = 6371.0 \text{ km}$.

### Query Optimization (Two-Phase Filter):
Running a full trigonometric calculation across every single row in the database is slow. Instead, I implemented a two-phase query:
1. **Bounding Box Pre-Filter**: First calculate minimum and maximum latitude/longitude boundaries for the target radius. The database index on `(latitude, longitude)` filters out non-matching records quickly:
   ```sql
   WHERE latitude BETWEEN minLat AND maxLat AND longitude BETWEEN minLon AND maxLon
   ```
2. **Precise Distance**: The exact Haversine calculation is then only evaluated on candidate records that passed the bounding box check.

---

## Automated Tests

I wrote unit tests for geospatial formulas and integration tests covering the full API lifecycle:

```bash
go test -v ./tests/...
```

Test coverage includes:
- Degrees to radians conversions and boundary edge cases.
- Known real-world distances (Lagos to Abuja, London to Paris).
- Bounding box containment.
- CRUD lifecycle (Create -> Get -> Update -> Delete -> verify 404).
- Request validation errors (422 responses).
- Proximity search including nearby properties and excluding distant ones.
- Pagination metadata calculations.

---

## Production Deployment

### Docker
A multi-stage `Dockerfile` compiles a minimal static binary on Alpine running under a non-root `appuser`.

To run with PostgreSQL:
```bash
docker compose up -d --build
```

### CI/CD
`.github/workflows/deploy.yml` builds the container image and pushes to GitHub Container Registry (GHCR) on pushes to `main`.

---

## What I'd Improve with More Time

1. **PostGIS Support**: For multi-million record datasets, switch to PostgreSQL with PostGIS extensions using `ST_DWithin` with GiST spatial indexing (`R-Tree`).
2. **Redis Geospatial Indexing**: For high-read workloads, cache coordinates using Redis `GEOADD` and `GEOSEARCH` to serve proximity queries with sub-millisecond latency.
3. **Full-Text Search**: Integrate PostgreSQL `tsvector` or Meilisearch for searching keywords in descriptions and street names.
4. **Auth & Ownership**: Add JWT authentication with RBAC so agents can only edit or delete their own listings.
5. **OpenAPI Client Generation**: Add automated client SDK exports from the Swagger definitions.
