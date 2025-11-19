# Go Migration Game Plan - Complete Feature Parity with Rust

## Current State (✅ Completed)

**Phase 1**: Core Domain Models (100% coverage)
- ✅ Data models (Idea, Telos, Analysis)
- ✅ Scoring engine (exact Rust parity, 85.6% coverage)
- ✅ Pattern detector (98.1% coverage)
- ✅ Telos parser (93.8% coverage)

**Phase 2**: CLI with Cobra (7 commands)
- ✅ dump, score, review, analyze, prune, analytics, link commands
- ⚠️ 0% test coverage on CLI commands (needs integration tests)

**Phase 3**: RESTful API Server (NEW - not in Rust!)
- ✅ 8 HTTP endpoints with Chi router (85.3% coverage)
- ✅ OpenAPI documentation
- ✅ CORS, middleware, graceful shutdown

**Infrastructure**:
- ✅ Docker configuration (Go only)
- ✅ SQLite database with repository pattern
- ⚠️ 0% test coverage on database layer beyond basic repository tests
- ⚠️ 0% test coverage on config layer

---

## Remaining Work - Feature Parity with Rust

### Phase 4: Production Infrastructure (HIGH PRIORITY)
**Objective**: Make Go implementation production-ready with observability

**Estimated Effort**: 24-32 hours | **Parallelization**: 3 concurrent tracks

#### Track 4A: Health Monitoring System (8-10 hours) 🔴
**TDD Subagent Task**: `internal/health/`

**Red Phase** - Write tests first:
```go
// internal/health/monitor_test.go
- TestHealthMonitor_AddCheck
- TestHealthMonitor_RunAllChecks
- TestHealthCheckState_Aggregation (healthy/degraded/unhealthy)
- TestDatabaseHealthChecker
- TestMemoryHealthChecker
- TestDiskSpaceHealthChecker
- TestUptime_Calculation
```

**Green Phase** - Implement:
- `internal/health/monitor.go` - Core health monitoring
  - HealthMonitor struct with check registration
  - HealthChecker interface (Name, Check methods)
  - HealthStatus aggregation (healthy/degraded/unhealthy)
- `internal/health/checkers.go` - Built-in checkers
  - DatabaseHealthChecker (connection pool stats)
  - MemoryHealthChecker (RSS, heap usage)
  - DiskSpaceHealthChecker (data dir available space)
- `internal/health/uptime.go` - Uptime tracking
  - Start time recording
  - Duration calculation

**Refactor Phase**:
- Extract common health check patterns
- Add configurable thresholds (memory %, disk %)
- Optimize check execution (parallel health checks)

**Integration**:
- Add `/health` endpoint to API server
- Add `tm health` CLI command
- Wire into graceful shutdown

**Success Criteria**:
- ✅ >90% test coverage
- ✅ Health checks complete in <100ms
- ✅ Proper state aggregation (any error = unhealthy)
- ✅ Matches Rust src/health.rs behavior

---

#### Track 4B: Structured Logging & Metrics (8-10 hours) 🟢
**TDD Subagent Task**: `internal/logging/` and `internal/metrics/`

**Red Phase** - Write tests first:
```go
// internal/logging/logger_test.go
- TestLogger_NewLogger_WithEnvironment
- TestLogger_Levels (debug/info/warn/error)
- TestLogger_StructuredFields
- TestLogger_FileRotation
- TestLogger_JsonOutput

// internal/metrics/collector_test.go
- TestMetricsCollector_RecordCounter
- TestMetricsCollector_RecordGauge
- TestMetricsCollector_RecordHistogram
- TestMetricsCollector_GetSnapshot
```

**Green Phase** - Implement:
- `internal/logging/logger.go` - Structured logger
  - Use zerolog or zap for structured logging
  - Log levels (Debug, Info, Warn, Error)
  - Context-aware logging (request ID, user ID, etc.)
  - File rotation support (logs directory)
- `internal/logging/middleware.go` - HTTP request logging
  - Request/response logging
  - Duration tracking
  - Error logging
