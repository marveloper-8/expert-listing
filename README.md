# Property Listings & Geospatial Search API

A backend service built in Go for managing real estate listings with full CRUD support, geospatial radius searching (Haversine formula), filtering, pagination, and input validation.

### Live Demo & API Docs
- **Interactive Swagger UI**: [https://expertlisting.nexalabs.site/swagger/index.html](https://expertlisting.nexalabs.site/swagger/index.html)
- **Base API URL**: `https://expertlisting.nexalabs.site/api/v1`

---

## Quick Start (Local Setup)

The project runs on an embedded SQLite database by default with zero setup required. On startup, it automatically seeds realistic demo properties in Lagos (Ikoyi, Victoria Island, Lekki, Ikeja).

```bash
# Clone the repository
git clone https://github.com/marveloper-8/expert-listing.git
cd expert-listing

# Start the server (runs on port 8082)
go run cmd/main.go
```

Once running, access Swagger locally at:
`http://localhost:8082/swagger/index.html`

To run using Docker:
```bash
docker compose up -d --build
```

---

## Data Structure

Listings group geographical details inside a nested `location` object:

### Create Listing (`POST /api/v1/listings`)
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

### Standard Response
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

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/listings` | List properties with pagination (`page`, `limit`) |
| `GET` | `/api/v1/listings/:id` | Get single property by UUID |
| `POST` | `/api/v1/listings` | Create a new property listing |
| `PUT` | `/api/v1/listings/:id` | Update an existing property listing |
| `DELETE` | `/api/v1/listings/:id` | Remove a property listing |
| `GET` | `/api/v1/listings/search` | Search by radius (`lat`, `lng`, `radius_km`) and filters |
| `GET` | `/health` | Health check endpoint |

---

## Geospatial Proximity Search

Radius queries (`GET /api/v1/listings/search?lat=6.4281&lng=3.4219&radius_km=5`) use a two-step approach:

1. **Bounding Box Pre-filter**: Calculates the rough min/max latitude and longitude boundaries for the requested radius. This lets the database index on `(latitude, longitude)` do the heavy lifting by ignoring records outside the box.
2. **Exact Haversine Calculation**: The server applies the exact Haversine formula only to matching candidates, computing the distance in kilometers and sorting nearest-first.

---

## Engineering Details

- **Concurrency & Caching**: In-memory read cache with `sync.RWMutex` invalidation on writes. Read endpoints respond in fractions of a millisecond.
- **SQLite Optimization**: Tuned with Write-Ahead Logging (`WAL`), `synchronous = NORMAL`, and memory-mapped I/O (`mmap_size`) for fast concurrent reads.
- **Validation**: Strict validation with field-level errors (e.g. valid property types: `sale`, `rent`, `shortlet`; latitude -90 to 90; longitude -180 to 180).
- **Security**:
  - OWASP recommended security headers (`X-Frame-Options`, `Content-Security-Policy`, `X-Content-Type-Options`).
  - Request body size capped at 1MB to prevent memory exhaustion attacks.
  - Per-IP rate limiting (120 req/min) returning `429 Too Many Requests`.
- **Dual DB Support**: Works with zero-dependency embedded SQLite (`glebarez/sqlite`) out of the box, or PostgreSQL in production by passing `DATABASE_URL`.
- **Graceful Shutdown**: Intercepts `SIGINT` and `SIGTERM` to safely finish active requests before exiting.

---

## Running Tests

Unit and integration tests run against an in-memory SQLite database:

```bash
# Run all tests
go test -v ./tests/...

# Run benchmarks
cd tests && go test -bench=. -benchmem -run=^$
```

Benchmark snapshot:
```text
BenchmarkGetListings-4        	  123540	      8683 ns/op (0.008 ms)
BenchmarkGeospatialSearch-4   	   58843	     18517 ns/op (0.018 ms)
BenchmarkHaversineMath-4      	1000000000	         0.56 ns/op
```

---

## Project Structure

```
.
├── cmd/
│   └── main.go                 # Entrypoint, route setup & graceful shutdown
├── internal/
│   ├── cache/                  # Thread-safe in-memory cache
│   ├── config/                 # Environment and DB connection pooling
│   ├── database/               # Initial demo seed data
│   ├── handlers/               # HTTP request handlers with Swagger tags
│   ├── middleware/             # Rate limit, security headers, logger, CORS
│   ├── models/                 # Domain models and request/response DTOs
│   ├── repository/             # Database queries & bounding box filtering
│   ├── services/               # Core business logic & cache orchestration
│   └── utils/                  # Haversine math & validation error formatters
├── tests/                      # Unit, integration, and benchmark tests
├── docs/                       # Auto-generated Swagger spec files
├── Dockerfile                  # Alpine production container
└── docker-compose.yml          # Local containerized setup
```

---

## Next Steps for Scale

1. **PostGIS**: For multi-million record datasets, use PostgreSQL with PostGIS (`ST_DWithin` with GiST spatial indexing).
2. **Distributed Cache**: Swap the in-memory cache with Redis (`GEOADD` / `GEOSEARCH`) for distributed instances.
3. **Authentication**: Add JWT authentication so agents can only edit or delete listings they own.
