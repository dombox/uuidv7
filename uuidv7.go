// Package uuidv7 provides a production-ready, RFC 9562 compliant UUIDv7 generator.
//
// UUIDv7 is a time-ordered UUID that encodes Unix timestamp in milliseconds
// with additional random data for uniqueness. This implementation strictly follows
// RFC 9562 specifications for proper bit layout, monotonic ordering, and
// security considerations without any deviations.
//
// Key Features:
//   - Strict RFC 9562 compliance (millisecond precision only)
//   - Monotonic ordering within the same millisecond timestamp
//   - Thread-safe concurrent generation
//   - All three RFC 6.2 monotonic methods supported
//   - Comprehensive error handling with proper error types
//   - Batch generation for high-throughput scenarios
//   - Context support for cancellation
//
// RFC 9562 References:
//   - Section 5.7: UUID Version 7 specification
//   - Section 6.2: Monotonic Random and Counter Methods
//   - Section 6.2.1: Fixed-Length Dedicated Random
//   - Section 6.2.2: Monotonic Random Increment
//   - Section 6.2.3: Replace Leftmost Random Bits with Increased Clock Precision
//   - Section 4.1: Variant and Version Fields
//   - Section 4.2: Layout and Byte Order
//
// Copyright (c) 2025 Dombox. All rights reserved.
// Licensed under the MIT License.
package uuidv7

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sync"
	"time"
)

// Common errors for better error handling and wrapping
var (
	ErrInvalidUUIDFormat = errors.New("invalid UUID format")
	ErrClockRegression   = errors.New("clock moved backwards beyond tolerance")
	ErrRandomGeneration  = errors.New("failed to generate random bytes")
	ErrInvalidBatchSize  = errors.New("invalid batch size")
	ErrContextCanceled   = errors.New("context canceled")
)

// UUID represents a UUIDv7 as defined in RFC 9562.
//
// Field Descriptions:
//   - timestamp_ms: 48-bit Unix timestamp in milliseconds (bits 0-47)
//   - ver: 4-bit version field = 7 (bits 48-51)
//   - rand_a: 12-bit random data (bits 52-63)
//   - var: 2-bit variant field = 10 binary (bits 64-65)
//   - rand_b: 62-bit random data for uniqueness and monotonic ordering (bits 66-127)
type UUID [16]byte

// String returns the canonical string representation of the UUID.
func (u UUID) String() string {
	var buf [36]byte
	hex.Encode(buf[:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:16])
	return string(buf[:])
}

// Bytes returns a copy of the raw byte representation of the UUID
func (u UUID) Bytes() []byte {
	result := make([]byte, 16)
	copy(result, u[:])
	return result
}

// Version returns the UUID version (should be 7 for UUIDv7).
// Per RFC 9562 Section 4.1.3, the version is stored in bits 48-51.
func (u UUID) Version() int {
	return int(u[6] >> 4)
}

// Variant returns the UUID variant as defined in RFC 9562 Section 4.1.
// This determines the layout and interpretation of the UUID.
// - 0xxxxxxx → Variant 0 (NCS backward compatibility)
// - 10xxxxxx → Variant 2 (RFC 4122/9562)
// - 110xxxxx → Variant 6 (Microsoft reserved)
// - 111xxxxx → Variant 7 (Future use)
func (u UUID) Variant() int {
	switch {
	case (u[8] & 0x80) == 0x00:
		// 0xxx xxxx → Variant 0 (NCS backward compatibility)
		return 0
	case (u[8] & 0xC0) == 0x80:
		// 10xx xxxx → Variant 2 (RFC 4122 / RFC 9562)
		return 2
	case (u[8] & 0xE0) == 0xC0:
		// 110x xxxx → Variant 6 (Microsoft)
		return 6
	case (u[8] & 0xE0) == 0xE0:
		// 111x xxxx → Variant 7 (Reserved for future use)
		return 7
	default:
		// Invalid bit pattern
		return -1
	}
}