- `internal/metrics/collector.go` - Metrics collection
  - Counter, Gauge, Histogram types
  - In-memory metrics storage
  - Prometheus-compatible format (future)
- `internal/metrics/metrics.go` - Application metrics
  - Ideas created/updated/deleted counters
  - Scoring duration histogram
  - Database query duration histogram
  - LLM API call counter (for Phase 5)

**Refactor Phase**:
- Add metrics middleware for API
- Optimize log levels for production
- Add sampling for high-volume logs

**Integration**:
- Replace fmt.Println with structured logging
- Add `tm logs tail` CLI command
- Add `/metrics` endpoint to API

**Success Criteria**:
- ✅ >85% test coverage
- ✅ JSON log output
- ✅ Log rotation working (max 10MB per file, 7 days retention)
- ✅ Metrics endpoint returns Prometheus format

---

#### Track 4C: Background Task Manager (8-12 hours) 🔵
**TDD Subagent Task**: `internal/tasks/`

**Red Phase** - Write tests first:
```go
// internal/tasks/manager_test.go
- TestTaskManager_SpawnTask
- TestTaskManager_GracefulShutdown
- TestTaskManager_TaskCompletion
- TestTaskManager_TaskFailure
- TestTaskManager_ConcurrentTasks
- TestTaskManager_ShutdownTimeout
```

**Green Phase** - Implement:
- `internal/tasks/manager.go` - Task supervision
  - TaskManager struct with JoinSet
  - Spawn supervised tasks
  - Graceful shutdown coordination
  - Signal handling (SIGTERM, SIGINT)
- `internal/tasks/task.go` - Task abstraction
  - Task interface (Run, Name, Timeout methods)
  - Task result tracking
  - Error recovery
- `internal/tasks/scheduler.go` - Scheduled tasks (cron-like)
  - Periodic task execution
  - Cleanup tasks (prune old ideas)
  - Health check scheduling

**Refactor Phase**:
- Add task priority levels
- Add task retry policies
- Extract signal handling utilities

**Integration**:
- Wire into API server startup
- Add database cleanup task (weekly prune)
- Add metrics collection task (hourly snapshot)

**Success Criteria**:
- ✅ >85% test coverage
- ✅ Graceful shutdown in <5 seconds
- ✅ Tasks properly canceled on shutdown
- ✅ No goroutine leaks

**Dependencies**: None (can run in parallel with 4A, 4B)

---

### Phase 5: LLM Integration (CRITICAL - HIGHEST VALUE)
**Objective**: Port AI-powered analysis from Rust to Go

**Estimated Effort**: 32-40 hours | **Parallelization**: 4 concurrent tracks

#### Track 5A: Ollama Client & Provider Abstraction (10-12 hours) 🔴
**TDD Subagent Task**: `internal/llm/client/`

**Red Phase** - Write tests first:
```go
// internal/llm/client/ollama_test.go
- TestOllamaClient_Generate
- TestOllamaClient_Timeout
- TestOllamaClient_ConnectionError
- TestOllamaClient_ModelNotFound
- TestOllamaClient_Streaming (optional)

// internal/llm/provider_test.go
- TestProvider_Analyze
- TestProvider_FallbackChain
- TestClaudeProvider_Analyze
```

**Green Phase** - Implement:
- `internal/llm/client/ollama.go` - Ollama HTTP client
  - Generate API call
  - Model listing
  - Service health check
  - Configurable timeout (30s default)
- `internal/llm/provider.go` - Provider interface
  - Provider interface (Analyze, Name, IsAvailable methods)
  - OllamaProvider implementation
  - ClaudeProvider stub (for future Claude API)
  - FallbackProvider chain (Ollama → Claude → rule-based)
- `internal/llm/prompts.go` - Prompt templates
  - Idea analysis prompt
  - Score extraction prompt
  - Explanation generation prompt

**Refactor Phase**:
- Extract HTTP client configuration
- Add streaming support for real-time feedback
- Optimize prompt templates

