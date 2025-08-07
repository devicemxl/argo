package main

import (
	"context"
	"fmt"
	"time"

	"github.com/devicemxl/argo"
)

// simulateLongRunningOperation simula una tarea que procesa datos en un bucle
// y devuelve un booleano para indicar si la operación finalizó con éxito.
func simulateLongRunningOperation(ctx context.Context, arena *argo.Arena, totalIterations int) bool {
	fmt.Printf("Iniciando una operación con %d iteraciones.\n", totalIterations)
	startTime := time.Now()

	for i := 0; i < totalIterations; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("⚠️ Operación cancelada por el contexto después de %d iteraciones. Razón: %s\n", i, ctx.Err())
			// Retornamos false para indicar que la operación no se completó
			return false
		default:
			goStr := fmt.Sprintf("data-%d", i)
			arena.CString(goStr)
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("✅ Operación completada con éxito en %s.\n", elapsed)
	// Retornamos true para indicar que la operación se completó
	return true
}

func main() {
	// --- Test 1: Timeout ---
	fmt.Println("\n=== Test 1: Operación con Timeout (esperamos cancelación) ===")
	timeout := 50 * time.Millisecond
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), timeout)
	defer cancelTimeout()

	// La función anónima ahora devuelve un bool, que es el resultado de la función.
	resultTimeout := argo.WithArenaContext(ctxTimeout, func(ctx context.Context, arena *argo.Arena) bool {
		return simulateLongRunningOperation(ctx, arena, 1000000)
	})

	if resultTimeout {
		fmt.Println("Resultado del test: La operación terminó con éxito.")
	} else {
		fmt.Println("Resultado del test: La operación fue cancelada por un timeout.")
	}

	// --- Test 2: Finalización exitosa ---
	fmt.Println("\n=== Test 2: Operación que finaliza antes del Timeout (esperamos éxito) ===")
	successTimeout := 5 * time.Second
	ctxSuccess, cancelSuccess := context.WithTimeout(context.Background(), successTimeout)
	defer cancelSuccess()

	// La función anónima también debe devolver un bool.
	successResult := argo.WithArenaContext(ctxSuccess, func(ctx context.Context, arena *argo.Arena) bool {
		return simulateLongRunningOperation(ctx, arena, 100000)
	})

	if successResult {
		fmt.Println("Resultado del test: La operación terminó con éxito.")
	} else {
		fmt.Println("Resultado del test: La operación fue cancelada por un timeout.")
	}
}