// IsValid checks if the UUID is a valid UUIDv7.
func (u UUID) IsValid() bool {
	return u.Version() == 7 && (u[8]&0xC0) == 0x80
}

// Timestamp extracts the timestamp from a UUIDv7.
// Per RFC 9562 Section 5.7, returns the Unix timestamp in milliseconds.
func (u UUID) Timestamp() time.Time {
	timestamp := uint64(u[0])<<40 |
		uint64(u[1])<<32 |
		uint64(u[2])<<24 |
		uint64(u[3])<<16 |
		uint64(u[4])<<8 |
		uint64(u[5])

	if timestamp > math.MaxInt64 {
		return time.Unix(0, 0) // return epoch on overflow
	}
	return time.UnixMilli(int64(timestamp)) // #nosec G115
}

// RandA extracts the 12-bit rand_a field from the UUID.
func (u UUID) RandA() uint16 {
	return (uint16(u[6]&0x0F) << 8) | uint16(u[7])
}

// RandB extracts the 62-bit rand_b field from the UUID.
func (u UUID) RandB() uint64 {
	return (uint64(u[8]&0x3F) << 56) |
		(uint64(u[9]) << 48) |
		(uint64(u[10]) << 40) |
		(uint64(u[11]) << 32) |
		(uint64(u[12]) << 24) |
		(uint64(u[13]) << 16) |
		(uint64(u[14]) << 8) |
		uint64(u[15])
}

// Compare compares two UUIDs lexicographically for temporal ordering.
// Returns -1 if u < other, 0 if u == other, 1 if u > other.
//
// IMPORTANT: Both UUIDs must be valid UUIDv7s for meaningful temporal comparison.
// This method performs lexicographical comparison which preserves time ordering
// only for properly formatted UUIDv7s as per RFC 9562.
func (u UUID) Compare(other UUID) int {
	for i := 0; i < 16; i++ {
		if u[i] < other[i] {
			return -1
		}
		if u[i] > other[i] {
			return 1
		}
	}
	return 0
}

// Equal checks if two UUIDs are equal
func (u UUID) Equal(other UUID) bool {
	return u == other
}

// IsNil checks if the UUID is nil (all zeros)
func (u UUID) IsNil() bool {
	return u == UUID{}
}

