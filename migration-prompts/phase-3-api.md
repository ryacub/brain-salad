# Phase 3: API Server

**Duration:** 5-7 days
**Goal:** RESTful API server with Chi router
**Dependencies:** Phase 1 complete
**Can Run Parallel:** Phase 2

## Context

Build API server that wraps core domain logic. Enable SvelteKit frontend.

## Endpoints to Implement

```
GET    /api/v1/ideas           List ideas
POST   /api/v1/ideas           Create idea
GET    /api/v1/ideas/:id       Get idea
PUT    /api/v1/ideas/:id       Update idea
DELETE /api/v1/ideas/:id       Delete idea
POST   /api/v1/analyze         Analyze text
GET    /api/v1/analytics/stats Statistics
GET    /health                 Health check
```

## TDD Approach

Use `httptest` for handler tests:

```go
func TestCreateIdeaHandler(t *testing.T) {
	repo := setupTestDB(t)
	server := NewServer(repo, cfg)

	body := `{"title":"Test","description":"Testing"}`
	req := httptest.NewRequest("POST", "/api/v1/ideas", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	server.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
```

## Deliverables

- [ ] `internal/api/server.go` - Server setup
- [ ] `internal/api/handlers.go` - All handlers
- [ ] `internal/api/middleware/` - CORS, logging, recovery
- [ ] `cmd/web/main.go` - Entry point
- [ ] `docs/api.yaml` - OpenAPI spec
- [ ] Tests (>85% coverage)

## Validation

```bash
go build -o bin/tm-web ./cmd/web
./bin/tm-web &

curl http://localhost:8080/health
curl http://localhost:8080/api/v1/ideas
curl -X POST http://localhost:8080/api/v1/ideas -H "Content-Type: application/json" -d '{"title":"Test"}'
```

## Success Criteria

✅ All endpoints working
✅ CORS configured
✅ OpenAPI docs complete  
✅ Tests passing (>85%)
✅ Ready for frontend integration
