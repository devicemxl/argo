package main

import (
	"sync"

	"github.com/devicemxl/argo" // Tu paquete argo
)

func main() {
	// Create a new arena
	arena := argo.NewArena()
	defer arena.Free()

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Launch multiple goroutines to perform operations on the arena
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Safe concurrent access to arena
			data := id * id
			print("goroutine ", id, " result: ", data, "\n")
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}