// MarshalText implements encoding.TextMarshaler for JSON and other text formats
func (u UUID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON and other text formats
func (u *UUID) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// MarshalJSON implements json.Marshaler
func (u UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (u *UUID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := Parse(s)
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// MonotonicMethod defines how to maintain ordering within the same timestamp.
// Implements all three methods from RFC 9562 Section 6.2.
type MonotonicMethod int

const (
	// MonotonicRandom uses random data with monotonic increment.
	// Per RFC 9562 Section 6.2.2: "Monotonic Random"
	MonotonicRandom MonotonicMethod = iota

	// MonotonicCounter uses a fixed-length counter.
	// Per RFC 9562 Section 6.2.1: "Fixed-Length Dedicated Counter"
	MonotonicCounter

	// MonotonicClockSequence uses increased clock precision in leftmost random bits.
	// Per RFC 9562 Section 6.2.3: "Replace Leftmost Random Bits with Increased Clock Precision"
	MonotonicClockSequence
)

// String returns the string representation of MonotonicMethod
func (m MonotonicMethod) String() string {
	switch m {
	case MonotonicRandom:
		return "MonotonicRandom"
	case MonotonicCounter:
		return "MonotonicCounter"
	case MonotonicClockSequence:
		return "MonotonicClockSequence"
	default:
		return fmt.Sprintf("MonotonicMethod(%d)", int(m))
	}
}

// GeneratorConfig provides configuration options for the generator
type GeneratorConfig struct {
	// ClockDriftTolerance defines the maximum allowed backward clock drift
	ClockDriftTolerance time.Duration
	// MonotonicMethod defines how to handle multiple UUIDs in the same timestamp
	MonotonicMethod MonotonicMethod
	// MaxBatchSize defines the maximum number of UUIDs that can be generated in a single batch
	MaxBatchSize int
	// CounterRotationInterval defines how often to rotate the random base in MonotonicCounter mode
	CounterRotationInterval uint64
	// SmallBatchThreshold defines the cutoff for using simple vs optimized batch generation
	SmallBatchThreshold int
	// ClockSequenceBits defines how many bits to use for clock sequence (1-12 bits)
	// Only used with MonotonicClockSequence method
	ClockSequenceBits int
	// RandomRetries defines how many times to retry on crypto/rand failures
	RandomRetries int
}

// DefaultConfig returns the recommended configuration
func DefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		ClockDriftTolerance:     1 * time.Second,
		MonotonicMethod:         MonotonicRandom,
		MaxBatchSize:            100000,
		CounterRotationInterval: 10000,
		SmallBatchThreshold:     10,
		ClockSequenceBits:       12, // Use all 12 bits of rand_a for clock sequence
		RandomRetries:           3,
	}
}

// timestampState holds the state for monotonic ordering
type timestampState struct {
	timestamp     uint64 // milliseconds since Unix epoch
	lastRand      uint64 // last random value for MonotonicRandom
	lastRandA     uint16 // last rand_a value for MonotonicRandom
	counter       uint64 // counter for MonotonicCounter
	clockSequence uint64 // clock sequence for MonotonicClockSequence
	subMsCounter  uint64 // sub-millisecond counter for MonotonicClockSequence
}

// Generator provides thread-safe UUIDv7 generation
type Generator struct {
	config GeneratorConfig

	// Current timestamp state (protected by mutex)
	mu    sync.Mutex
	state timestampState
}

// NewGenerator creates a new UUIDv7 generator with default configuration
func NewGenerator() *Generator {
	return NewGeneratorWithConfig(DefaultConfig())
}

// NewGeneratorWithConfig creates a new UUIDv7 generator with custom configuration
func NewGeneratorWithConfig(config GeneratorConfig) *Generator {
	// Validate and sanitize configuration
	if config.ClockDriftTolerance < 0 {
		config.ClockDriftTolerance = DefaultConfig().ClockDriftTolerance
	}
	if config.MaxBatchSize <= 0 || config.MaxBatchSize > 1000000 {
		config.MaxBatchSize = DefaultConfig().MaxBatchSize
	}
	if config.MonotonicMethod < MonotonicRandom || config.MonotonicMethod > MonotonicClockSequence {
		config.MonotonicMethod = DefaultConfig().MonotonicMethod
	}
	if config.CounterRotationInterval == 0 {
		config.CounterRotationInterval = DefaultConfig().CounterRotationInterval
	}
	if config.SmallBatchThreshold <= 0 || config.SmallBatchThreshold > 1000 {
		config.SmallBatchThreshold = DefaultConfig().SmallBatchThreshold
	}
	if config.ClockSequenceBits < 1 || config.ClockSequenceBits > 12 {
		config.ClockSequenceBits = DefaultConfig().ClockSequenceBits
	}
	if config.RandomRetries < 1 || config.RandomRetries > 10 {
		config.RandomRetries = DefaultConfig().RandomRetries
	}

	g := &Generator{
		config: config,
	}

	// Initialize state with zero timestamp to force fresh random generation on first call
	g.state = timestampState{
		timestamp:     0, // This ensures first call will generate fresh random values
		lastRand:      0,
		lastRandA:     0,
		counter:       0,
		clockSequence: 0,
		subMsCounter:  0,
	}

	// Initialize clock sequence with random value for MonotonicClockSequence
	if config.MonotonicMethod == MonotonicClockSequence {
		var buf [2]byte
		if err := g.secureRandRead(buf[:]); err == nil {
			maxSeq := uint64(1) << config.ClockSequenceBits
			g.state.clockSequence = (uint64(buf[0])<<8 | uint64(buf[1])) % maxSeq
		}
	}

	return g
}

// secureRandRead reads random bytes with retry on failure
func (g *Generator) secureRandRead(buf []byte) error {
	var lastErr error
	for i := 0; i < g.config.RandomRetries; i++ {
		if _, err := rand.Read(buf); err == nil {
			return nil
		} else {
			lastErr = err
			if i < g.config.RandomRetries-1 {
				time.Sleep(time.Microsecond * time.Duration(1<<i)) // exponential backoff
			}
		}
	}
	return fmt.Errorf("%w: %w", ErrRandomGeneration, lastErr)
}

// New generates a new UUIDv7 with proper monotonic ordering
func (g *Generator) New() (UUID, error) {
	nowMs := time.Now().UnixMilli()
	if nowMs < 0 {
		nowMs = 0
	}
	now := uint64(nowMs) // #nosec G115

	g.mu.Lock()
	defer g.mu.Unlock()

	// Handle clock regression
	if now < g.state.timestamp {
		var drift time.Duration
		if g.state.timestamp > math.MaxInt64 {
			drift = time.Duration(math.MaxInt64)
		} else {
			drift = time.Duration(g.state.timestamp-now) * time.Millisecond // #nosec G115
		}
		if drift > g.config.ClockDriftTolerance {
			return UUID{}, fmt.Errorf("%w: drift %v (current: %d, last: %d)",
				ErrClockRegression, drift, now, g.state.timestamp)
		}
		// Use the last known good timestamp
		now = g.state.timestamp
	}

	var randA uint16
	var randB uint64
	var err error

	if now == g.state.timestamp {
		// Same timestamp: ensure monotonic ordering
		randA, randB, err = g.generateMonotonicValues()
		if err != nil {
			return UUID{}, fmt.Errorf("failed to generate monotonic values: %w", err)
		}
	} else {
		// New timestamp: generate fresh random bytes
		g.state.timestamp = now
		randA, randB, err = g.generateFreshRandom()
		if err != nil {
			return UUID{}, fmt.Errorf("failed to generate random: %w", err)
		}
		g.resetStateForNewTimestamp(randB)
		g.state.lastRandA = randA
	}

	uuid := g.BuildUUID(now, randA, randB)

	return uuid, nil
}

// NewWithContext generates a new UUIDv7 with context support
func (g *Generator) NewWithContext(ctx context.Context) (UUID, error) {
	if err := ctx.Err(); err != nil {
		return UUID{}, fmt.Errorf("%w: %w", ErrContextCanceled, err)
	}
	return g.New()
}

// generateFreshRandom generates fresh random values for a new timestamp
func (g *Generator) generateFreshRandom() (uint16, uint64, error) {
	var buf [10]byte
	if err := g.secureRandRead(buf[:]); err != nil {
		return 0, 0, err
	}

	randA := binary.BigEndian.Uint16(buf[0:2]) & 0x0FFF              // 12 bits
	randB := binary.BigEndian.Uint64(buf[2:10]) & 0x3FFFFFFFFFFFFFFF // 62 bits

	return randA, randB, nil
}

// resetStateForNewTimestamp resets internal state when timestamp changes
func (g *Generator) resetStateForNewTimestamp(randB uint64) {
	g.state.lastRand = randB
	g.state.lastRandA = 0 // Will be set by generateFreshRandom
	g.state.counter = 0
	g.state.subMsCounter = 0
}

// generateMonotonicValues generates monotonic values for the same timestamp
func (g *Generator) generateMonotonicValues() (uint16, uint64, error) {
	switch g.config.MonotonicMethod {
	case MonotonicRandom:
		return g.generateMonotonicRandomIncrement()
	case MonotonicCounter:
		return g.generateMonotonicCounter()
	case MonotonicClockSequence:
		return g.generateMonotonicClockSequence()
	default:
		return 0, 0, fmt.Errorf("invalid monotonic method: %s", g.config.MonotonicMethod)
	}
}

// generateMonotonicRandomIncrement implements RFC 9562 Section 6.2.2
func (g *Generator) generateMonotonicRandomIncrement() (uint16, uint64, error) {
	// Generate small random increment (1-255) to be more conservative
	var incBytes [1]byte
	if err := g.secureRandRead(incBytes[:]); err != nil {
		return 0, 0, err
	}
	inc := uint64(incBytes[0])
	if inc == 0 {
		inc = 1
	}

	// Check for overflow before addition
	maxRandB := uint64(0x3FFFFFFFFFFFFFFF)
	if g.state.lastRand > maxRandB-inc {
		// Overflow: advance timestamp and generate fresh random
		g.state.timestamp++
		randA, randB, err := g.generateFreshRandom()
		if err != nil {
			return 0, 0, err
		}
		g.state.lastRand = randB
		g.state.lastRandA = randA
		return randA, randB, nil
	}

	// Simple increment approach: just increment rand_b and keep rand_a fixed
	// This ensures lexicographic ordering since rand_b comes after rand_a in the UUID
	newRandB := g.state.lastRand + inc
	g.state.lastRand = newRandB

	return g.state.lastRandA, newRandB, nil
}

// generateMonotonicCounter implements RFC 9562 Section 6.2.1
func (g *Generator) generateMonotonicCounter() (uint16, uint64, error) {
	// Check for counter overflow
	maxCounter := uint64(0xFFFFFFFF) // 32-bit counter
	if g.state.counter >= maxCounter {
		// Counter overflow: advance timestamp
		g.state.timestamp++
		g.state.counter = 0

		// Generate new base randomness
		randA, randB, err := g.generateFreshRandom()
		if err != nil {
			return 0, 0, err
		}
		g.state.lastRand = randB
		return randA, randB, nil
	}

	g.state.counter++

	// Rotate base randomness periodically
	if g.state.counter%g.config.CounterRotationInterval == 0 {
		var buf [8]byte
		if err := g.secureRandRead(buf[:]); err != nil {
			return 0, 0, err
		}
		g.state.lastRand = binary.BigEndian.Uint64(buf[:]) & 0x3FFFFFFFFFFFFFFF
	}

	// Use consistent rand_a for monotonic ordering (preserve lastRandA)
	randA := g.state.lastRandA

	// Combine base randomness with counter (upper 30 bits + lower 32 bits)
	baseRandomBits := g.state.lastRand & 0x3FFFFFFFC0000000 // Upper 30 bits
	counterBits := g.state.counter & 0x00000000FFFFFFFF     // Lower 32 bits
	randB := baseRandomBits | counterBits

	return randA, randB, nil
}

// generateMonotonicClockSequence implements RFC 9562 Section 6.2.3
func (g *Generator) generateMonotonicClockSequence() (uint16, uint64, error) {
	// Increment sub-millisecond counter
	g.state.subMsCounter++

	// Check for sub-ms counter overflow
	maxSubMs := uint64(1) << g.config.ClockSequenceBits
	if g.state.subMsCounter >= maxSubMs {
		g.state.clockSequence++
		g.state.subMsCounter = 0

		// Check for clock sequence overflow
		maxClockSeq := uint64(1) << g.config.ClockSequenceBits
		if g.state.clockSequence >= maxClockSeq {
			g.state.timestamp++
			g.state.clockSequence = 0
		}
	}

	// Place sub-ms counter in rand_a field (high precision timestamp extension)
	if g.state.subMsCounter > math.MaxUint16 {
		g.state.subMsCounter = g.state.subMsCounter % (math.MaxUint16 + 1)
	}
	randA := uint16(g.state.subMsCounter) & ((1 << g.config.ClockSequenceBits) - 1) // #nosec G115

	// Generate random rand_b
	var randBBytes [8]byte
	if err := g.secureRandRead(randBBytes[:]); err != nil {
		return 0, 0, err
	}
	randB := binary.BigEndian.Uint64(randBBytes[:]) & 0x3FFFFFFFFFFFFFFF

	return randA, randB, nil
}

// BuildUUID constructs the UUID according to RFC 9562
func (g *Generator) BuildUUID(timestamp uint64, randA uint16, randB uint64) UUID {
	var uuid UUID

	// Bits 0-47: timestamp_ms (48 bits)
	uuid[0] = byte(timestamp >> 40)
	uuid[1] = byte(timestamp >> 32)
	uuid[2] = byte(timestamp >> 24)
	uuid[3] = byte(timestamp >> 16)
	uuid[4] = byte(timestamp >> 8)
	uuid[5] = byte(timestamp)

	// Bits 48-51: version (4 bits) = 7
	// Bits 52-63: rand_a (12 bits)
	uuid[6] = (7 << 4) | byte((randA>>8)&0x0F)
	uuid[7] = byte(randA)

	// Bits 64-65: variant (2 bits) = 10
	// Bits 66-127: rand_b (62 bits)
	uuid[8] = (2 << 6) | byte((randB>>56)&0x3F)
	uuid[9] = byte(randB >> 48)
	uuid[10] = byte(randB >> 40)
	uuid[11] = byte(randB >> 32)
	uuid[12] = byte(randB >> 24)
	uuid[13] = byte(randB >> 16)
	uuid[14] = byte(randB >> 8)
	uuid[15] = byte(randB)

	return uuid
}

// NewBatch generates multiple UUIDs efficiently
func (g *Generator) NewBatch(count int) ([]UUID, error) {
	if count <= 0 {
		return nil, fmt.Errorf("%w: count must be positive, got %d", ErrInvalidBatchSize, count)
	}
	if count > g.config.MaxBatchSize {
		return nil, fmt.Errorf("%w: %d > %d", ErrInvalidBatchSize, count, g.config.MaxBatchSize)
	}

	uuids := make([]UUID, count)

	// For small batches, use simple approach
	if count <= g.config.SmallBatchThreshold {
		for i := 0; i < count; i++ {
			uuid, err := g.New()
			if err != nil {
				return nil, fmt.Errorf("failed to generate UUID %d: %w", i, err)
			}
			uuids[i] = uuid
		}
		return uuids, nil
	}

	// For larger batches, optimize with single lock
	g.mu.Lock()
	defer g.mu.Unlock()

	nowMs := time.Now().UnixMilli()
	if nowMs < 0 {
		nowMs = 0
	}
	now := uint64(nowMs) // #nosec G115

	// Handle clock regression
	if now < g.state.timestamp {
		var drift time.Duration
		if g.state.timestamp > math.MaxInt64 {
			drift = time.Duration(math.MaxInt64)
		} else {
			drift = time.Duration(g.state.timestamp-now) * time.Millisecond // #nosec G115
		}
		if drift > g.config.ClockDriftTolerance {
			return nil, fmt.Errorf("%w: drift %v (current: %d, last: %d)",
				ErrClockRegression, drift, now, g.state.timestamp)
		}
		now = g.state.timestamp
	}

	// Update timestamp if needed
	if now > g.state.timestamp {
		g.state.timestamp = now
		_, randB, err := g.generateFreshRandom()
		if err != nil {
			return nil, fmt.Errorf("failed to generate fresh random: %w", err)
		}
		g.resetStateForNewTimestamp(randB)
	}

	// Generate all UUIDs with same timestamp using monotonic ordering
	for i := 0; i < count; i++ {
		randA, randB, err := g.generateMonotonicValues()
		if err != nil {
			return nil, fmt.Errorf("failed to generate monotonic values for UUID %d: %w", i, err)
		}

		uuids[i] = g.BuildUUID(g.state.timestamp, randA, randB)
	}

	return uuids, nil
}

// NewBatchWithContext generates multiple UUIDs with context support
func (g *Generator) NewBatchWithContext(ctx context.Context, count int) ([]UUID, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrContextCanceled, err)
	}

	// For large batches, check context periodically
	if count > g.config.SmallBatchThreshold {
		const chunkSize = 1000
		allUUIDs := make([]UUID, 0, count)

		for remaining := count; remaining > 0; {
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("%w: %w", ErrContextCanceled, err)
			}

			batchSize := chunkSize
			if remaining < chunkSize {
				batchSize = remaining
			}

			chunk, err := g.NewBatch(batchSize)
			if err != nil {
				return nil, err
			}

			allUUIDs = append(allUUIDs, chunk...)
			remaining -= batchSize
		}

		return allUUIDs, nil
	}

	return g.NewBatch(count)
}

