package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dombox/uuidv7"
)

// TestBasicUsageExample tests the basic usage example from README
func TestBasicUsageExample(t *testing.T) {
	// Generate a single UUID
	id, err := uuidv7.New()
	if err != nil {
		t.Fatal(err)
	}

	if id.String() == "" {
		t.Error("UUID string should not be empty")
	}
	if id.Timestamp().IsZero() {
		t.Error("UUID timestamp should not be zero")
	}
	if id.Version() != 7 {
		t.Errorf("Expected version 7, got %d", id.Version())
	}

	// Must variant (panics on error)
	id2 := uuidv7.MustNew()
	if id2.String() == "" {
		t.Error("MustNew UUID string should not be empty")
	}

	t.Logf("Generated UUID: %s", id.String())
	t.Logf("Timestamp: %s", id.Timestamp())
	t.Logf("Version: %d", id.Version())
	t.Logf("Another UUID: %s", id2)
}

// TestJSONExample tests the JSON marshaling example from README
func TestJSONExample(t *testing.T) {
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
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("JSON data should not be empty")
	}

	// Unmarshal from JSON
	var newUser User
	if err := json.Unmarshal(data, &newUser); err != nil {
		t.Fatal(err)
	}

	if newUser.ID != user.ID {
		t.Errorf("UUIDs don't match: %s != %s", newUser.ID, user.ID)
	}
	if newUser.Name != user.Name {
		t.Errorf("Names don't match: %s != %s", newUser.Name, user.Name)
	}

	t.Logf("JSON: %s", string(data))
	t.Logf("Parsed: %+v", newUser)
}

// TestBatchGenerationExample tests the batch generation example from README
func TestBatchGenerationExample(t *testing.T) {
	// Generate 1000 UUIDs efficiently
	uuids, err := uuidv7.NewBatch(1000)
	if err != nil {
		t.Fatal(err)
	}

	if len(uuids) != 1000 {
		t.Errorf("Expected 1000 UUIDs, got %d", len(uuids))
	}

	// Verify they're properly ordered
	if err := uuidv7.ValidateOrdering(uuids); err != nil {
		t.Fatal("Ordering violation:", err)
	}

	t.Logf("Generated %d UUIDs", len(uuids))
	t.Logf("First: %s", uuids[0])
	t.Logf("Last: %s", uuids[len(uuids)-1])
	t.Log("✅ All UUIDs are properly time-ordered")
}

// TestCustomGeneratorExample tests the custom generator configuration example from README
func TestCustomGeneratorExample(t *testing.T) {
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
		t.Fatal(err)
	}

	if id.Version() != 7 {
		t.Errorf("Expected version 7, got %d", id.Version())
	}
	if id.Variant() != 2 {
		t.Errorf("Expected variant 2, got %d", id.Variant())
	}
	if id.Timestamp().IsZero() {
		t.Error("Timestamp should not be zero")
	}

	t.Logf("Custom UUID: %s", id)
	t.Logf("Version: %d", id.Version())
	t.Logf("Variant: %d", id.Variant())
	t.Logf("Timestamp: %s", id.Timestamp())
}

// TestParsingExample tests the parsing and validation example from README
func TestParsingExample(t *testing.T) {
	// First generate a valid UUIDv7 to parse
	originalID := uuidv7.MustNew()
	uuidStr := originalID.String()

	// Parse from string (accepts both upper and lowercase)
	id, err := uuidv7.Parse(uuidStr)
	if err != nil {
		t.Fatal("Parse error:", err)
	}

	if !id.IsValid() {
		t.Error("Parsed UUID should be valid")
	}
	if id.Version() != 7 {
		t.Errorf("Expected version 7, got %d", id.Version())
	}
	if id.Variant() != 2 {
		t.Errorf("Expected variant 2, got %d", id.Variant())
	}
	if id.Timestamp().IsZero() {
		t.Error("Timestamp should not be zero")
	}

	// Compare UUIDs
	id2 := uuidv7.MustNew()
	comparison := id.Compare(id2)
	if comparison == 0 && id != id2 {
		t.Error("Different UUIDs should not have comparison result 0")
	}

	t.Logf("Parsed UUID: %s", id)
	t.Logf("Is valid UUIDv7: %t", id.IsValid())
	t.Logf("Version: %d", id.Version())
	t.Logf("Variant: %d", id.Variant())
	t.Logf("Timestamp: %s", id.Timestamp())
	t.Logf("Comparison result: %d", comparison)
}