**Success Criteria**:
- ✅ >85% test coverage
- ✅ Works with Ollama running locally
- ✅ Proper timeout handling
- ✅ Graceful fallback on connection failure

---

#### Track 5B: Semantic Cache System (10-12 hours) 🟢
**TDD Subagent Task**: `internal/llm/cache/`

**Red Phase** - Write tests first:
```go
// internal/llm/cache/cache_test.go
- TestCache_StoreAndRetrieve
- TestCache_SimilarityMatching (>0.85 threshold)
- TestCache_TTL_Expiration (24 hours)
- TestCache_MaxSize_Eviction (1000 entries)
- TestCache_HitCount_Tracking
- TestCache_Stats
- TestNormalizeIdea_Canonicalization
```

**Green Phase** - Implement:
- `internal/llm/cache/cache.go` - Semantic cache
  - LlmCacheEntry struct (analysis, normalized text, provider, timestamp, hit count)
  - In-memory cache with sync.RWMutex
  - TTL-based expiration (24 hours default)
  - LRU eviction when max size reached
- `internal/llm/cache/similarity.go` - Text similarity
  - Jaccard similarity for quick matching
  - Cosine similarity (optional, for better results)
  - Normalize idea text (lowercase, remove punctuation, trim)
  - Configurable similarity threshold (0.85 default)
- `internal/llm/cache/stats.go` - Cache statistics
  - Hit rate calculation
  - Average similarity scores
  - Cache size monitoring

**Refactor Phase**:
- Add cache persistence (save to disk on shutdown)
- Optimize similarity calculation (skip low-similarity candidates early)
- Add cache warming (preload common queries)

**Success Criteria**:
- ✅ >90% test coverage
- ✅ Cache hits return in <5ms
- ✅ Similarity matching accuracy >90%
- ✅ Proper LRU eviction

**Dependencies**: 5A (needs LlmAnalysisResult types)

---

#### Track 5C: Quality Metrics & Response Processing (8-10 hours) 🔵
**TDD Subagent Task**: `internal/llm/quality/`

**Red Phase** - Write tests first:
```go
// internal/llm/quality/tracker_test.go
- TestQualityTracker_ScoreResponse
- TestQualityTracker_ConfidenceLevel
- TestQualityTracker_ConsistencyCheck
- TestQualityTracker_GetAverageQuality

// internal/llm/processing/processor_test.go
- TestProcessor_ParseLlmResponse
- TestProcessor_ExtractScores
- TestProcessor_HandleMalformedResponse
- TestProcessor_FallbackToRuleBased
```

**Green Phase** - Implement:
- `internal/llm/quality/tracker.go` - Quality tracking
  - QualityMetrics struct (completeness, consistency, confidence)
  - Score response quality (0.0-1.0 scale)
  - Track quality over time
  - Generate quality reports
- `internal/llm/quality/metrics.go` - Metric calculations
  - Completeness: Are all scoring dimensions present?
  - Consistency: Do scores match explanations?
  - Confidence: Is LLM confident in analysis?
- `internal/llm/processing/processor.go` - Response processing
  - Parse LLM JSON response
  - Extract scores and explanations
  - Validate score ranges (0.0-10.0)
  - Fallback to rule-based scoring on failure
- `internal/llm/processing/validator.go` - Response validation
  - Schema validation
  - Score range validation
  - Explanation presence validation

**Refactor Phase**:
- Add quality thresholds (reject low-quality responses)
- Optimize parsing for different LLM response formats
- Add retries on validation failure

**Success Criteria**:
- ✅ >85% test coverage
- ✅ Handles malformed responses gracefully
- ✅ Quality scoring accurate (validated against manual review)
- ✅ Fallback to rule-based scoring works

**Dependencies**: 5A (needs provider types)

---

#### Track 5D: LLM CLI Commands & Service Management (6-8 hours) 🟡
**TDD Subagent Task**: `internal/cli/llm.go` and `internal/llm/service/`

