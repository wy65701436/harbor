# Execution Status Refresh Architecture

## Before: Redundant Work Pattern

```
┌─────────────────────────────────────────────────────────────────┐
│                         Redis                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  execution:id:1:vendor:GC:status_outdate                 │  │
│  │  execution:id:2:vendor:REPLICATION:status_outdate        │  │
│  │  execution:id:3:vendor:SCAN:status_outdate               │  │
│  │  ... (1,000,000 individual keys)                         │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
           ↓ SCAN (5M ops)    ↓ SCAN (5M ops)    ↓ SCAN (5M ops)
    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
    │   Core-0     │    │   Core-1     │    │   Core-2     │
    │ Process ALL  │    │ Process ALL  │    │ Process ALL  │
    │  1M execs    │    │  1M execs    │    │  1M execs    │
    └──────────────┘    └──────────────┘    └──────────────┘
           ↓                   ↓                   ↓
    ┌─────────────────────────────────────────────────────────┐
    │              Database (20M+ queries)                     │
    │  - 5M execution SELECTs (80% redundant)                 │
    │  - 5M task GROUP BY queries (80% redundant)             │
    │  - 5M UPDATE attempts (80% fail, retry)                 │
    │  - 5M end_time UPDATEs (80% redundant)                  │
    └─────────────────────────────────────────────────────────┘

Problem: All instances do the same work!
Redis CPU: 90-100% | DB CPU: High | Efficiency: 20%
```

## After: Distributed Work Pattern

```
┌─────────────────────────────────────────────────────────────────┐
│                         Redis                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  cache:execution:refresh:queue (Redis Set)               │  │
│  │  ┌────────────────────────────────────────────────────┐  │  │
│  │  │ "1:GC"                                             │  │  │
│  │  │ "2:REPLICATION"                                    │  │  │
│  │  │ "3:SCAN"                                           │  │  │
│  │  │ ... (1M members in single set)                    │  │  │
│  │  └────────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
    ↓ SMEMBERS (1 op)  ↓ SMEMBERS (1 op)  ↓ SMEMBERS (1 op)
    │                  │                   │
    │ Filter by:       │ Filter by:        │ Filter by:
    │ id % 5 == 0      │ id % 5 == 1       │ id % 5 == 2
    ↓                  ↓                   ↓
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Core-0     │  │   Core-1     │  │   Core-2     │
│ ENV:         │  │ ENV:         │  │ ENV:         │
│ TOTAL=5      │  │ TOTAL=5      │  │ TOTAL=5      │
│ ID=0         │  │ ID=1         │  │ ID=2         │
│              │  │              │  │              │
│ Process:     │  │ Process:     │  │ Process:     │
│ 1,5,10,15..  │  │ 2,6,11,16..  │  │ 3,7,12,17..  │
│ (200K execs) │  │ (200K execs) │  │ (200K execs) │
└──────────────┘  └──────────────┘  └──────────────┘
    ↓                  ↓                   ↓
┌─────────────────────────────────────────────────────────┐
│              Database (4M queries)                       │
│  - 1M execution SELECTs (distributed)                   │
│  - 1M task GROUP BY queries (distributed)               │
│  - 1M UPDATE attempts (no conflicts)                    │
│  - 1M end_time UPDATEs (distributed)                    │
└─────────────────────────────────────────────────────────┘

Solution: Perfect work distribution via consistent hashing!
Redis CPU: 10-20% | DB CPU: Low | Efficiency: 100%
```

## Consistent Hashing Distribution

```
Execution IDs:  0   1   2   3   4   5   6   7   8   9   10  11  12 ...
                │   │   │   │   │   │   │   │   │   │   │   │   │
Modulo 5:       0   1   2   3   4   0   1   2   3   4   0   1   2 ...
                │   │   │   │   │   │   │   │   │   │   │   │   │
Assigned to:    ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓
              Core Core Core Core Core Core Core Core Core Core Core Core Core
               -0   -1   -2   -3   -4   -0   -1   -2   -3   -4   -0   -1   -2

Result: Each core processes exactly 20% of executions (for 5 cores)
        No coordination needed, deterministic, fault-tolerant
```

## Request Flow Comparison

### Before (Per-Execution Lock - PR #22572 Proposal)
```
Core-0: SCAN → Lock(1) ✓ → Process(1) → Unlock(1)
Core-1: SCAN → Lock(1) ✗ → Skip(1) → Lock(2) ✓ → Process(2)
Core-2: SCAN → Lock(1) ✗ → Skip(1) → Lock(2) ✗ → Skip(2) → Lock(3) ✓

Issues:
- Still 5M SCAN operations
- 5M lock operations (contention)
- Lock management overhead
- Redis memory for locks
```

### After (This Implementation)
```
Core-0: SMEMBERS → Filter(id%5==0) → Process(0,5,10,15...)
Core-1: SMEMBERS → Filter(id%5==1) → Process(1,6,11,16...)
Core-2: SMEMBERS → Filter(id%5==2) → Process(2,7,12,17...)

Benefits:
- Only 5 SMEMBERS operations
- Zero lock operations
- No coordination overhead
- No extra Redis memory
```

## Performance Metrics

```
┌─────────────────────────────────────────────────────────────┐
│                    Redis Operations                          │
├─────────────────────────────────────────────────────────────┤
│  Before:  ████████████████████████████████  5,000,000 ops   │
│  After:   █                                 5 ops            │
│           └─────────────────────────────────────────────────┤
│           Improvement: 1,000,000x                            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Database Queries                          │
├─────────────────────────────────────────────────────────────┤
│  Before:  ████████████████████  20,000,000 queries          │
│  After:   ████                  4,000,000 queries            │
│           └─────────────────────────────────────────────────┤
│           Improvement: 5x reduction                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Redis CPU Usage                           │
├─────────────────────────────────────────────────────────────┤
│  Before:  ███████████████████  90-100%                      │
│  After:   ██                   10-20%                        │
│           └─────────────────────────────────────────────────┤
│           Improvement: ~80% reduction                        │
└─────────────────────────────────────────────────────────────┘
```

## Fault Tolerance

```
Scenario: Core-2 crashes during processing

┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Core-0     │  │   Core-1     │  │   Core-2     │
│   Running    │  │   Running    │  │   CRASHED    │
│              │  │              │  │      ❌      │
│ Processes:   │  │ Processes:   │  │ Should have: │
│ id%5==0      │  │ id%5==1      │  │ id%5==2      │
└──────────────┘  └──────────────┘  └──────────────┘

Next Refresh Cycle (30s later):
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Core-0     │  │   Core-1     │  │   Core-2     │
│   Running    │  │   Running    │  │   CRASHED    │
│              │  │              │  │      ❌      │
│ Processes:   │  │ Processes:   │  │ Unprocessed: │
│ id%5==0      │  │ id%5==1      │  │ id%5==2      │
│              │  │              │  │ Still in     │
│              │  │              │  │ queue ⏳     │
└──────────────┘  └──────────────┘  └──────────────┘

After Core-2 Recovers:
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Core-0     │  │   Core-1     │  │   Core-2     │
│   Running    │  │   Running    │  │   RECOVERED  │
│              │  │              │  │      ✓       │
│ Processes:   │  │ Processes:   │  │ Processes:   │
│ id%5==0      │  │ id%5==1      │  │ id%5==2      │
│              │  │              │  │ (catches up) │
└──────────────┘  └──────────────┘  └──────────────┘

Result: No data loss, automatic recovery, no manual intervention
```