// TestRFC9562ExampleA6 tests the RFC 9562 A.6 example from README
func TestRFC9562ExampleA6(t *testing.T) {
	gen := uuidv7.NewGenerator()

	// Build the exact RFC 9562 A.6 example UUID
	uuid := gen.BuildUUID(
		uint64(0x017F22E279B0),     // unix_ts_ms: Tuesday, February 22, 2022 2:22:22.00 PM GMT-05:00
		uint16(0xCC3),              // rand_a
		uint64(0x18C4DC0C0C07398F), // rand_b (0b01, 0x8C4DC0C0C07398F)
	)

	expectedUUID := "017f22e2-79b0-7cc3-98c4-dc0c0c07398f"
	if uuid.String() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, uuid.String())
	}

	if uuid.Version() != 7 {
		t.Errorf("Expected version 7, got %d", uuid.Version())
	}
	if uuid.Variant() != 2 {
		t.Errorf("Expected variant 2, got %d", uuid.Variant())
	}

	// Check timestamp (should be 2022-02-22T19:22:22Z)
	expectedTime := time.Date(2022, 2, 22, 19, 22, 22, 0, time.UTC)
	if !uuid.Timestamp().Equal(expectedTime) {
		t.Errorf("Expected timestamp %s, got %s", expectedTime, uuid.Timestamp())
	}

	t.Logf("UUID String: %s", uuid.String())
	t.Logf("Version: %d", uuid.Version())
	t.Logf("Variant: %d", uuid.Variant())
	t.Logf("Timestamp: %s", uuid.Timestamp())
	t.Log("✅ Perfect RFC 9562 A.6 match!")
}

// TestMonotonicMethodsExamples tests the monotonic method examples from README
func TestMonotonicMethodsExamples(t *testing.T) {
	// Test Monotonic Random (Default)
	config1 := uuidv7.DefaultConfig()
	config1.MonotonicMethod = uuidv7.MonotonicRandom
	generator1 := uuidv7.NewGeneratorWithConfig(config1)

	id1, err := generator1.New()
	if err != nil {
		t.Fatal(err)
	}
	if id1.Version() != 7 {
		t.Errorf("MonotonicRandom: Expected version 7, got %d", id1.Version())
	}

	// Test Monotonic Counter
	config2 := uuidv7.DefaultConfig()
	config2.MonotonicMethod = uuidv7.MonotonicCounter
	config2.CounterRotationInterval = 10000
	generator2 := uuidv7.NewGeneratorWithConfig(config2)

	id2, err := generator2.New()
	if err != nil {
		t.Fatal(err)
	}
	if id2.Version() != 7 {
		t.Errorf("MonotonicCounter: Expected version 7, got %d", id2.Version())
	}

	// Test Clock Sequence
	config3 := uuidv7.DefaultConfig()
	config3.MonotonicMethod = uuidv7.MonotonicClockSequence
	config3.ClockSequenceBits = 12
	generator3 := uuidv7.NewGeneratorWithConfig(config3)

	id3, err := generator3.New()
	if err != nil {
		t.Fatal(err)
	}
	if id3.Version() != 7 {
		t.Errorf("MonotonicClockSequence: Expected version 7, got %d", id3.Version())
	}

	t.Logf("MonotonicRandom UUID: %s", id1)
	t.Logf("MonotonicCounter UUID: %s", id2)
	t.Logf("MonotonicClockSequence UUID: %s", id3)
}

// TestContextExamples tests context-based generation examples
func TestContextExamples(t *testing.T) {
	ctx := context.Background()

	// Test single UUID with context
	id, err := uuidv7.NewWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if id.Version() != 7 {
		t.Errorf("Expected version 7, got %d", id.Version())
	}

	// Test batch generation with context
	uuids, err := uuidv7.NewBatchWithContext(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(uuids) != 100 {
		t.Errorf("Expected 100 UUIDs, got %d", len(uuids))
	}

	// Test with cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	_, err = uuidv7.NewWithContext(cancelledCtx)
	if err == nil {
		t.Error("Expected error with cancelled context")
	}

	t.Logf("Context UUID: %s", id)
	t.Logf("Batch count: %d", len(uuids))
}

// TestDatabaseIntegrationExample tests the simulated database integration
func TestDatabaseIntegrationExample(t *testing.T) {
	type Product struct {
		ID        uuidv7.UUID
		Name      string
		CreatedAt time.Time
	}

	// Create products with small delays to ensure different timestamps
	products := make([]Product, 5)
	for i := range products {
		products[i] = Product{
			ID:        uuidv7.MustNew(),
			Name:      "Product " + string(rune('A'+i)),
			CreatedAt: time.Now(),
		}
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// Extract UUIDs to verify ordering
	var uuids []uuidv7.UUID
	for _, product := range products {
		uuids = append(uuids, product.ID)
	}

	// Verify they're properly ordered
	if err := uuidv7.ValidateOrdering(uuids); err != nil {
		t.Fatal("Products are not properly ordered:", err)
	}

	t.Log("✅ Products are naturally ordered by creation time!")
	for i, product := range products {
		t.Logf("Product %d: ID=%s, Name=%s", i+1, product.ID, product.Name)
	}
}

// TestMainFunction tests that the main function runs without errors
func TestMainFunction(t *testing.T) {
	// This test ensures that main() can be called without panicking
	// We can't easily capture stdout in a unit test, but we can verify it doesn't crash
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("main() panicked: %v", r)
		}
	}()

	// Don't actually call main() as it would produce a lot of output
	// Instead, test each example function individually
	basicUsageExample()
	jsonExample()
	batchGenerationExample()
	customGeneratorExample()
	parsingExample()
	rfc9562ExampleA6()
	monotonicMethodsExamples()
	contextExamples()
	databaseIntegrationExample()

	t.Log("✅ All example functions executed successfully")
}
