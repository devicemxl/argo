package main

/*
#cgo CFLAGS: -std=c99 -O2
#include "../cgo_utils.h" // La ruta debe ser relativa a este archivo main.go
*/
import "C" // Importa el paquete C para las funciones CGO

import (
	"context"
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/devicemxl/argo" // Importa tu paquete argo
)

// === Ejemplos de uso de Argo + CGO ===

// Ejemplo 1: Procesamiento básico mejorado
func advancedStringProcessing() {
	fmt.Println("=== Advanced String Processing ===")

	result := argo.WithArena(func(arena *argo.Arena) map[string]string {
		input := "Hello World"

		// Convertir a C string usando Argo
		cstr := argo.CString(input)

		results := make(map[string]string)

		// Procesar con múltiples funciones C
		cUpper := C.to_uppercase((*C.char)(unsafe.Pointer(cstr)))
		defer C.free(unsafe.Pointer(cUpper))
		results["uppercase"] = argo.GoString((*byte)(unsafe.Pointer(cUpper)))

		cLower := C.to_lowercase((*C.char)(unsafe.Pointer(cstr)))
		defer C.free(unsafe.Pointer(cLower))
		results["lowercase"] = argo.GoString((*byte)(unsafe.Pointer(cLower)))

		cRepeated := C.process_string((*C.char)(unsafe.Pointer(cstr)), 3)
		defer C.free(unsafe.Pointer(cRepeated))
		results["repeated"] = argo.GoString((*byte)(unsafe.Pointer(cRepeated)))

		// Calcular hash
		hash := C.simple_hash((*C.char)(unsafe.Pointer(cstr)))
		results["hash"] = fmt.Sprintf("%d", hash)

		return results
	})

	for key, value := range result {
		fmt.Printf("  %s: %s\n", key, value)
	}
}

// Ejemplo 2: Procesamiento avanzado de arrays
func advancedArrayProcessing() {
	fmt.Println("\n=== Advanced Array Processing ===")

	// Usar []int en lugar de []int32 para compatibilidad con C.int
	originalArray := []int{5, 2, 8, 1, 9, 3, 7, 4, 6}

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Usar NewHandle con el array de datos
		handle := argo.NewHandle[int](arena, originalArray)

		// Obtener puntero para C
		cArrayPtr := (*C.int)(handle.Ptr())
		arrayLen := C.size_t(len(originalArray))

		fmt.Printf("  Original: %v\n", handle.Get())

		// Sum
		sum := C.sum_array(cArrayPtr, arrayLen)
		fmt.Printf("  Sum: %d", sum)

		// Max
		maxVal := C.max_array(cArrayPtr, arrayLen)
		fmt.Printf(", Max: %d", maxVal)

		// Min
		minVal := C.min_array(cArrayPtr, arrayLen)
		fmt.Printf(", Min: %d\n", minVal)

		// Sort (opera in-place en el array de la arena)
		C.sort_array(cArrayPtr, arrayLen)
		fmt.Printf("  Sorted: %v\n", handle.Get())

		// Process (opera in-place en el array de la arena)
		C.process_array(cArrayPtr, arrayLen, 10)
		fmt.Printf("  Processed (x10): %v\n", handle.Get())

		return nil
	})
}

// Ejemplo 3: Generación y procesamiento de datos grandes
func bigDataProcessing() {
	fmt.Println("\n=== Big Data Processing ===")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		size := 1024 * 1024 // 1MB

		start := time.Now()

		// Generar datos en C
		cdata := C.generate_data(C.size_t(size))
		defer C.free(unsafe.Pointer(cdata))

		// Copiar a Go usando la arena
		godata := argo.GoBytes((*byte)(unsafe.Pointer(cdata)), size)

		genTime := time.Since(start)

		fmt.Printf("Generated %d bytes in %v\n", len(godata), genTime)
		fmt.Printf("First 20 bytes: %s\n", string(godata[:20]))
		fmt.Printf("Last 20 bytes: %s\n", string(godata[len(godata)-20:]))

		// Calcular hash de los datos
		start = time.Now()
		cstr := argo.CString(string(godata[:1000])) // Hash de los primeros 1000 bytes
		hash := C.simple_hash((*C.char)(unsafe.Pointer(cstr)))
		hashTime := time.Since(start)

		fmt.Printf("Hash of first 1000 bytes: %d (calculated in %v)\n", hash, hashTime)

		return nil
	})
}

