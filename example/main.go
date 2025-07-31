package main

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/devicemxl/argo" // Importa tu paquete argo
)

// Test para verificar que la arena está libre del GC
func testGCFreedom() {
	fmt.Println("=== Testing GC Freedom ===")

	var m1, m2, m3 runtime.MemStats

	// Estadísticas iniciales
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Crear arena y asignar mucha memoria
	arena := argo.NewArena()

	// Asignar 10MB de datos en la arena
	const dataSize = 10 * 1024 * 1024
	const iterations = 1000

	ptrs := make([]unsafe.Pointer, iterations)

	for i := 0; i < iterations; i++ {
		// Asignar bloques de 10KB cada uno
		ptr := arena.Alloc(dataSize/iterations, 8)
		ptrs[i] = ptr

		// Escribir datos para asegurar que la memoria es real
		if ptr != nil {
			*(*byte)(ptr) = byte(i % 256)
		}
	}

	// Estadísticas después de asignación
	runtime.GC()
	runtime.ReadMemStats(&m2)

	fmt.Printf("Go heap antes: %d KB\n", m1.Alloc/1024)
	fmt.Printf("Go heap después de arena: %d KB\n", m2.Alloc/1024)
	fmt.Printf("Diferencia en heap Go: %d KB\n", int64(m2.Alloc-m1.Alloc)/1024)

	// Verificar que los datos siguen siendo accesibles
	fmt.Printf("Verificando acceso a datos...\n")
	validAccess := 0
	for i, ptr := range ptrs {
		if ptr != nil && *(*byte)(ptr) == byte(i%256) {
			validAccess++
		}
	}
	fmt.Printf("Accesos válidos: %d/%d\n", validAccess, iterations)

	// Forzar múltiples GC para ver si la memoria de la arena se ve afectada
	fmt.Printf("Ejecutando múltiples GC...\n")
	for i := 0; i < 10; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	// Verificar nuevamente acceso después de GC intensivo
	validAccessAfterGC := 0
	for i, ptr := range ptrs {
		if ptr != nil && *(*byte)(ptr) == byte(i%256) {
			validAccessAfterGC++
		}
	}
	fmt.Printf("Accesos válidos después de GC: %d/%d\n", validAccessAfterGC, iterations)

	// Liberar arena
	arena.Free()

	// Estadísticas finales
	runtime.GC()
	runtime.ReadMemStats(&m3)

	fmt.Printf("Go heap después de liberar arena: %d KB\n", m3.Alloc/1024)
	fmt.Printf("Diferencia heap Go (inicio vs final): %d KB\n", int64(m3.Alloc-m1.Alloc)/1024)

	// Los punteros ahora deberían ser inválidos, pero no podemos probarlos
	// sin riesgo de crash

	if validAccess == iterations && validAccessAfterGC == iterations {
		fmt.Printf("✅ SUCCESS: Arena parece estar libre del GC\n")
	} else {
		fmt.Printf("❌ FAIL: Arena puede estar siendo afectada por GC\n")
	}

	// Test adicional: verificar que no hay memory leaks en Go
	if int64(m3.Alloc-m1.Alloc) < 1024*100 { // Menos de 100KB de diferencia
		fmt.Printf("✅ SUCCESS: No hay memory leaks significativos en Go heap\n")
	} else {
		fmt.Printf("❌ WARNING: Posible memory leak en Go heap\n")
	}
}

// Test de benchmark comparativo
func benchmarkGCImpact() {
	fmt.Println("\n=== Benchmark: GC Impact ===")

	iterations := 100000

	// Test con asignaciones normales de Go (sujetas al GC)
	start := time.Now()
	var normalPtrs [][]byte
	for i := 0; i < iterations; i++ {
		data := make([]byte, 1024) // 1KB cada uno
		data[0] = byte(i % 256)
		normalPtrs = append(normalPtrs, data)

		if i%10000 == 0 {
			runtime.GC() // Forzar GC periódicamente
		}
	}
	normalTime := time.Since(start)

	// Test con arena (libre del GC)
	start = time.Now()
	arena := argo.NewArena()
	var arenaPtrs []unsafe.Pointer
	for i := 0; i < iterations; i++ {
		ptr := arena.Alloc(1024, 1) // 1KB cada uno
		if ptr != nil {
			*(*byte)(ptr) = byte(i % 256)
			arenaPtrs = append(arenaPtrs, ptr)
		}

		if i%10000 == 0 {
			runtime.GC() // Forzar GC, no debería afectar arena
		}
	}
	arenaTime := time.Since(start)
	arena.Free()

	fmt.Printf("Tiempo con Go heap (con GC): %v\n", normalTime)
	fmt.Printf("Tiempo con Arena (sin GC): %v\n", arenaTime)
	fmt.Printf("Mejora: %.2fx\n", float64(normalTime)/float64(arenaTime))

	// Limpiar para evitar OOM
	normalPtrs = nil
	runtime.GC()
}

func main() {
	testGCFreedom()
	benchmarkGCImpact()
}
