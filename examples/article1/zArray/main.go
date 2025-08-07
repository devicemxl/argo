package main

/*
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Simulamos una función de C que procesa un array de enteros.
// En un escenario real, esta función haría algún cálculo con los datos.
// Aquí solo imprimimos un mensaje para simular la operación.
void process_array(int* data, size_t length) {
    // No hacemos nada con los datos, solo simulamos que la función existe
    // printf("Processing array of size %zu from C...\n", length);
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"github.com/devicemxl/argo"
)

// processTraditional simula el procesamiento de un array de manera tradicional en Go.
// Cada iteración del bucle crea un nuevo slice, lo que genera basura y sobrecarga al GC.
func processTraditional(data []int32) {
	fmt.Println("🚀 Iniciando processTraditional...")
	start := time.Now()

	for i := 0; i < 10000; i++ {
		// Asignación de memoria en el heap de Go. Esto es lento y crea basura.
		cdata := make([]C.int, len(data))
		for j, v := range data {
			cdata[j] = C.int(v)
		}
		// Llamada a la función de C
		C.process_array((*C.int)(unsafe.Pointer(&cdata[0])), C.size_t(len(data)))
	}
	elapsed := time.Since(start)
	fmt.Printf("✅ processTraditional finalizado en %v\n", elapsed)
}

// processWithArgo utiliza la arena para una única asignación de memoria,
// evitando la sobrecarga del GC.
func processWithArgo(data []int32) {
	fmt.Println("🚀 Iniciando processWithArgo...")
	start := time.Now()

	// argo.WithArena asegura que la memoria de la arena se libere al final de la función.
	argo.WithArena(func(arena *argo.Arena) bool {
		// argo.NewHandle crea una única asignación en la memoria de la arena (fuera del GC).
		// Esta memoria se reutiliza en todas las iteraciones.
		handle := argo.NewHandle(arena, data)
		if handle == nil {
			fmt.Println("Error: No se pudo crear el handle.")
			return false
		}

		for i := 0; i < 10000; i++ {
			// Llamada a la función de C usando el puntero del handle.
			C.process_array((*C.int)(handle.Ptr()), C.size_t(len(data)))
		}

		return true
	})

	elapsed := time.Since(start)
	fmt.Printf("✅ processWithArgo finalizado en %v\n", elapsed)
}

func main() {
	fmt.Println("=== Comparación de procesamiento de arrays ===")

	// Datos de prueba
	data := make([]int32, 1000)
	for i := range data {
		data[i] = int32(i)
	}

	// Ejecutar la prueba tradicional
	processTraditional(data)
	fmt.Println("------------------------------------------")

	// Forzar el GC para que las mediciones sean más justas
	runtime.GC()

	// Ejecutar la prueba con argo
	processWithArgo(data)
	fmt.Println("------------------------------------------")
}