**Red Phase** - Write tests first:
```go
// internal/llm/service/manager_test.go
- TestServiceManager_StartOllama
- TestServiceManager_StopOllama
- TestServiceManager_CheckStatus
- TestServiceManager_ListModels

// internal/cli/llm_test.go (integration)
- TestLlmCommand_Status
- TestLlmCommand_Start
- TestLlmCommand_Stop
- TestLlmCommand_Models
```

**Green Phase** - Implement:
- `internal/llm/service/manager.go` - Ollama service management
  - Start Ollama (exec command)
  - Stop Ollama (SIGTERM, wait, SIGKILL if needed)
  - Check status (HTTP ping)
  - List available models
- `internal/cli/llm.go` - LLM CLI commands
  - `tm llm status` - Check Ollama status
  - `tm llm start` - Start Ollama service
  - `tm llm stop` - Stop Ollama service
  - `tm llm models` - List available models
- `internal/cli/analyze_llm.go` - AI-powered analysis
  - `tm analyze --ai <idea>` - Use LLM for analysis
  - `tm dump --ai <idea>` - Analyze and save with LLM
  - Cache integration
  - Quality metrics display

**Refactor Phase**:
- Add Ollama auto-start on first use
- Optimize model download/pull
- Add model recommendation based on hardware

**Integration**:
- Add `--ai` flag to existing commands
- Wire cache into analysis flow
- Display cache hit/miss in output

**Success Criteria**:
- ✅ >75% test coverage (lower due to exec/system calls)
- ✅ Service management works on Linux/macOS
- ✅ Graceful handling of missing Ollama
- ✅ Cache integration reduces API calls by >60%

**Dependencies**: 5A, 5B, 5C

---

### Phase 6: Advanced CLI Features (MEDIUM PRIORITY)
**Objective**: Feature parity with advanced Rust CLI commands

**Estimated Effort**: 16-20 hours | **Parallelization**: 2 concurrent tracks

#### Track 6A: Bulk Operations (10-12 hours) 🔴
**TDD Subagent Task**: `internal/cli/bulk.go`

**Red Phase** - Write tests first:
```go
// internal/cli/bulk_test.go
- TestBulkTag_WithFilters
- TestBulkArchive_WithFilters
- TestBulkDelete_WithConfirmation
- TestBulkImport_FromCSV
- TestBulkExport_ToCSV
- TestBulkExport_ToJSON
```

**Green Phase** - Implement:
- `internal/cli/bulk.go` - Bulk operations
  - `tm bulk tag <tag> --min-score 7.0` - Tag multiple ideas
  - `tm bulk archive --older-than 90 --max-score 5.0` - Archive old low-scoring ideas
  - `tm bulk delete --older-than 180 --confirm` - Delete old ideas
- `internal/export/csv.go` - CSV import/export
  - Export ideas to CSV (id, content, score, status, created_at)
  - Import ideas from CSV
  - Validation and duplicate detection
- `internal/export/json.go` - JSON export
  - Export full idea data (including analysis)
  - Export multiple ideas as array
  - Pretty-print option

**Refactor Phase**:
- Add progress bars for bulk operations
- Optimize bulk database operations (batch inserts/updates)
- Add undo functionality (export backup before bulk delete)

**Success Criteria**:
- ✅ >85% test coverage
- ✅ Confirmation prompts prevent accidental deletion
- ✅ CSV import handles 1000+ rows
- ✅ Matches Rust src/commands/bulk.rs functionality

---

#### Track 6B: Enhanced Analytics (6-8 hours) 🟢
**TDD Subagent Task**: `internal/cli/analytics_enhanced.go`

**Red Phase** - Write tests first:
```go
// internal/cli/analytics_enhanced_test.go
- TestAnalytics_TrendOverTime
- TestAnalytics_PatternFrequency
- TestAnalytics_ScoreDistribution
- TestAnalytics_TopIdeas
- TestAnalytics_ExportReport
```

