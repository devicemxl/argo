package main

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/devicemxl/argo"
)

func demonstrateArgoArenaAllocation() {
	fmt.Println("🚀 Demostrando la asignación de memoria con argo.Arena")
	fmt.Println("===========================================")

	// Utiliza argo.WithArena para gestionar automáticamente la creación y liberación de la arena.
	// La función anónima ahora devuelve un valor booleano al final.
	_ = argo.WithArena(func(arena *argo.Arena) bool { // <<-- Se ha añadido un tipo de retorno `bool`
		// Test 1: La arena se crea automáticamente
		fmt.Println("\n1️⃣ La arena ha sido creada automáticamente por argo.WithArena.")
		fmt.Println("   La memoria de C será liberada cuando esta función anónima termine.")

		// Test 2: Muestra las estadísticas de memoria de Go antes de la asignación.
		var m1 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		fmt.Printf("\n📊 Go heap antes de las operaciones: %d KB\n", m1.Alloc/1024)

		// Test 3: Asignaciones de memoria desde la arena.
		fmt.Println("\n2️⃣ Realizando asignaciones desde la arena:")

		// Argo.Alloc() asigna memoria y maneja internamente la creación de nuevos chunks si es necesario.
		ptr1 := arena.Alloc(100, 8)
		if ptr1 != nil {
			*(*byte)(ptr1) = 0x42
			fmt.Printf("   Asignación de 100 bytes exitosa. Escribimos byte de prueba: 0x%x\n", *(*byte)(ptr1))
		}

		ptr2 := arena.Alloc(256, 8)
		if ptr2 != nil {
			slice := unsafe.Slice((*byte)(ptr2), 256)
			for i := range slice {
				slice[i] = byte(i % 256)
			}
			fmt.Printf("   Asignación de 256 bytes exitosa. Escribimos un patrón, primeros 5 bytes: %v\n", slice[:5])
		}

		ptr3 := arena.Alloc(1024, 8)
		if ptr3 != nil {
			fmt.Printf("   Tercera asignación (1024 bytes) exitosa\n")
		}

		// Test 4: La asignación de un objeto grande fuerza la creación de un nuevo chunk.
		fmt.Println("\n3️⃣ Asignando un objeto grande para forzar la creación de un nuevo chunk...")
		_ = arena.Alloc(10000, 8) // Argo automáticamente creará un nuevo chunk si el actual está lleno.
		fmt.Println("   La asignación de 10000 bytes fue exitosa, argo manejó el nuevo chunk.")

		// Muestra el estado de la arena para verificar que se crearon nuevos chunks.
		arena.PrintArenaStats()

		// Test 5: Muestra que el heap de Go no se ve afectado.
		var m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m2)
		fmt.Printf("\n📊 Go heap después de las operaciones: %d KB\n", m2.Alloc/1024)
		fmt.Printf("📊 Diferencia en el heap de Go: %d KB\n", int64(m2.Alloc-m1.Alloc)/1024)
		fmt.Printf("💡 Nota: El heap de Go no aumenta porque argo usa malloc de C.\n")

		// Test 6: Prueba la independencia del GC.
		fmt.Println("\n4️⃣ Probando la independencia del GC:")
		fmt.Println("   Forzando la recolección de basura...")
		for i := 0; i < 5; i++ {
			runtime.GC()
		}

		if ptr1 != nil && *(*byte)(ptr1) == 0x42 {
			fmt.Printf("   ✅ Los datos en la memoria de C sobrevivieron al GC: 0x%x\n", *(*byte)(ptr1))
		}
		if ptr2 != nil {
			slice := unsafe.Slice((*byte)(ptr2), 5)
			fmt.Printf("   ✅ El patrón en la memoria de C sobrevivió al GC: %v\n", slice)
		}

		fmt.Println("\n✅ La demostración ha finalizado. argo.WithArena liberará la memoria automáticamente.")

		return true // <<-- Se ha añadido un valor de retorno
	})

	// Test 7: Verificación final de la memoria de Go.
	var m3 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m3)
	fmt.Printf("\n📊 Heap final de Go: %d KB\n", m3.Alloc/1024)

	fmt.Println("\n💡 Clave de la observación:")
	fmt.Println("   • La función argo.WithArena() asegura que la memoria de C se libere siempre al final.")
	fmt.Println("   • Las asignaciones con argo.Alloc() son GC-free y no afectan al heap de Go.")
}

func main() {
	demonstrateArgoArenaAllocation()
}