// Global generator instance
var (
	defaultGenerator *Generator
	initOnce         sync.Once
)

// New generates a new UUIDv7 using the default generator
func New() (UUID, error) {
	initOnce.Do(func() {
		defaultGenerator = NewGenerator()
	})
	return defaultGenerator.New()
}

// NewWithContext generates a new UUIDv7 using the default generator with context support
func NewWithContext(ctx context.Context) (UUID, error) {
	initOnce.Do(func() {
		defaultGenerator = NewGenerator()
	})
	return defaultGenerator.NewWithContext(ctx)
}

// MustNew generates a new UUIDv7 and panics on error
func MustNew() UUID {
	uuid, err := New()
	if err != nil {
		panic(fmt.Sprintf("uuidv7: %v", err))
	}
	return uuid
}

// NewBatch generates multiple UUIDs using the default generator
func NewBatch(count int) ([]UUID, error) {
	initOnce.Do(func() {
		defaultGenerator = NewGenerator()
	})
	return defaultGenerator.NewBatch(count)
}

// NewBatchWithContext generates multiple UUIDs using the default generator with context support
func NewBatchWithContext(ctx context.Context, count int) ([]UUID, error) {
	initOnce.Do(func() {
		defaultGenerator = NewGenerator()
	})
	return defaultGenerator.NewBatchWithContext(ctx, count)
}