**Green Phase** - Implement:
- Enhance `internal/cli/analytics.go`
  - `tm analytics trends --days 30` - Score trends over time
  - `tm analytics patterns` - Pattern frequency analysis
  - `tm analytics top --limit 10` - Top scoring ideas
  - `tm analytics report` - Generate comprehensive report
- `internal/analytics/trends.go` - Trend analysis
  - Score trends by week/month
  - Pattern detection trends
  - Idea creation rate trends
- `internal/analytics/reports.go` - Report generation
  - Summary statistics
  - Charts (ASCII art for CLI)
  - Export to Markdown

**Refactor Phase**:
- Add data visualization (ASCII charts)
- Optimize queries for large datasets
- Add caching for expensive analytics

**Success Criteria**:
- ✅ >80% test coverage
- ✅ Trends calculated correctly
- ✅ Reports readable and actionable
- ✅ Performance <1s for 10,000 ideas

**Dependencies**: None (can run in parallel with 6A)

---

### Phase 7: Database Resilience (MEDIUM PRIORITY)
**Objective**: Make database layer production-ready

**Estimated Effort**: 12-16 hours | **Single Track** (sequential)

#### Track 7: Enhanced Database Layer (12-16 hours) 🔴
**TDD Subagent Task**: `internal/database/` (enhance existing)

**Red Phase** - Write tests first:
```go
// internal/database/repository_test.go (expand existing)
- TestRepository_ConnectionPool_Config
- TestRepository_Retry_OnTransientError
- TestRepository_Metrics_QueryDuration
- TestRepository_HealthCheck
- TestRepository_ConcurrentAccess (1000 goroutines)

// internal/database/migrations_test.go
- TestMigrations_Apply
- TestMigrations_Rollback
- TestMigrations_Version
```

**Green Phase** - Implement:
- Enhance `internal/database/repository.go`
  - Connection pooling with configurable limits
  - Retry logic with exponential backoff (3 retries, 100ms-1s delays)
  - Query duration metrics
  - Transaction support with context
  - Prepared statement caching
- `internal/database/migrations.go` - Migration system
  - Versioned migrations (up/down)
  - Migration status tracking
  - Safe rollback
- `internal/database/health.go` - Database health checks
  - Connection pool stats (active/idle connections)
  - Query response time
  - Disk space usage
- `internal/database/relationships.go` - Idea relationships
  - IdeaRelationship struct (depends-on, related-to, blocks, etc.)
  - CRUD for relationships
  - Relationship graph queries

**Refactor Phase**:
- Extract connection pool configuration
- Optimize prepared statement usage
- Add connection pool monitoring

**Success Criteria**:
- ✅ >90% test coverage
- ✅ Handles 1000 concurrent connections
- ✅ Retries work on transient errors (locked database)
- ✅ Migrations reversible
- ✅ Matches Rust src/database_simple.rs resilience

---

### Phase 8: Polish & Documentation (LOW PRIORITY)
**Objective**: Final touches for production readiness

**Estimated Effort**: 8-12 hours | **Parallelization**: 3 concurrent tracks

#### Track 8A: Enhanced Telos Parsing (4-5 hours) 🔴
**TDD Subagent Task**: `internal/telos/parser.go` (enhance existing)

**Red Phase** - Write tests first:
```go
// internal/telos/parser_test.go (expand)
- TestParser_ProblemsSection
- TestParser_MissionsSection
- TestParser_ChallengesSection
- TestParser_FullTelosFile
```

**Green Phase** - Implement:
- Enhance `internal/telos/parser.go`
  - Parse Problems section (P1, P2, etc.)
  - Parse Missions section (M1, M2, etc.)
  - Parse Challenges section (C1, C2, etc.)
  - Full telos.md spec support

**Success Criteria**:
- ✅ >95% test coverage
- ✅ Matches Rust src/telos.rs parsing

---

#### Track 8B: Utilities & Helpers (2-3 hours) 🟢
**TDD Subagent Task**: `internal/utils/`

**Red Phase** - Write tests first:
```go
// internal/utils/clipboard_test.go
- TestClipboard_Copy
- TestClipboard_Paste
- TestClipboard_UnavailableHandler
```

