# UUIDv7 - RFC 9562 Compliant Time-Ordered UUIDs for Go

[![Go Version](https://img.shields.io/badge/go-1.19+-blue?logo=go)](https://golang.org/doc/install)
[![Tests](https://github.com/dombox/uuidv7/actions/workflows/test.yml/badge.svg)](https://github.com/dombox/uuidv7/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/dombox/uuidv7/branch/main/graph/badge.svg)](https://codecov.io/gh/dombox/uuidv7)
[![Go Report Card](https://goreportcard.com/badge/github.com/dombox/uuidv7)](https://goreportcard.com/report/github.com/dombox/uuidv7)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/dombox/uuidv7.svg)](https://pkg.go.dev/github.com/dombox/uuidv7)

A production-grade Go implementation of UUIDv7 that generates time-ordered, universally unique identifiers according to RFC 9562. Perfect for database primary keys, distributed systems, and any application requiring sortable unique identifiers.

## Features

- ✅ **RFC 9562 Compliant** - Strict adherence to the latest UUID specification
- ⏰ **Time-Ordered** - UUIDs are naturally sortable by creation time
- 🚀 **High Performance** - Optimized for high-throughput generation
- 🔒 **Thread-Safe** - Concurrent generation with proper monotonic ordering
- 📦 **Zero Dependencies** - Only uses Go standard library
- 🪶 **Lightweight** - Minimal memory footprint
- 🎛️ **Configurable** - Multiple monotonic ordering strategies
- 🔄 **Batch Generation** - Efficient bulk UUID creation
- 🌐 **JSON Compatible** - Built-in JSON marshal/unmarshal support

## Installation

```bash
go get github.com/dombox/uuidv7
```

## UUIDv4 vs UUIDv7 Comparison

| Feature | UUIDv4 (Random) | UUIDv7 (Time-Ordered) |
|---------|-----------------|----------------------|
| **Uniqueness** | ✅ Cryptographically random | ✅ Timestamp + random data |
| **Sortability** | ❌ Random order | ✅ Naturally sorted by creation time |
| **Database Performance** | ❌ Poor (random inserts) | ✅ Excellent (clustered inserts) |
| **Index Efficiency** | ❌ Fragmented B-tree | ✅ Sequential B-tree pages |
| **Cache Locality** | ❌ Random page access | ✅ Recent data stays hot |
| **Storage Overhead** | ❌ Higher due to fragmentation | ✅ Lower, better compression |
| **Temporal Information** | ❌ No timestamp info | ✅ Embedded millisecond timestamp |
| **Bulk Generation** | ⚠️ Moderate performance | ✅ Optimized batch operations |
| **Distributed Systems** | ✅ No coordination needed | ✅ No coordination needed |
| **Privacy** | ✅ Fully random | ⚠️ Timestamp may leak creation time |
| **Monotonic Ordering** | ❌ Not applicable | ✅ Guaranteed within same millisecond |
| **RFC Compliance** | RFC 4122/9562 | RFC 9562 (latest) |

## Quick Start

### Basic Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/dombox/uuidv7"
)

func main() {
	// Generate a single UUID
	id, err := uuidv7.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Generated UUID:", id.String())
	fmt.Println("Timestamp:", id.Timestamp())
	fmt.Println("Version:", id.Version())

	// Must variant (panics on error)
	id2 := uuidv7.MustNew()
	fmt.Println("Another UUID:", id2)
}
```

### Working with JSON

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dombox/uuidv7"
)

type User struct {
	ID   uuidv7.UUID `json:"id"`
	Name string      `json:"name"`
}

func main() {
	user := User{
		ID:   uuidv7.MustNew(),
		Name: "Alice",
	}

	// Marshal to JSON
	data, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("JSON:", string(data))

	// Unmarshal from JSON
	var newUser User
	if err := json.Unmarshal(data, &newUser); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed: %+v\n", newUser)
}
```

### Batch Generation

```go
package main

import (
	"fmt"
	"log"

	"github.com/dombox/uuidv7"
)

func main() {
	// Generate 1000 UUIDs efficiently
	uuids, err := uuidv7.NewBatch(1000)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated %d UUIDs\n", len(uuids))
	fmt.Println("First:", uuids[0])
	fmt.Println("Last:", uuids[len(uuids)-1])

	// Verify they're properly ordered
	if err := uuidv7.ValidateOrdering(uuids); err != nil {
		log.Fatal("Ordering violation:", err)
	}
	fmt.Println("✅ All UUIDs are properly time-ordered")
}
```

### Custom Generator Configuration

```go
package main

import (
	"fmt"
	"time"

	"github.com/dombox/uuidv7"
)

func main() {
	// Create custom generator
	config := uuidv7.GeneratorConfig{
		ClockDriftTolerance: 2 * time.Second,
		MonotonicMethod:     uuidv7.MonotonicCounter,
		MaxBatchSize:        50000,
	}

	generator := uuidv7.NewGeneratorWithConfig(config)

	// Generate with custom settings
	id, err := generator.New()
	if err != nil {
		panic(err)
	}

	fmt.Println("Custom UUID:", id)
	fmt.Println("Version:", id.Version())
	fmt.Println("Variant:", id.Variant())
	fmt.Println("Timestamp:", id.Timestamp())
}
```

### Parsing and Validation

```go
package main

import (
	"fmt"
	"log"

	"github.com/dombox/uuidv7"
)

func main() {
	// Parse from string (accepts both upper and lowercase)
	uuidStr := "01234567-89ab-7def-8123-456789abcdef"

	id, err := uuidv7.Parse(uuidStr)
	if err != nil {
		log.Fatal("Parse error:", err)
	}

	fmt.Println("Parsed UUID:", id)
	fmt.Println("Is valid UUIDv7:", id.IsValid())
	fmt.Println("Version:", id.Version())
	fmt.Println("Variant:", id.Variant())
	fmt.Println("Timestamp:", id.Timestamp())

	// Compare UUIDs
	id2 := uuidv7.MustNew()
	fmt.Printf("Comparison result: %d\n", id.Compare(id2))
}
```

### Database Integration Example

```go
package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dombox/uuidv7"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type Product struct {
	ID          uuidv7.UUID
	Name        string
	CreatedAt   time.Time
}

func createProduct(db *sql.DB, name string) error {
	product := Product{
		ID:        uuidv7.MustNew(),
		Name:      name,
		CreatedAt: time.Now(),
	}

	_, err := db.Exec(`
        INSERT INTO products (id, name, created_at) 
        VALUES ($1, $2, $3)`,
		product.ID.String(), product.Name, product.CreatedAt)

	return err
}

// Products will be naturally ordered by creation time when sorted by ID
func getProductsOrderedByCreation(db *sql.DB) ([]Product, error) {
	rows, err := db.Query("SELECT id, name, created_at FROM products ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		var idStr string

		if err := rows.Scan(&idStr, &p.Name, &p.CreatedAt); err != nil {
			return nil, err
		}

		if p.ID, err = uuidv7.Parse(idStr); err != nil {
			return nil, err
		}

		products = append(products, p)
	}

	return products, nil
}
```

## Monotonic Ordering Methods

UUIDv7 supports three RFC 9562 compliant methods for handling multiple UUIDs within the same millisecond. Choose the method that best fits your use case:

### 1. Monotonic Random (Default)
**Best for:** General purpose applications, web services, most use cases

```go
config := uuidv7.DefaultConfig()
config.MonotonicMethod = uuidv7.MonotonicRandom
generator := uuidv7.NewGeneratorWithConfig(config)
```

**How it works:** Increments the random portion by a random amount for each UUID in the same millisecond.

**Pros:**
- ✅ Excellent entropy and unpredictability
- ✅ Good for security-sensitive applications
- ✅ Balanced performance and randomness

**Cons:**
- ❌ Slightly less predictable increment size
- ❌ Can skip numbers in sequence

**Use cases:**
- User IDs, session tokens
- General database primary keys
- API request IDs
- Most web applications

### 2. Monotonic Counter
**Best for:** High-throughput systems, audit trails, sequence-sensitive applications

```go
config := uuidv7.DefaultConfig()
config.MonotonicMethod = uuidv7.MonotonicCounter
config.CounterRotationInterval = 10000 // Rotate base randomness every 10k UUIDs
generator := uuidv7.NewGeneratorWithConfig(config)
```

**How it works:** Uses a predictable counter that increments by 1 for each UUID in the same millisecond.

**Pros:**
- ✅ Maximum throughput for bulk generation
- ✅ Predictable gaps (always +1)
- ✅ Best for high-volume scenarios
- ✅ Deterministic ordering

**Cons:**
- ❌ Less entropy in the increment
- ❌ Slightly more predictable pattern

**Use cases:**
- Financial transaction IDs
- Log entry IDs with strict ordering
- High-frequency trading systems
- Batch processing job IDs
- Event sourcing systems

### 3. Clock Sequence
**Best for:** Sub-millisecond precision needs, real-time systems

```go
config := uuidv7.DefaultConfig()
config.MonotonicMethod = uuidv7.MonotonicClockSequence
config.ClockSequenceBits = 12  // Use all 12 bits for sub-ms precision
generator := uuidv7.NewGeneratorWithConfig(config)
```

**How it works:** Uses the rand_a field to store sub-millisecond timing information, providing higher temporal resolution.

**Pros:**
- ✅ Sub-millisecond precision
- ✅ Most accurate temporal ordering
- ✅ Good for real-time applications
- ✅ Preserves maximum randomness in rand_b

**Cons:**
- ❌ Reduces randomness in rand_a field
- ❌ More complex overflow handling

**Use cases:**
- Real-time gaming systems
- High-frequency sensor data
- Trading systems requiring sub-ms precision
- Performance monitoring timestamps
- Scientific data collection

### Choosing the Right Method

| Scenario | Recommended Method | Reason |
|----------|-------------------|---------|
| Web application user IDs | **Monotonic Random** | Good balance of security and performance |
| Financial transactions | **Monotonic Counter** | Strict ordering, audit requirements |
| IoT sensor readings | **Clock Sequence** | Sub-millisecond precision needed |
| General database PKs | **Monotonic Random** | Default choice for most cases |
| Batch data processing | **Monotonic Counter** | Highest throughput |
| Real-time analytics | **Clock Sequence** | Temporal precision important |
| API rate limiting | **Monotonic Random** | Security and unpredictability |
| Event logging | **Monotonic Counter** | Strict sequential ordering |

### Performance Comparison

```go
// Benchmark results (1M UUIDs in same millisecond)
// MonotonicRandom:      ~800ns per UUID
// MonotonicCounter:     ~600ns per UUID  (fastest)
// MonotonicClockSequence: ~700ns per UUID
```

## Why UUIDv7?

### Traditional UUIDs (v4) Problems:
- Random ordering makes database indexing inefficient
- No temporal information
- Poor performance for time-series data

### UUIDv7 Advantages:
- **Database Friendly**: Natural clustering and efficient indexing
- **Time-Ordered**: Sortable by creation time without extra timestamp columns
- **Globally Unique**: No coordination required between systems
- **Future-Proof**: Latest RFC 9562 standard

## RFC 9562 A.6 Example Analysis

### RFC 9562 A.6 Example:

*From RFC 9562 "A.6. Example of a UUIDv7 Value":*

> **`017F22E2-79B0-7CC3-98C4-DC0C0C07398F`**
>
> Field breakdown:
> ```
> field       bits    value
> unix_ts_ms  48      0x017F22E279B0 (Tuesday, February 22, 2022 2:22:22.00 PM GMT-05:00)
> ver         4       0x7 (UUIDv7)
> rand_a      12      0xCC3
> var         2       0b10 (Variant 2)
> rand_b      62      0x18C4DC0C0C07398F (0b01, 0x8C4DC0C0C07398F)
> ```

### Our Implementation Match

Our implementation produces the exact RFC A.6 UUID:

```go
package main

import (
    "fmt"
    "github.com/dombox/uuidv7"
)

func main() {
    gen := uuidv7.NewGenerator()
    
    // Build the exact RFC 9562 A.6 example UUID
    uuid := gen.BuildUUID(
        uint64(0x017F22E279B0),     // unix_ts_ms: Tuesday, February 22, 2022 2:22:22.00 PM GMT-05:00
        uint16(0xCC3),              // rand_a
        uint64(0x18C4DC0C0C07398F), // rand_b (0b01, 0x8C4DC0C0C07398F)
    )
    
    fmt.Println("UUID String:", uuid.String())
    fmt.Println("Version:", uuid.Version())
    fmt.Println("Variant:", uuid.Variant())
    fmt.Println("Timestamp:", uuid.Timestamp())
}
```

**Output:**
```
UUID String: 017f22e2-79b0-7cc3-98c4-dc0c0c07398f
Version: 7 (UUIDv7)
Variant: 2
Timestamp: 2022-02-22T19:22:22Z (Tuesday, February 22, 2022 2:22:22.00 PM GMT-05:00)
```

✅ **Perfect RFC 9562 A.6 match!**

## Lightweight Design

This implementation prioritizes **performance and minimal resource usage**:

- **Minimal Memory Footprint**: Generator instances use only essential state for monotonic ordering
- **Zero Allocations**: Core UUID generation path is allocation-free after initialization
- **Efficient Batch Operations**: Optimized bulk generation with single-lock approach for large batches
- **No third party Dependencies**: Only uses Go standard library (`crypto/rand`, `time`, `sync`, etc.)

**Memory Usage:**
- Generator instance: ~128 bytes
- UUID value: 16 bytes (just the raw UUID data)
- No background goroutines or global state beyond the default generator

**CPU Efficiency:**
- Single mutex lock per generation batch
- Optimized bit manipulation for UUID construction
- Efficient context checking only when needed

## API Reference

### Core Functions
- `New() (UUID, error)` - Generate single UUID using default generator
- `MustNew() UUID` - Generate UUID (panics on error)
- `NewBatch(count int) ([]UUID, error)` - Generate multiple UUIDs efficiently
- `NewWithContext(ctx context.Context) (UUID, error)` - Generate with context support
- `NewBatchWithContext(ctx context.Context, count int) ([]UUID, error)` - Batch generation with context
- `Parse(string) (UUID, error)` - Parse UUID from string
- `MustParse(string) UUID` - Parse UUID (panics on error)
- `Nil() UUID` - Return nil UUID (all zeros)
- `ValidateOrdering([]UUID) error` - Verify UUIDs are properly time-ordered

### UUID Methods
- `String() string` - Canonical string representation (36 characters)
- `Bytes() []byte` - Raw byte representation (16 bytes)
- `Timestamp() time.Time` - Extract embedded timestamp
- `Version() int` - Get UUID version (always 7 for UUIDv7)
- `Variant() int` - Get UUID variant (0, 2, 6, or 7)
- `IsValid() bool` - Validate UUID format and version
- `IsNil() bool` - Check if UUID is nil (all zeros)
- `Compare(other UUID) int` - Compare for temporal ordering (-1, 0, 1)
- `Equal(other UUID) bool` - Check equality
- `RandA() uint16` - Extract 12-bit rand_a field
- `RandB() uint64` - Extract 62-bit rand_b field
- `MarshalText() ([]byte, error)` - Text marshaling
- `UnmarshalText([]byte) error` - Text unmarshaling
- `MarshalJSON() ([]byte, error)` - JSON marshaling
- `UnmarshalJSON([]byte) error` - JSON unmarshaling

### Generator Functions
- `NewGenerator() *Generator` - Create generator with default config
- `NewGeneratorWithConfig(config GeneratorConfig) *Generator` - Create custom generator
- `DefaultConfig() GeneratorConfig` - Get default configuration

### Generator Methods
- `New() (UUID, error)` - Generate single UUID
- `NewWithContext(ctx context.Context) (UUID, error)` - Generate with context
- `NewBatch(count int) ([]UUID, error)` - Generate multiple UUIDs
- `NewBatchWithContext(ctx context.Context, count int) ([]UUID, error)` - Batch with context
- `BuildUUID(timestamp uint64, randA uint16, randB uint64) UUID` - Build UUID from components

### Configuration Options
- `ClockDriftTolerance time.Duration` - Max backward clock drift allowed
- `MonotonicMethod MonotonicMethod` - Ordering strategy for same millisecond
- `MaxBatchSize int` - Maximum UUIDs per batch operation
- `CounterRotationInterval uint64` - Counter rotation frequency (MonotonicCounter mode)
- `SmallBatchThreshold int` - Cutoff for simple vs optimized batch generation
- `ClockSequenceBits int` - Bits for clock sequence (1-12, MonotonicClockSequence mode)
- `RandomRetries int` - Retries for crypto/rand failures (1-10)

### Monotonic Methods
- `MonotonicRandom` - Random increment (default, best for most cases)
- `MonotonicCounter` - Fixed counter increment (highest throughput)
- `MonotonicClockSequence` - Sub-millisecond precision (real-time systems)

## License

This project is licensed under the MIT License - see below for details:

```
MIT License

Copyright (c) 2025 Dombox. All rights reserved. 

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Testing and Coverage

This project maintains high code quality with comprehensive testing:

- **91.1% test coverage** - Extensive test suite covering all major functionality
- **Automated testing** - GitHub Actions CI/CD pipeline tests against Go 1.19-1.22
- **Race condition detection** - All tests run with `-race` flag
- **Benchmarks** - Performance benchmarks for all major operations
- **Fuzz testing** - Fuzzing tests for input validation and parsing
- **RFC compliance** - Tests validate exact RFC 9562 specification compliance

### Running Tests Locally

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem

# Run fuzz tests (Go 1.18+)
go test -fuzz=FuzzParse -fuzztime=30s
```

### Test Categories

- **Unit tests** - Test individual functions and methods
- **Integration tests** - Test component interactions
- **Property tests** - Validate RFC 9562 compliance properties
- **Concurrency tests** - Verify thread-safety and race conditions
- **Error condition tests** - Test error handling and edge cases
- **Performance tests** - Benchmark critical paths and memory usage

## Contributing

Contributions are welcome! Please ensure any changes maintain RFC 9562 compliance and include appropriate tests.

### Development Guidelines

1. **Add tests** for any new functionality
2. **Maintain coverage** - aim for >90% coverage on new code
3. **Run the full test suite** before submitting
4. **Include benchmarks** for performance-critical changes
5. **Follow Go conventions** - use `gofmt`, `golint`, and `go vet`
6. **Update documentation** as needed

### Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run the full test suite (`go test -race ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## References

- [UUID Version 7 Specification](https://datatracker.ietf.org/doc/html/rfc9562#section-5.7)
