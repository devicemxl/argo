package main

/*
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Simular una función C que procesa strings
char* process_data(const char* input) {
    if (input == NULL) return NULL;

    size_t len = strlen(input);
    char* result = malloc(len + 20); // Espacio extra para procesamiento

    if (result == NULL) return NULL;

    // Simular algún procesamiento: agregar prefijo y convertir a mayúsculas
    strcpy(result, "PROCESSED_");
    strcat(result, input);

    // Convertir a mayúsculas la parte original
    for (size_t i = 10; i < len + 10; i++) {
        if (result[i] >= 'a' && result[i] <= 'z') {
            result[i] = result[i] - 'a' + 'A';
        }
    }

    return result;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/devicemxl/argo" // Tu paquete argo
)

// Test tradicional con C.CString (cada string se aloca en heap de Go)
func benchmarkTraditionalCString(iterations int) time.Duration {
	fmt.Printf("Iniciando benchmark tradicional con %d iteraciones...\n", iterations)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	start := time.Now()

	results := make([]string, iterations)

	for i := 0; i < iterations; i++ {
		// Crear C string usando método tradicional (aloca en Go heap)
		goStr := fmt.Sprintf("data-%d", i)
		cstr := C.CString(goStr) // ← Esto aloca en Go heap

		// Procesar con función C
		cresult := C.process_data(cstr)

		// Convertir resultado de vuelta a Go
		results[i] = C.GoString(cresult)

		// Liberar memoria C
		C.free(unsafe.Pointer(cstr))
		C.free(unsafe.Pointer(cresult))

		// Forzar GC periódicamente para simular condiciones reales
		if i%10000 == 0 && i > 0 {
			runtime.GC()
		}
	}

	elapsed := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&m2)

	fmt.Printf("Método tradicional completado en: %v\n", elapsed)
	fmt.Printf("Memoria Go antes: %d KB, después: %d KB, diferencia: %d KB\n",
		m1.Alloc/1024, m2.Alloc/1024, int64(m2.Alloc-m1.Alloc)/1024)

	// Verificar algunos resultados
	if len(results) > 0 {
		fmt.Printf("Ejemplo resultado: %s\n", results[0])
	}

	return elapsed
}

// Test con Arena (strings se alocan fuera del GC)
func benchmarkArenaString(iterations int) time.Duration {
	fmt.Printf("Iniciando benchmark con Arena con %d iteraciones...\n", iterations)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	start := time.Now()

	result := argo.WithArena(func(arena *argo.Arena) []string {
		results := make([]string, iterations)

		for i := 0; i < iterations; i++ {
			// Crear string en memoria libre del GC
			goStr := fmt.Sprintf("data-%d", i)
			cstr := arena.CString(goStr) // ← Esto NO aloca en Go heap

			// Procesar con función C
			cresult := C.process_data((*C.char)(unsafe.Pointer(cstr)))

			// Convertir resultado de vuelta a Go
			results[i] = C.GoString(cresult)

			// Solo liberar el resultado de process_data (cstr se libera automáticamente)
			C.free(unsafe.Pointer(cresult))

			// Forzar GC periódicamente - no debería afectar las allocaciones de la arena
			if i%10000 == 0 && i > 0 {
				runtime.GC()
			}
		}

		return results
		// Arena se libera automáticamente aquí
	})

	elapsed := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&m2)

	fmt.Printf("Método Arena completado en: %v\n", elapsed)
	fmt.Printf("Memoria Go antes: %d KB, después: %d KB, diferencia: %d KB\n",
		m1.Alloc/1024, m2.Alloc/1024, int64(m2.Alloc-m1.Alloc)/1024)

	// Verificar algunos resultados
	if len(result) > 0 {
		fmt.Printf("Ejemplo resultado: %s\n", result[0])
	}

	return elapsed
}

// Test de memoria intensivo para ver el impacto del GC
func benchmarkMemoryPressure(iterations int) {
	fmt.Println("\n=== Test de Presión de Memoria ===")

	// Test 1: Muchas allocaciones pequeñas con C.CString
	fmt.Println("Test 1: C.CString con alta presión de memoria")
	start := time.Now()

	for i := 0; i < iterations; i++ {
		str := fmt.Sprintf("pressure-test-%d-with-longer-string-to-increase-memory-pressure", i)
		cstr := C.CString(str)

		// Simular uso
		_ = C.strlen(cstr)

		C.free(unsafe.Pointer(cstr))

		// Crear presión adicional en Go heap
		waste := make([]byte, 1024)
		_ = waste

		if i%1000 == 0 {
			runtime.GC()
		}
	}
	traditionalPressure := time.Since(start)

	// Test 2: Arena bajo presión
	fmt.Println("Test 2: Arena con alta presión de memoria")
	start = time.Now()

	argo.WithArena(func(arena *argo.Arena) interface{} {
		for i := 0; i < iterations; i++ {
			str := fmt.Sprintf("pressure-test-%d-with-longer-string-to-increase-memory-pressure", i)
			cstr := arena.CString(str)

			// Simular uso
			_ = C.strlen((*C.char)(unsafe.Pointer(cstr)))

			// Crear presión adicional en Go heap (arena no se ve afectada)
			waste := make([]byte, 1024)
			_ = waste

			if i%1000 == 0 {
				runtime.GC()
			}
		}
		return nil
	})
	arenaPressure := time.Since(start)

	fmt.Printf("Bajo presión de memoria:\n")
	fmt.Printf("  C.CString: %v\n", traditionalPressure)
	fmt.Printf("  Arena: %v\n", arenaPressure)
	fmt.Printf("  Mejora: %.2fx\n", float64(traditionalPressure)/float64(arenaPressure))
}

func main() {
	fmt.Println("🚀 Benchmark: C.CString vs Arena.CString")
	fmt.Println("===========================================")

	// Test con diferentes tamaños para ver el comportamiento
	testSizes := []int{10000, 50000, 1000000}

	for _, size := range testSizes {
		fmt.Printf("\n=== Benchmark con %d iteraciones ===\n", size)

		// Warm up
		fmt.Println("Realizando warm-up...")
		benchmarkTraditionalCString(1000)
		benchmarkArenaString(1000)

		// Test real
		fmt.Println("\n--- Test Real ---")
		traditionalTime := benchmarkTraditionalCString(size)

		fmt.Println() // Separador

		arenaTime := benchmarkArenaString(size)

		// Comparación
		fmt.Printf("\n📊 RESULTADOS para %d iteraciones:\n", size)
		fmt.Printf("  Método tradicional (C.CString): %v\n", traditionalTime)
		fmt.Printf("  Método Arena: %v\n", arenaTime)
		if arenaTime > 0 {
			speedup := float64(traditionalTime) / float64(arenaTime)
			if speedup > 1.0 {
				fmt.Printf("  🎉 Arena es %.2fx MÁS RÁPIDA\n", speedup)
			} else {
				fmt.Printf("  ⚠️  Arena es %.2fx más lenta\n", 1.0/speedup)
			}
		}

		// Pausa entre tests
		runtime.GC()
		time.Sleep(100 * time.Millisecond)
	}

	// Test de presión de memoria
	benchmarkMemoryPressure(20000)

	fmt.Println("\n✅ Todos los benchmarks completados!")
	fmt.Println("\n💡 La Arena debería mostrar mejores resultados especialmente:")
	fmt.Println("   - Con muchas allocaciones pequeñas")
	fmt.Println("   - Bajo alta presión de memoria/GC")
	fmt.Println("   - En aplicaciones con lifecycle largo de strings")
}
