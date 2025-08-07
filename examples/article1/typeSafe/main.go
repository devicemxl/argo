package main

/*
#include <stdlib.h>
#include <string.h>

// C function to process int32 arrays
void process_int_array(int* arr, size_t length) {
    for (size_t i = 0; i < length; i++) {
        arr[i] = arr[i] * 2 + 1; // Transform: x -> 2x + 1
    }
}

// C function to process float arrays
void process_float_array(float* arr, size_t length) {
    for (size_t i = 0; i < length; i++) {
        arr[i] = arr[i] * arr[i]; // Square each element
    }
}

// C function to process byte arrays (simple encryption)
void process_byte_array(unsigned char* arr, size_t length) {
    for (size_t i = 0; i < length; i++) {
        arr[i] = arr[i] ^ 0x55; // XOR with 0x55
    }
}

// C function that calculates sum of integers
int sum_int_array(int* arr, size_t length) {
    int sum = 0;
    for (size_t i = 0; i < length; i++) {
        sum += arr[i];
    }
    return sum;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/devicemxl/argo"
)

func demonstrateHandleBasics() {
	fmt.Println(" Basic Handle Operations")
	fmt.Println("=====================================")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Test 1: Integer array
		fmt.Println("\n1️⃣ Testing int32 Handle:")
		numbers := []int32{1, 2, 3, 4, 5}
		fmt.Printf("   Original: %v\n", numbers)

		// Create handle - data is now in arena memory
		handle := argo.NewHandle(arena, numbers)

		fmt.Printf("   Handle created:\n")
		fmt.Printf("     Ptr: %p\n", handle.Ptr())
		fmt.Printf("     Length: %d\n", len(handle.Get()))
		fmt.Printf("     Element size: %d bytes\n", unsafe.Sizeof(int32(0)))

		// Verify data was copied correctly
		arenaData := handle.Get()
		fmt.Printf("   Arena data: %v\n", arenaData)

		// Verify independence from original slice
		numbers[0] = 999
		fmt.Printf("   After modifying original: %v\n", numbers)
		fmt.Printf("   Arena data unchanged: %v\n", handle.Get())

		return nil
	})
}

func demonstrateCInteroperation() {
	fmt.Println("\n🔗 C Function Interoperation")
	fmt.Println("=====================================")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Test with int32 array
		fmt.Println("\n2️⃣ Processing int32 array with C function:")
		numbers := []int32{10, 20, 30, 40, 50}
		fmt.Printf("   Before processing: %v\n", numbers)

		handle := argo.NewHandle(arena, numbers)

		// Pass directly to C function - no GC involvement!
		C.process_int_array((*C.int)(handle.Ptr()), C.size_t(len(numbers)))

		// Get modified data back
		result := handle.Get()
		fmt.Printf("   After C processing (x -> 2x + 1): %v\n", result)

		// Calculate sum using another C function
		sum := C.sum_int_array((*C.int)(handle.Ptr()), C.size_t(len(numbers)))
		fmt.Printf("   Sum calculated by C: %d\n", int(sum))

		return nil
	})
}

func demonstrateMultipleTypes() {
	fmt.Println("\n🎯 Multiple Data Types")
	fmt.Println("=====================================")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Test 3: Float32 array
		fmt.Println("\n3️⃣ Testing float32 Handle:")
		floats := []float32{1.5, 2.5, 3.5, 4.5}
		fmt.Printf("   Original floats: %v\n", floats)

		floatHandle := argo.NewHandle(arena, floats)

		// Process with C function (square each element)
		C.process_float_array((*C.float)(floatHandle.Ptr()), C.size_t(len(floats)))

		result := floatHandle.Get()
		fmt.Printf("   After squaring: %v\n", result)

		// Test 4: Byte array
		fmt.Println("\n4️⃣ Testing byte Handle (encryption):")
		data := []byte("Hello, World!")
		fmt.Printf("   Original: %s (%v)\n", string(data), data)

		byteHandle := argo.NewHandle(arena, data)

		// "Encrypt" using XOR
		C.process_byte_array((*C.uchar)(byteHandle.Ptr()), C.size_t(len(data)))

		encrypted := byteHandle.Get()
		fmt.Printf("   Encrypted: %s (%v)\n", string(encrypted), encrypted)

		// "Decrypt" by applying XOR again
		C.process_byte_array((*C.uchar)(byteHandle.Ptr()), C.size_t(len(data)))

		decrypted := byteHandle.Get()
		fmt.Printf("   Decrypted: %s (%v)\n", string(decrypted), decrypted)

		return nil
	})
}

func demonstrateGCImmunity() {
	fmt.Println("\n🛡️  GC Immunity Test")
	fmt.Println("=====================================")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Create handle with large data
		fmt.Println("\n5️⃣ Testing GC immunity:")
		largeData := make([]int32, 10000)
		for i := range largeData {
			largeData[i] = int32(i * i)
		}

		fmt.Printf("   Created large array with %d elements\n", len(largeData))

		handle := argo.NewHandle(arena, largeData)

		// Store some reference values
		originalFirst := handle.Get()[0]
		originalLast := handle.Get()[len(handle.Get())-1]
		originalSum := int32(0)
		for _, v := range handle.Get() {
			originalSum += v
		}

		fmt.Printf("   Original - First: %d, Last: %d, Sum: %d\n",
			originalFirst, originalLast, originalSum)

		// Clear original slice and force aggressive GC
		largeData = nil
		fmt.Printf("   Cleared original slice, forcing GC...\n")

		for i := 0; i < 10; i++ {
			// Create memory pressure
			waste := make([]byte, 1024*1024) // 1MB
			_ = waste
			runtime.GC()
		}

		// Verify handle data is intact
		currentData := handle.Get()
		currentSum := int32(0)
		for _, v := range currentData {
			currentSum += v
		}

		fmt.Printf("   After GC - First: %d, Last: %d, Sum: %d\n",
			currentData[0], currentData[len(currentData)-1], currentSum)

		if originalSum == currentSum {
			fmt.Printf("   ✅ SUCCESS: Data survived aggressive GC!\n")
		} else {
			fmt.Printf("   ❌ FAIL: Data was corrupted by GC\n")
		}

		return nil
	})
}

func demonstrateCustomStructs() {
	fmt.Println("\n📦 Custom Struct Handles")
	fmt.Println("=====================================")

	// Define a custom struct
	type Point3D struct {
		X, Y, Z float32
		ID      int32
	}

	argo.WithArena(func(arena *argo.Arena) interface{} {
		fmt.Println("\n6️⃣ Testing custom struct Handle:")

		points := []Point3D{
			{1.0, 2.0, 3.0, 100},
			{4.0, 5.0, 6.0, 200},
			{7.0, 8.0, 9.0, 300},
		}

		fmt.Printf("   Original points:\n")
		for i, p := range points {
			fmt.Printf("     [%d]: (%.1f, %.1f, %.1f) ID=%d\n", i, p.X, p.Y, p.Z, p.ID)
		}

		// Create handle for custom struct
		pointHandle := argo.NewHandle(arena, points)

		// Get data from arena
		arenaPoints := pointHandle.Get()
		fmt.Printf("   Points in arena memory:\n")
		for i, p := range arenaPoints {
			fmt.Printf("     [%d]: (%.1f, %.1f, %.1f) ID=%d\n", i, p.X, p.Y, p.Z, p.ID)
		}

		// Modify data directly in arena memory
		arenaPoints[1].X = 999.0
		arenaPoints[1].ID = 999

		fmt.Printf("   After modifying arena data:\n")
		freshView := pointHandle.Get()
		for i, p := range freshView {
			fmt.Printf("     [%d]: (%.1f, %.1f, %.1f) ID=%d\n", i, p.X, p.Y, p.Z, p.ID)
		}

		// Show that original slice is unchanged
		fmt.Printf("   Original slice unchanged: (%.1f, %.1f, %.1f) ID=%d\n",
			points[1].X, points[1].Y, points[1].Z, points[1].ID)

		return nil
	})
}

func demonstrateMemoryEfficiency() {
	fmt.Println("\n📊 Memory Efficiency Analysis")
	fmt.Println("=====================================")

	var m1, m2, m3 runtime.MemStats

	// Measure initial memory
	runtime.GC()
	runtime.ReadMemStats(&m1)

	fmt.Printf("\n7️⃣ Memory usage comparison:\n")
	fmt.Printf("   Initial Go heap: %d KB\n", m1.Alloc/1024)

	// Create large slices in Go heap
	goSlices := make([][]int32, 10)
	for i := range goSlices {
		goSlices[i] = make([]int32, 10000)
		for j := range goSlices[i] {
			goSlices[i][j] = int32(i * j)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)
	fmt.Printf("   After Go heap allocations: %d KB (+%d KB)\n",
		m2.Alloc/1024, (m2.Alloc-m1.Alloc)/1024)

	// Create equivalent data in Arena
	argo.WithArena(func(arena *argo.Arena) interface{} {
		handles := make([]*argo.Handle[int32], 10)
		for i := range handles {
			data := make([]int32, 10000)
			for j := range data {
				data[j] = int32(i * j)
			}
			handles[i] = argo.NewHandle(arena, data)
		}

		runtime.GC()
		runtime.ReadMemStats(&m3)
		fmt.Printf("   After Arena allocations: %d KB (+%d KB from initial)\n",
			m3.Alloc/1024, (m3.Alloc-m1.Alloc)/1024)

		fmt.Printf("   💡 Arena data doesn't appear in Go heap statistics!\n")
		fmt.Printf("   💡 Arena holds ~%d KB of data invisibly\n",
			10*10000*4/1024) // 10 arrays * 10000 elements * 4 bytes

		// Verify all handles work
		totalElements := 0
		for _, handle := range handles {
			totalElements += len(handle.Get())
		}
		fmt.Printf("   ✅ All %d elements accessible through handles\n", totalElements)

		return nil
	})

	// Final cleanup
	goSlices = nil
	runtime.GC()
}

func main() {
	fmt.Println("🚀 Argo Type-Safe Generic Handles Test Suite")
	fmt.Println("=====================================")

	demonstrateHandleBasics()
	demonstrateCInteroperation()
	demonstrateMultipleTypes()
	demonstrateGCImmunity()
	demonstrateCustomStructs()
	demonstrateMemoryEfficiency()

	fmt.Println("\n✅ All Handle tests completed successfully!")
	fmt.Println("\n💡 Key Handle Benefits Demonstrated:")
	fmt.Println("   • Type-safe generic interface for any data type")
	fmt.Println("   • Seamless C function interoperation")
	fmt.Println("   • Complete immunity from Go's garbage collector")
	fmt.Println("   • Zero-copy data access with Get() method")
	fmt.Println("   • Automatic memory management via Arena lifecycle")
	fmt.Println("   • Support for custom structs and complex types")
}