// Ejemplo 4: Uso con contexto y timeout
func contextualProcessing() {
	fmt.Println("\n=== Contextual Processing ===")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := argo.WithArenaContext(ctx, func(ctx context.Context, arena *argo.Arena) bool {
		// Simular procesamiento pesado
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				fmt.Printf("Processing cancelled at step %d\n", i)
				return false
			default:
				// Simular trabajo con C
				data := fmt.Sprintf("step_%d", i)
				cstr := argo.CString(data)

				// Procesar con C
				cresult := C.process_string((*C.char)(unsafe.Pointer(cstr)), 2)
				defer C.free(unsafe.Pointer(cresult))

				result := argo.GoString((*byte)(unsafe.Pointer(cresult)))
				fmt.Printf("Step %d: %s\n", i, result)

				time.Sleep(200 * time.Millisecond)
			}
		}
		return true
	})

	fmt.Printf("Processing completed successfully: %v\n", result)
}

// Ejemplo 5: Procesamiento de arrays de strings
func stringArrayProcessing() {
	fmt.Println("\n=== String Array Processing ===")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		strings := []string{"hello", "world", "from", "go", "and", "c"}

		// Convertir array de strings a C
		cstrings := argo.CStringArray(strings)

		fmt.Printf("Original strings: %v\n", strings)

		// Procesar cada string individualmente
		results := make(map[string]string)

		for i, str := range strings {
			// Obtener puntero al string C
			cstr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cstrings)) + uintptr(i)*unsafe.Sizeof((*C.char)(nil))))

			// Procesar con diferentes funciones
			cUpper := C.to_uppercase(cstr)
			defer C.free(unsafe.Pointer(cUpper))

			results[str] = argo.GoString((*byte)(unsafe.Pointer(cUpper)))
		}

		fmt.Printf("Processed strings:\n")
		for orig, processed := range results {
			fmt.Printf("  %s -> %s\n", orig, processed)
		}

		return nil
	})
}

// Ejemplo 6: Criptografía simple
func cryptoProcessing() {
	fmt.Println("\n=== Crypto Processing ===")
	original := "Hello, Secret World!"
	shift := 13 // ROT13

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Encrypt
		cOriginal := argo.CString(original)
		cEncrypted := C.caesar_cipher((*C.char)(unsafe.Pointer(cOriginal)), C.int(shift))
		defer C.free(unsafe.Pointer(cEncrypted))
		encrypted := argo.GoString((*byte)(unsafe.Pointer(cEncrypted)))
		fmt.Printf("Original: %s\n", original)
		fmt.Printf("Encrypted: %s\n", encrypted)

		// Decrypt (para ROT13, aplicar el mismo shift otra vez)
		cEncryptedForDecryption := argo.CString(encrypted)
		cDecrypted := C.caesar_cipher((*C.char)(unsafe.Pointer(cEncryptedForDecryption)), C.int(shift))
		defer C.free(unsafe.Pointer(cDecrypted))
		decrypted := argo.GoString((*byte)(unsafe.Pointer(cDecrypted)))
		fmt.Printf("Decrypted: %s\n", decrypted)

		return nil
	})
}

// Ejemplo 7: Operaciones matemáticas avanzadas
func mathProcessing() {
	fmt.Println("\n=== Mathematical Processing ===")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		// Pruebas matemáticas
		base := 2.5
		exp := 10

		// Calcular potencia
		result := float64(C.fast_pow(C.double(base), C.int(exp)))
		fmt.Printf("%.2f^%d = %.2f\n", base, exp, result)

		// Calcular raíz cuadrada
		x := 144.0
		sqrt := float64(C.sqrt_newton(C.double(x)))
		fmt.Printf("sqrt(%.2f) = %.2f\n", x, sqrt)

		// Calcular factorial
		n := 10
		fact := uint64(C.factorial(C.int(n)))
		fmt.Printf("%d! = %d\n", n, fact)

		// Operaciones con bits
		number := uint32(255)
		bits := int(C.count_bits(C.uint32_t(number)))
		reversed := uint32(C.reverse_bits(C.uint32_t(number)))

		fmt.Printf("Number: %d, Bits set: %d, Reversed: %d\n", number, bits, reversed)

		return nil
	})
}

