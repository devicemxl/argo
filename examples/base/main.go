package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/devicemxl/argo"
)

func demonstrateStringFunctions() {
	fmt.Println("=== String Functions ===")

	result := argo.WithArena(func(arena *argo.Arena) map[string]string {
		results := make(map[string]string)

		// Test basic arena functionality
		results["basic"] = arena.GoString(arena.CString("Hello, Arena!"))

		// Test specific string functions
		results["process"] = ProcessStringWithArena(arena, "Go! ", 3)
		results["concat"] = ConcatStringsWithArena(arena, "Hello", " World")
		results["upper"] = ToUppercaseWithArena(arena, "lowercase text")
		results["lower"] = ToLowercaseWithArena(arena, "UPPERCASE TEXT")
		results["cipher"] = CaesarCipherWithArena(arena, "Hello", 3)

		return results
	})

	for name, value := range result {
		fmt.Printf("  %s: %s\n", name, value)
	}
}

func demonstrateArrayFunctions() {
	fmt.Println("\n=== Array Functions ===")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Test array processing
		numbers := []int32{5, 2, 8, 1, 9, 3}
		fmt.Printf("  Original: %v\n", numbers)

		// Create handle for in-place operations
		handle := argo.NewHandle(arena, numbers)

		// Test aggregation functions
		sum := SumArrayWithArena(arena, handle.Get())
		max := MaxArrayWithArena(arena, handle.Get())
		min := MinArrayWithArena(arena, handle.Get())
		fmt.Printf("  Sum: %d, Max: %d, Min: %d\n", sum, max, min)

		// Test in-place modification
		ProcessArrayInPlace(arena, handle.Get(), 2)
		fmt.Printf("  After *2: %v\n", handle.Get())

		// Test sorting
		SortArrayInPlace(arena, handle.Get())
		fmt.Printf("  Sorted: %v\n", handle.Get())

		// Test data generation
		sequence := GenerateSequenceWithArena(arena, 10, 5)
		fmt.Printf("  Generated sequence: %v\n", sequence)

		randomArr := GenerateRandomArrayWithArena(arena, 5, 1, 100)
		fmt.Printf("  Random array: %v\n", randomArr)

		return nil
	})
}

func demonstrateMathFunctions() {
	fmt.Println("\n=== Math Functions ===")

	fmt.Printf("  FastPow(2, 10): %.0f\n", FastPow(2, 10))
	fmt.Printf("  SqrtNewton(16): %.2f\n", SqrtNewton(16))
	fmt.Printf("  Factorial(5): %d\n", Factorial(5))
}

func demonstrateCryptoFunctions() {
	fmt.Println("\n=== Crypto Functions ===")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		original := "Hello World"
		shift := 5

		encrypted := CaesarCipherWithArena(arena, original, shift)
		decrypted := CaesarDecipherWithArena(arena, encrypted, shift)

		fmt.Printf("  Original: %s\n", original)
		fmt.Printf("  Encrypted: %s\n", encrypted)
		fmt.Printf("  Decrypted: %s\n", decrypted)

		return nil
	})
}

func demonstrateValidationFunctions() {
	fmt.Println("\n=== Validation Functions ===")

	emails := []string{"test@example.com", "invalid-email", "user@domain.org"}
	urls := []string{"https://example.com", "http://test.org", "not-a-url"}

	fmt.Println("  Email validation:")
	for _, email := range emails {
		valid := IsValidEmail(email)
		fmt.Printf("    %s: %t\n", email, valid)
	}

	fmt.Println("  URL validation:")
	for _, url := range urls {
		valid := IsValidURL(url)
		fmt.Printf("    %s: %t\n", url, valid)
	}
}

func demonstrateConversionFunctions() {
	fmt.Println("\n=== Conversion Functions ===")

	// Test hex conversion
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	hex := BytesToHex(data)
	fmt.Printf("  Bytes to hex: %v -> %s\n", data, hex)

	converted := HexToBytes(hex)
	fmt.Printf("  Hex to bytes: %s -> %v\n", hex, converted)

	// Test hash
	text := "Hello, World!"
	hash := SimpleHash(text)
	fmt.Printf("  Hash of '%s': %d\n", text, hash)

	// Test number validation
	numbers := []string{"123", "456abc", "-789", "not-a-number"}
	fmt.Println("  Number validation:")
	for _, num := range numbers {
		if IsValidNumber(num) {
			if val, ok := SafeStrToInt(num); ok {
				fmt.Printf("    '%s': valid -> %d\n", num, val)
			}
		} else {
			fmt.Printf("    '%s': invalid\n", num)
		}
	}
}

func demonstrateBitFunctions() {
	fmt.Println("\n=== Bit Functions ===")

	numbers := []uint32{7, 15, 255, 1023}
	for _, num := range numbers {
		bits := CountBits(num)
		reversed := ReverseBits(num)
		fmt.Printf("  %d: bits=%d, reversed=%d\n", num, bits, reversed)
	}
}

func demonstrateUtilityFunctions() {
	fmt.Println("\n=== Utility Functions ===")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Generate large data
		data := GenerateDataWithArena(arena, 100)
		fmt.Printf("  Generated data (first 20 chars): %.20s...\n", data)

		return nil
	})
}

// Función de benchmark mejorada
func benchmarkComparison() {
	fmt.Println("\n=== Performance Comparison ===")

	iterations := 50000

	// Test con Go nativo
	start := time.Now()
	for i := 0; i < iterations; i++ {
		text := fmt.Sprintf("test string %d", i)
		_ = len(text) * 2 // Simulación de procesamiento
		if i%5000 == 0 {
			runtime.GC()
		}
	}
	goTime := time.Since(start)

	// Test con Arena + C
	start = time.Now()
	argo.WithArena(func(arena *argo.Arena) interface{} {
		for i := 0; i < iterations; i++ {
			text := fmt.Sprintf("test string %d", i)
			_ = ProcessStringWithArena(arena, text, 2)
			if i%5000 == 0 {
				runtime.GC() // No afecta la arena
			}
		}
		return nil
	})
	arenaTime := time.Since(start)

	fmt.Printf("  Go native: %v\n", goTime)
	fmt.Printf("  Arena + C: %v\n", arenaTime)
	fmt.Printf("  Speedup: %.2fx\n", float64(goTime)/float64(arenaTime))
}

func main() {
	fmt.Println("🚀 Argo Library - Complete Example Demonstration")
	fmt.Println("================================================")

	demonstrateStringFunctions()
	demonstrateArrayFunctions()
	demonstrateMathFunctions()
	demonstrateCryptoFunctions()
	demonstrateValidationFunctions()
	demonstrateConversionFunctions()
	demonstrateBitFunctions()
	demonstrateUtilityFunctions()
	benchmarkComparison()

	fmt.Println("\n✅ All demonstrations completed successfully!")
}
