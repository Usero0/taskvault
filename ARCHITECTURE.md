# TaskVault Architecture & Design

## System Overview

TaskVault is a **content-aware cache layer** for deterministic task execution. It stores task outputs keyed by cryptographic hash of inputs, enabling instant result retrieval for identical or re-executed work.

### Core Principles

1. **Content-Addressable Storage**: Hashes based on actual data, not parameters
2. **Format Agnostic**: Works with any binary output (JSON, ML models, images, etc.)
3. **Transparent Integration**: CLI wrappers, SDK, or programmatic
4. **Distributed Ready**: Single-node SQLite scales to PostgreSQL + object storage
5. **Resilient**: Corruption detection, atomic writes, audit trails

---

## Component Architecture

### 1. Hash Engine (`internal/hash/engine.go`)

**Purpose**: Compute content fingerprints for determinism.

**Algorithms**:
- **Blake3** (default): 256-bit cryptographic hash, SIMD acceleration, fastest for large files
- **SHA256**: Standard, universal, slightly slower but compatible

**Key Functions**:
- `HashData([]byte)` → hex string hash
- `HashFile(path)` → hash of file contents
- `HashDirectory(path)` → composite hash of tree (for detecting structure changes)

**Properties**:
- Deterministic: same input → same hash
- Collision-resistant: different input → different hash (crypto-grade)
- Fast: Blake3 at 3+ GB/s on modern CPUs

---

### 2. Storage Layer (`internal/storage/store.go`)

**Purpose**: Persistent cache with SQLite metadata + blob storage.

**Architecture**:
```
Cache Directory/
├── cache.db               # SQLite: metadata, indexes, TTL
└── blobs/                 # Flat directory: hash → blob file
    ├── a3f2b1c8d5e...    # Content-addressed blobs
    └── ...
```

**Schema**:
```sql
CREATE TABLE cache_entries (
    hash TEXT PRIMARY KEY,          -- Blake3/SHA256 hash
    metadata TEXT,                  -- JSON metadata
    created_at TIMESTAMP,
    accessed_at TIMESTAMP,          -- For LRU eviction
    expires_at TIMESTAMP,           -- For TTL cleanup
    size INTEGER,                   -- Bytes
    blob_path TEXT                  -- Path to actual data
);

CREATE INDEX idx_accessed ON cache_entries(accessed_at);
CREATE INDEX idx_expires ON cache_entries(expires_at);
```

**Operations**:
- `Set(entry)`: Write blob + insert metadata
- `Get(hash)`: Retrieve blob + update accessed_at
- `Delete(hash)`: Remove entry and blob file
- `evictIfNeeded()`: LRU cleanup when cache exceeds limit

**Concurrency**: All DB operations serialized; filesystem operations atomic per entry.

**Resilience**: 
- Missing blob → cache miss (not error)
- Corrupted blob → auto-delete entry
- DB corruption → recoverable from blobs

---

### 3. Cache Manager (`internal/cache/manager.go`)

**Purpose**: Orchestrate cache operations, policies, and audit logging.

**Key Responsibilities**:
- Coordinate between hasher, storage, and audit logger
- Apply eviction policies (per-task TTL/max size)
- Track hit/miss metrics
- Thread-safe concurrent access

**Public API**:
```go
SaveResult(taskName, input, output, metadata) → cacheKey
GetResult(taskName, input) → (output, metadata, hit, error)
InvalidateTask(taskName) → evicted_count
GetStats() → map[stats]
```

**Policies** (registered per task):
```
type EvictionPolicy struct {
    Name     string        // e.g., "ml_training"
    TTL      time.Duration // e.g., 30*24*time.Hour
    MaxSize  int64         // e.g., 5 GB
    Strategy string        // "lru", "lfu", "fifo"
}
```

---

### 4. Audit Logger (`internal/audit/logger.go`)

**Purpose**: Immutable record of all cache operations.

**Log Format** (line-per-operation):
```
[2025-02-02T14:30:45.123456Z] HIT get task=model_training hash=a3f2b1c8d5e...
[2025-02-02T14:30:46.234567Z] MISS get task=data_pipeline hash=c8d9f0e1a2b...
[2025-02-02T14:30:47.345678Z] ERROR hash_error task=invalid_input error=EOF
```

**Usage**:
```bash
# Find hit rate for specific task
grep "task=ml_training" audit.log | grep -c "HIT" / grep -c "MISS"

# Monitor errors
grep "ERROR" audit.log | tail -20

# Export to analytics
cat audit.log | awk '{print $2, $3}' > cache_ops.csv
```

---

### 5. Configuration (`internal/config/config.go`)

**YAML Format**:
```yaml
cache_dir: ~/.taskvault/cache
max_size_gb: 50
hash_algorithm: blake3
policies:
  default:
    ttl_seconds: 604800
    max_size_bytes: 104857600
    strategy: lru
  ml_training:
    ttl_seconds: 2592000
    max_size_bytes: 5368709120
```

**Validation**:
- `cache_dir` required
- `max_size_gb` >= 1
- `hash_algorithm` ∈ {blake3, sha256}

---

## Data Flow

### Write Path (SaveResult)
```
Input Data (bytes)
    ↓
[HashEngine] → Blake3 Hash
    ↓
[Manager] → Check if exists
    ↓
Exists? → Return cache key (no-op)
Not exists? ↓
[Storage] → Write blob to disk
    ↓
[Storage] → Insert metadata to DB
    ↓
[Manager] → Check eviction policy
    ↓
[Storage] → evictIfNeeded()
    ↓
[AuditLogger] → Log SAVE operation
    ↓
Return cache key
```

