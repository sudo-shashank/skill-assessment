# Go PDF Report Service

Standalone microservice that generates a downloadable student PDF by calling the existing Node.js backend. It does **not** connect to PostgreSQL.

## Prerequisites

- Go 1.22 or newer
- PostgreSQL running with `seed_db/` applied
- Node.js backend running on port `5007` (default)

## Setup

```bash
cd go-service
cp .env.example .env
# edit .env if your Node API URL or credentials differ
go mod tidy
```

## Run

```bash
go run .
```

The service listens on `http://localhost:8080` by default.

## Endpoint

```
GET /api/v1/students/:id/report
```

Returns `application/pdf` with `Content-Disposition: attachment`.

### Example

```bash
# Insert a student first if seed data has none, then:
curl -D - -o student-report.pdf http://localhost:8080/api/v1/students/2/report
```

Open `student-report.pdf` and confirm the student name and details appear.

## Tests

```bash
go test ./...
```

## Environment

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Go service listen port |
| `NODE_API_URL` | `http://localhost:5007` | Node backend base URL |
| `NODE_API_USERNAME` | `admin@school-admin.com` | Login username for Node API |
| `NODE_API_PASSWORD` | `3OU4zn3q6Zh9` | Login password for Node API |