**Green Phase** - Implement:
- `internal/utils/clipboard.go` - Clipboard integration
  - Copy idea to clipboard
  - Paste from clipboard for quick capture
  - Cross-platform support (Linux/macOS/Windows)

**Success Criteria**:
- ✅ >80% test coverage
- ✅ Works on Linux and macOS

---

#### Track 8C: Testing & Documentation (4-5 hours) 🔵
**Non-TDD Task**: Documentation and integration tests

**Tasks**:
1. Add integration tests for CLI commands (currently 0% coverage)
   ```go
   // internal/cli/integration_test.go
   - TestDumpCommand_Integration
   - TestReviewCommand_Integration
   - TestPruneCommand_Integration
   ```
2. Add integration tests for config loading (currently 0% coverage)
   ```go
   // internal/config/config_test.go
   - TestConfig_LoadFromFile
   - TestConfig_LoadFromEnv
   - TestConfig_Defaults
   ```
3. Update README.md
   - Add Go vs Rust feature comparison table
   - Document LLM setup (Ollama installation)
   - Update API documentation
4. Create MIGRATION.md
   - Document migration process from Rust to Go
   - Database compatibility notes
   - Feature parity checklist
5. Update Docker documentation
   - Multi-stage build explanation
   - Environment variables
   - Volume mounts for data persistence

**Success Criteria**:
- ✅ CLI tests achieve >70% coverage
- ✅ Config tests achieve >85% coverage
- ✅ Documentation accurate and complete
- ✅ Migration guide tested by fresh user

**Dependencies**: None

---

## Execution Plan - Parallel Subagent Workflow

### Sprint 1: Production Infrastructure (Week 1)
**Parallel Execution** - 3 subagents:
- Subagent A: Track 4A (Health Monitoring)
- Subagent B: Track 4B (Logging & Metrics)
- Subagent C: Track 4C (Background Tasks)

**Deliverables**:
- Production-ready infrastructure
- Observability (health, logs, metrics)
- Graceful shutdown

**Validation**:
- Run all tests: `go test ./... -cover`
- Manual testing: Start API, check `/health`, trigger shutdown
- Load testing: 1000 concurrent requests

---

### Sprint 2: LLM Integration Part 1 (Week 2)
**Parallel Execution** - 3 subagents:
- Subagent A: Track 5A (Ollama Client)
- Subagent B: Track 5B (Semantic Cache)
- Subagent C: Track 5C (Quality Metrics)

**Deliverables**:
- Working Ollama integration
- Intelligent caching (60%+ hit rate)
- Quality tracking

**Validation**:
- Test with real Ollama: `tm analyze --ai "Build a tool"`
- Verify cache hits on similar ideas
- Check quality scores are reasonable

---

### Sprint 3: LLM Integration Part 2 + Advanced CLI (Week 3)
**Parallel Execution** - 3 subagents:
- Subagent A: Track 5D (LLM CLI Commands)
- Subagent B: Track 6A (Bulk Operations)
- Subagent C: Track 6B (Enhanced Analytics)

**Deliverables**:
- Complete LLM CLI integration
- Bulk operations (tag, archive, delete, import/export)
- Advanced analytics

**Validation**:
- `tm llm status` shows Ollama running
- `tm bulk archive --older-than 90 --dry-run` shows candidates
- `tm analytics trends --days 30` displays trends

---

### Sprint 4: Database + Polish (Week 4)
**Parallel Execution** - 4 subagents:
- Subagent A: Track 7 (Database Resilience)
- Subagent B: Track 8A (Enhanced Telos Parsing)
- Subagent C: Track 8B (Utilities)
- Subagent D: Track 8C (Testing & Documentation)

**Deliverables**:
- Resilient database layer
- Full telos.md support
- Clipboard integration
- Complete documentation
- >85% overall test coverage

**Validation**:
- Full integration test suite passes
- Documentation reviewed and tested
- Performance benchmarks met
- Ready for production deployment