// Ejemplo 8: Validación de datos
func validationProcessing() {
	fmt.Println("\n=== Data Validation ===")

	argo.WithArena(func(arena *argo.Arena) interface{} {
		testData := []string{
			"user@example.com",
			"invalid-email",
			"https://example.com",
			"not-a-url",
			"12345",
			"not-a-number",
		}

		for _, data := range testData {
			cstr := argo.CString(data)

			// Validar email
			isEmail := bool(C.is_valid_email((*C.char)(unsafe.Pointer(cstr))))

			// Validar URL
			isURL := bool(C.is_valid_url((*C.char)(unsafe.Pointer(cstr))))

			// Validar número
			isNumber := bool(C.is_valid_number((*C.char)(unsafe.Pointer(cstr))))

			fmt.Printf("'%s' -> Email: %v, URL: %v, Number: %v\n",
				data, isEmail, isURL, isNumber)
		}

		return nil
	})
}

// Ejemplo 9: Benchmark comparativo detallado
func detailedBenchmark() {
	fmt.Println("\n=== Detailed Performance Benchmark ===")

	iterations := 50000

	// Método tradicional
	start := time.Now()
	for i := 0; i < iterations; i++ {
		input := fmt.Sprintf("benchmark_test_%d", i)
		cstr := C.CString(input) // C.CString estándar

		// Múltiples operaciones
		cUpper := C.to_uppercase(cstr)
		cHash := C.simple_hash(cstr)
		cRepeated := C.process_string(cstr, 2)

		// Limpiar
		C.free(unsafe.Pointer(cstr))
		C.free(unsafe.Pointer(cUpper))
		C.free(unsafe.Pointer(cRepeated))

		_ = cHash // Usar variable
	}
	traditional := time.Since(start)

	// Método con Argo
	start = time.Now()
	argo.WithArena(func(arena *argo.Arena) interface{} {
		for i := 0; i < iterations; i++ {
			input := fmt.Sprintf("benchmark_test_%d", i)
			cstr := argo.CString(input)

			// Múltiples operaciones
			cUpper := C.to_uppercase((*C.char)(unsafe.Pointer(cstr)))
			cHash := C.simple_hash((*C.char)(unsafe.Pointer(cstr)))
			cRepeated := C.process_string((*C.char)(unsafe.Pointer(cstr)), 2)

			// Solo liberar las que vienen de C
			C.free(unsafe.Pointer(cUpper))
			C.free(unsafe.Pointer(cRepeated))

			_ = cHash // Usar variable
		}
		return nil
	})
	argoTime := time.Since(start)

	// Estadísticas de memoria
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Forzar GC
	runtime.GC()
	runtime.ReadMemStats(&m2)

	fmt.Printf("Performance Results:\n")
	fmt.Printf("  Traditional CGO: %v\n", traditional)
	fmt.Printf("  Argo CGO: %v\n", argoTime)
	fmt.Printf("  Improvement: %.2fx faster\n", float64(traditional)/float64(argoTime))
	fmt.Printf("  Memory after GC: %d KB\n", m2.Alloc/1024)
	fmt.Printf("  Total allocations: %d KB\n", m2.TotalAlloc/1024)
}

// Ejemplo 10: Monitoreo de Argo
func argoMonitoring() {
	fmt.Println("\n=== Argo Monitoring ===")

	arena := argo.NewArena()
	defer argo.Free()

	// Mostrar estadísticas iniciales
	fmt.Println("Initial stats:")
	argo.PrintArenaStats()

	// Realizar algunas operaciones
	for i := 0; i < 100; i++ {
		input := fmt.Sprintf("monitoring_test_%d", i)
		cstr := argo.CString(input)

		cUpper := C.to_uppercase((*C.char)(unsafe.Pointer(cstr)))
		C.free(unsafe.Pointer(cUpper))
	}

	// Mostrar estadísticas finales
	fmt.Println("\nFinal stats:")
	argo.PrintArenaStats()
}

func main() {
	fmt.Println("=== Argo + CGO Complete Examples ===")

	advancedStringProcessing()
	advancedArrayProcessing()
	bigDataProcessing()
	contextualProcessing()
	stringArrayProcessing()
	cryptoProcessing()
	mathProcessing()
	validationProcessing()
	detailedBenchmark()
	argoMonitoring()

	fmt.Println("\n=== All examples completed ===")
}