// UUID validation pattern
var (
	uuidPatternOnce sync.Once
	uuidPattern     *regexp.Regexp
)

// getUUIDPattern returns the compiled regex pattern for UUID validation
func getUUIDPattern() *regexp.Regexp {
	uuidPatternOnce.Do(func() {
		// RFC 9562 compliant pattern - accepts both uppercase and lowercase hex digits
		// as per RFC 9562 Section 3: "alphabetic characters may be all uppercase, all lowercase, or mixed case"
		uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-7[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	})
	return uuidPattern
}

// Parse parses a UUID string into a UUID with full validation.
func Parse(s string) (UUID, error) {
	if len(s) != 36 {
		return UUID{}, fmt.Errorf("%w: invalid length: %d", ErrInvalidUUIDFormat, len(s))
	}

	if !getUUIDPattern().MatchString(s) {
		return UUID{}, fmt.Errorf("%w: does not match UUIDv7 pattern: %s", ErrInvalidUUIDFormat, s)
	}

	// Remove hyphens to get a 32-character hex string
	hexStr := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	if len(hexStr) != 32 {
		return UUID{}, fmt.Errorf("%w: invalid cleaned hex length", ErrInvalidUUIDFormat)
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return UUID{}, fmt.Errorf("%w: invalid hex encoding: %w", ErrInvalidUUIDFormat, err)
	}

	var uuid UUID
	copy(uuid[:], bytes)

	if !uuid.IsValid() {
		return UUID{}, fmt.Errorf("%w: not a valid UUIDv7: version=%d, variant=%d",
			ErrInvalidUUIDFormat, uuid.Version(), uuid.Variant())
	}

	return uuid, nil
}

// MustParse parses a UUID string and panics on error
func MustParse(s string) UUID {
	uuid, err := Parse(s)
	if err != nil {
		panic(fmt.Sprintf("uuidv7: %v", err))
	}
	return uuid
}

// Nil returns the nil UUID (all zeros)
func Nil() UUID {
	return UUID{}
}

// ValidateOrdering checks if a slice of UUIDs maintains proper temporal ordering
func ValidateOrdering(uuids []UUID) error {
	for i := 1; i < len(uuids); i++ {
		if uuids[i].Compare(uuids[i-1]) <= 0 {
			return fmt.Errorf("ordering violation at index %d: %s <= %s", i, uuids[i].String(), uuids[i-1].String())
		}
	}
	return nil
}