---

## Final Validation Checklist

Before declaring feature parity complete:

### Functional Parity
- [ ] All 11+ Rust CLI commands replicated in Go
- [ ] LLM integration working (Ollama + cache + quality)
- [ ] Bulk operations match Rust behavior
- [ ] Analytics reports accurate
- [ ] Health monitoring operational
- [ ] Database handles 10,000+ ideas

### Performance
- [ ] Scoring: <50ms per idea
- [ ] LLM cache hits: <5ms
- [ ] API endpoints: <100ms p95
- [ ] Bulk operations: >100 ideas/second
- [ ] Database queries: <10ms p95

### Quality
- [ ] Overall test coverage: >85%
- [ ] Critical paths: >95% coverage
- [ ] No flaky tests
- [ ] CI/CD passing on all PRs
- [ ] Memory usage: <100MB idle, <500MB under load

### Production Readiness
- [ ] Structured logging enabled
- [ ] Metrics collection working
- [ ] Health checks responsive
- [ ] Graceful shutdown tested
- [ ] Docker image builds successfully
- [ ] Documentation complete

### Migration Safety
- [ ] Database schema compatible with Rust
- [ ] Scoring results match Rust (±0.1 points)
- [ ] Pattern detection parity verified
- [ ] Can run Go and Rust side-by-side

---

## Post-Parity Next Steps

Once feature parity achieved:

### Option 1: Prune Rust (Recommended)
1. Archive Rust codebase to `legacy/` directory
2. Update CI/CD to remove Rust workflows
3. Update README to remove Rust references
4. Keep Rust as reference documentation
5. Maintain database compatibility for migration

### Option 2: Hybrid Approach
1. Keep both implementations
2. Document use cases for each
3. Share SQLite database
4. Cross-validate scoring results
5. Use Rust as regression test suite

### Option 3: Rust Special Features
1. Keep Rust for advanced LLM features
2. Use Go for API/web serving
3. Share database and telos.md
4. Document which to use when

---

## Estimated Timeline

**Total Effort**: 92-120 hours

**With 4 Parallel Subagents**:
- Sprint 1: 8-10 hours (Phase 4)
- Sprint 2: 10-12 hours (Phase 5 Part 1)
- Sprint 3: 8-10 hours (Phase 5 Part 2 + Phase 6)
- Sprint 4: 12-16 hours (Phase 7 + Phase 8)

**Calendar Time**: 4 weeks (assuming 8-10 hours per week per subagent)

**Fast Track** (6 subagents, 16 hours/week): 2-3 weeks

---

## Risk Mitigation

### Technical Risks
1. **LLM API changes**: Mock Ollama in tests, versioned API
2. **Cache complexity**: Start with simple Jaccard similarity, iterate
3. **Database concurrency**: Use transactions, test with race detector
4. **Service management**: Platform-specific code isolated, well-tested

### Process Risks
1. **Scope creep**: Stick to Rust parity, no new features
2. **Testing burden**: TDD enforced, no code without tests
3. **Integration complexity**: Integration tests in each sprint
4. **Documentation lag**: Update docs in parallel with code

---

## Success Metrics

**Code Quality**:
- Test coverage: >85% overall, >95% critical paths
- No critical bugs in production
- CI/CD green on all commits

**Performance**:
- API latency: <100ms p95
- LLM cache hit rate: >60%
- Memory usage: <500MB under load

**Feature Completeness**:
- All Rust CLI commands: ✅
- LLM integration: ✅
- Bulk operations: ✅
- Production infrastructure: ✅

**Migration Readiness**:
- Documentation complete: ✅
- Migration guide tested: ✅
- Rust can be safely archived: ✅

---

## Communication Plan

**Daily**: Subagent status updates (automated via CI)
**Weekly**: Sprint retrospective, plan next sprint
**End of Phase**: Demo completed features, validate against Rust

**Escalation**: If any track blocks >2 days, re-evaluate dependencies or split tasks

---

Ready to execute! 🚀