### Read Path (GetResult)
```
Input Data (bytes) + TaskName
    ↓
[HashEngine] → Blake3 Hash
    ↓
[Manager] → Check cache
    ↓
[Storage] → Query by hash
    ↓
Found? ↓ (No) Return MISS
    ↓ (Yes)
[Storage] → Read blob from disk
    ↓
Blob exists? → Yes: Return data
             → No: Delete metadata, return MISS
    ↓
[Storage] → Update accessed_at
    ↓
[AuditLogger] → Log HIT
    ↓
Return (output, metadata, true)
```

---

## Eviction Strategy

### When Eviction Triggers
- Cache total size > `max_size_gb`

### Algorithm
1. **Scan expired entries**: Remove anything with `expires_at < now`
2. **LRU eviction**: Sort by `accessed_at`, remove oldest until `usage < 50% of limit`
3. **Conservative**: Never evict below 50% to prevent thrashing

### Example
```
Max cache: 10 GB
Current usage: 12 GB (triggered)
Target: 5 GB (50% of max)

Evict entries sorted by accessed_at until 5 GB remaining
```

---

## Concurrency Model

All operations protected by `sync.RWMutex`:
- **Read lock** held during `Get()` operations (readers don't block each other)
- **Write lock** held during `SaveResult()` (exclusive)
- **Eviction** runs under write lock (atomic cleanup)

**Thread Safety**:
- Multiple goroutines can `Get()` simultaneously ✓
- Single `SaveResult()` at a time (serialized) ✓
- Audit log is mutex-protected ✓
- SQLite connections pool-managed ✓

---

## Error Handling

| Scenario | Behavior |
|---|---|
| **Blob missing, metadata exists** | Delete metadata, return MISS (not error) |
| **Hash computation fails** | Log error, return error to caller |
| **Database corrupt** | Attempt recovery from audit log, fail hard if unrecoverable |
| **Disk full** | Evict aggressively, error if can't free space |
| **Concurrent access collision** | SQLite serializes writes, read lock prevents thrashing |

---

## Performance Characteristics

| Operation | Latency | Notes |
|---|---|---|
| `HashData()` - 1MB | ~0.3ms | Blake3 SIMD optimized |
| `Get()` (cache hit) | ~2-5ms | DB lookup + file read |
| `Set()` | ~10-50ms | Depends on blob size + disk I/O |
| `GetStats()` | ~5ms | Single SQL aggregation |

**Throughput**:
- Blake3: 3+ GB/s on modern CPUs
- SQLite: 10,000+ ops/sec (single connection)
- With connection pool: 100,000+ ops/sec

---

## Future Architecture Extensions

### Distributed Sync (Q3 2025)
```
TaskVault Cluster
├── Node 1 (cache)
├── Node 2 (cache)
├── Node 3 (cache)
└── gRPC gossip protocol
    → Share hashes
    → Replicate hot entries
    → Consensus on eviction
```

### Cloud Backends (Q1 2026)
```
├── Local layer (SQLite)
├── Cloud layer (S3/GCS/Azure)
│   → Tier cold data to cloud
│   → Transparent retrieval
│   → Cost optimization
```

### ML Registry Integration (Q4 2025)
```
Integration with:
- Hugging Face Model Hub
- MLflow
- Weights & Biases
→ Auto-detect ML artifacts
→ Version tracking
→ Cross-team model sharing
```

---

## Scalability Considerations

| Scale | Recommendation |
|---|---|
| **Single machine (< 100GB cache)** | SQLite + local filesystem |
| **Team (1-10TB cache)** | PostgreSQL + NAS |
| **Enterprise (10TB+)** | PostgreSQL + S3/GCS + distributed sync |

### Single-Machine Limits
- SQLite: ~millions of entries
- Filesystem: depends on filesystem (ext4, APFS, NTFS all fine)
- Concurrency: goroutine-based scaling up to CPU count

### Multi-Node Scaling
- gRPC gossip protocol for cache distribution
- Consistent hashing for data placement
- Replication factor for reliability

---

## SOLID Principles Applied

| Principle | Implementation |
|---|---|
| **S**ingle Responsibility | Each module handles one concern (hash, storage, audit) |
| **O**pen/Closed | Extensible policies; eviction strategies pluggable |
| **L**iskov Substitution | `HashAlgorithm` interface allows swappable implementations |
| **I**nterface Segregation | Minimal interface: `Get()`, `Set()`, `Delete()` |
| **D**ependency Inversion | Config object injected; no global state |

---

## Security Considerations

### Input Validation
- Files checked for read permission before hashing
- Path traversal protection in blob storage
- Metadata JSON validated before unmarshaling

### Access Control
- File permissions preserved (644 for blobs)
- Audit log immutable once written
- No plaintext secrets in config

### Integrity
- Blake3 cryptographic hashing prevents tampering
- Blob checksums validate on read
- Audit trail detects unexpected deletions

---

## Monitoring & Observability

### Key Metrics to Track
- **Cache hit rate**: HIT / (HIT + MISS)
- **Average latency**: per operation
- **Eviction rate**: entries removed per hour
- **Total cache size**: current usage vs. limit
- **Error rate**: by operation type

### Prometheus Format (Future)
```
taskvault_cache_hits_total
taskvault_cache_misses_total
taskvault_cache_size_bytes
taskvault_evictions_total
taskvault_operation_duration_seconds
```

---

## References

- [Blake3 Spec](https://github.com/BLAKE3-team/BLAKE3-specs)
- [SQLite ACID Properties](https://www.sqlite.org/atomiccommit.html)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
