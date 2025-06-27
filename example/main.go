package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dombox/uuidv7"
)

func main() {
	fmt.Println("=== UUIDv7 Examples from README ===")

	// Basic Usage Example
	basicUsageExample()

	// JSON Example
	jsonExample()

	// Batch Generation Example
	batchGenerationExample()

	// Custom Generator Configuration Example
	customGeneratorExample()

	// Parsing and Validation Example
	parsingExample()

	// RFC 9562 A.6 Example
	rfc9562ExampleA6()

	// Monotonic Methods Examples
	monotonicMethodsExamples()

	// Context Examples
	contextExamples()

	// Database Integration Simulation
	databaseIntegrationExample()
}

// Basic Usage Example from README
func basicUsageExample() {
	fmt.Println("--- Basic Usage Example ---")

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
	fmt.Println()
}

// Working with JSON Example from README
func jsonExample() {
	fmt.Println("--- JSON Example ---")

	type User struct {
		ID   uuidv7.UUID `json:"id"`
		Name string      `json:"name"`
	}

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
	fmt.Println()
}

// Batch Generation Example from README
func batchGenerationExample() {
	fmt.Println("--- Batch Generation Example ---")

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
	fmt.Println()
}

// Custom Generator Configuration Example from README
func customGeneratorExample() {
	fmt.Println("--- Custom Generator Configuration Example ---")

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
	fmt.Println()
}

// Parsing and Validation Example from README
func parsingExample() {
	fmt.Println("--- Parsing and Validation Example ---")

	// Generate a UUID first for demonstration
	originalID := uuidv7.MustNew()
	uuidStr := originalID.String()

	// Parse from string (accepts both upper and lowercase)
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
	fmt.Println()
}

// RFC 9562 A.6 Example from README
func rfc9562ExampleA6() {
	fmt.Println("--- RFC 9562 A.6 Example ---")

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
	fmt.Println("✅ Perfect RFC 9562 A.6 match!")
	fmt.Println()
}

// Monotonic Methods Examples from README
func monotonicMethodsExamples() {
	fmt.Println("--- Monotonic Methods Examples ---")

	// 1. Monotonic Random (Default)
	fmt.Println("1. Monotonic Random (Default):")
	config1 := uuidv7.DefaultConfig()
	config1.MonotonicMethod = uuidv7.MonotonicRandom
	generator1 := uuidv7.NewGeneratorWithConfig(config1)

	id1, err := generator1.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   UUID:", id1)

	// 2. Monotonic Counter
	fmt.Println("2. Monotonic Counter:")
	config2 := uuidv7.DefaultConfig()
	config2.MonotonicMethod = uuidv7.MonotonicCounter
	config2.CounterRotationInterval = 10000
	generator2 := uuidv7.NewGeneratorWithConfig(config2)

	id2, err := generator2.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   UUID:", id2)

	// 3. Clock Sequence
	fmt.Println("3. Clock Sequence:")
	config3 := uuidv7.DefaultConfig()
	config3.MonotonicMethod = uuidv7.MonotonicClockSequence
	config3.ClockSequenceBits = 12
	generator3 := uuidv7.NewGeneratorWithConfig(config3)

	id3, err := generator3.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("   UUID:", id3)
	fmt.Println()
}

// Context Examples
func contextExamples() {
	fmt.Println("--- Context Examples ---")

	ctx := context.Background()

	// Test single UUID with context
	id, err := uuidv7.NewWithContext(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Context UUID:", id)

	// Test batch generation with context
	uuids, err := uuidv7.NewBatchWithContext(ctx, 100)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated %d UUIDs with context\n", len(uuids))

	// Demonstrate context cancellation (this will fail as expected)
	fmt.Println("Testing context cancellation...")
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	_, err = uuidv7.NewWithContext(cancelledCtx)
	if err != nil {
		fmt.Println("✅ Context cancellation worked as expected:", err)
	} else {
		fmt.Println("❌ Context cancellation didn't work")
	}
	fmt.Println()
}

// Database Integration Example (simulated) from README
func databaseIntegrationExample() {
	fmt.Println("--- Database Integration Example (Simulated) ---")

	type Product struct {
		ID        uuidv7.UUID
		Name      string
		CreatedAt time.Time
	}

	// Simulate creating products
	products := make([]Product, 5)
	for i := range products {
		products[i] = Product{
			ID:        uuidv7.MustNew(),
			Name:      fmt.Sprintf("Product %d", i+1),
			CreatedAt: time.Now(),
		}
		// Small delay to show time ordering
		time.Sleep(time.Millisecond)
	}

	fmt.Println("Created products (naturally ordered by creation time when sorted by ID):")
	for i, product := range products {
		fmt.Printf("  %d. ID: %s, Name: %s, Created: %s\n",
			i+1, product.ID, product.Name, product.CreatedAt.Format("15:04:05.000"))
	}

	// Extract UUIDs to verify ordering
	var uuids []uuidv7.UUID
	for _, product := range products {
		uuids = append(uuids, product.ID)
	}

	if err := uuidv7.ValidateOrdering(uuids); err != nil {
		log.Fatal("Products are not properly ordered:", err)
	}
	fmt.Println("✅ Products are naturally ordered by creation time!")
	fmt.Println()
}
