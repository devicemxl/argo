package argo

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Constantes
const (
	pageSize      = 4096
	minChunkSize  = 8192
	ptrAlign      = 8
	maxObjectSize = 1 << 20 // 1MB
)

// chunk representa un bloque de memoria C (fuera del GC)
type chunk struct {
	base   uintptr // Puntero a memoria C
	size   uintptr
	offset uintptr
	next   *chunk
}

// Arena completamente libre del GC
type Arena struct {
	mu     sync.Mutex
	chunks *chunk
	freed  uint32 // atomic
	stats  arenaStats
}

type arenaStats struct {
	totalAlloc   uint64
	totalChunks  uint32
	currentUsage uint64
}

type ArenaOption func(*Arena)

// NewArena crea una nueva arena completamente libre del GC
func NewArena(opts ...ArenaOption) *Arena {
	a := &Arena{}

	// Aplicar opciones
	for _, opt := range opts {
		opt(a)
	}

	// Crear primer chunk usando malloc de C
	a.newChunkC(minChunkSize)

	// Importante: Evitar que el GC mueva esta estructura
	runtime.SetFinalizer(a, (*Arena).Free)

	return a
}

// newChunkC asigna un nuevo chunk usando malloc de C (fuera del GC)
func (a *Arena) newChunkC(size uintptr) {
	// Alinear al tamaño de página
	size = (size + pageSize - 1) &^ (pageSize - 1)
	if size < minChunkSize {
		size = minChunkSize
	}

	// Asignar memoria usando malloc de C (completamente fuera del GC)
	cptr := C.malloc(C.size_t(size))
	if cptr == nil {
		panic("arena: failed to allocate memory")
	}

	// Limpiar la memoria
	C.memset(cptr, 0, C.size_t(size))

	newChunk := &chunk{
		base:   uintptr(cptr),
		size:   size,
		offset: 0,
	}

	// Agregar al frente de la lista
	newChunk.next = a.chunks
	a.chunks = newChunk

	a.stats.totalChunks++
	a.stats.totalAlloc += uint64(size)
}

// Free libera toda la memoria C asignada
func (a *Arena) Free() {
	if atomic.LoadUint32(&a.freed) == 1 {
		return // Ya liberado
	}

	atomic.StoreUint32(&a.freed, 1)

	a.mu.Lock()
	defer a.mu.Unlock()

	// Liberar todos los chunks usando free de C
	current := a.chunks
	for current != nil {
		if current.base != 0 {
			C.free(unsafe.Pointer(current.base))
		}
		next := current.next
		current.base = 0
		current.size = 0
		current.offset = 0
		current.next = nil
		current = next
	}

	a.chunks = nil
	a.stats = arenaStats{}

	// Remover finalizer
	runtime.SetFinalizer(a, nil)
}

// Alloc asigna memoria completamente fuera del GC
func (a *Arena) Alloc(size uintptr, align uintptr) unsafe.Pointer {
	if atomic.LoadUint32(&a.freed) == 1 {
		panic("arena: use after free")
	}

	if size == 0 {
		return nil
	}

	if align == 0 {
		align = 1
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	current := a.chunks
	if current == nil {
		a.newChunkC(max(minChunkSize, size+align))
		current = a.chunks
	}

	// Alinear el offset
	alignedOffset := (current.offset + align - 1) &^ (align - 1)

	// Verificar espacio disponible
	if alignedOffset+size > current.size {
		// Crear nuevo chunk más grande si es necesario
		newSize := max(minChunkSize, size+align)
		a.newChunkC(newSize)
		current = a.chunks
		alignedOffset = (current.offset + align - 1) &^ (align - 1)

		if alignedOffset+size > current.size {
			return nil // No debería pasar con el cálculo correcto
		}
	}

	ptr := unsafe.Pointer(current.base + alignedOffset)
	current.offset = alignedOffset + size

	// Actualizar estadísticas
	a.stats.currentUsage += uint64(size)

	return ptr
}

// CString crea un C string en memoria libre del GC
func (a *Arena) CString(s string) *C.char {
	if len(s) == 0 {
		nullPtr := a.Alloc(1, 1)
		*(*C.char)(nullPtr) = 0
		return (*C.char)(nullPtr)
	}

	size := uintptr(len(s)) + 1
	cptr := a.Alloc(size, 1)
	if cptr == nil {
		return nil
	}

	// Copiar string usando funciones C para evitar que Go toque la memoria
	C.memcpy(cptr, unsafe.Pointer(unsafe.StringData(s)), C.size_t(len(s)))

	// Agregar terminador nulo
	*(*C.char)(unsafe.Pointer(uintptr(cptr) + uintptr(len(s)))) = 0

	return (*C.char)(cptr)
}

// GoString convierte C string de la arena a Go string
func (a *Arena) GoString(cstr *C.char) string {
	if cstr == nil {
		return ""
	}

	// Usar strlen de C para encontrar la longitud
	length := int(C.strlen(cstr))
	if length == 0 {
		return ""
	}

	// Crear slice sin que Go copie los datos
	return unsafe.String((*byte)(unsafe.Pointer(cstr)), length)
}

// GoBytes copia datos C a slice de Go
func (a *Arena) GoBytes(cptr unsafe.Pointer, length int) []byte {
	if cptr == nil || length <= 0 {
		return nil
	}

	// Crear slice de Go y copiar datos
	result := make([]byte, length)
	C.memcpy(unsafe.Pointer(&result[0]), cptr, C.size_t(length))

	return result
}

// Handle mejorado para arrays completamente fuera del GC
type Handle[T any] struct {
	ptr      unsafe.Pointer
	arena    *Arena
	length   int
	elemSize uintptr
}

// NewHandle crea un handle con datos en memoria libre del GC
func NewHandle[T any](arena *Arena, slice []T) *Handle[T] {
	if len(slice) == 0 {
		return &Handle[T]{arena: arena}
	}

	var dummy T
	elemSize := unsafe.Sizeof(dummy)
	totalSize := uintptr(len(slice)) * elemSize

	// Determinar alineación apropiada
	align := elemSize
	if align > 8 {
		align = 8
	}

	// Asignar memoria fuera del GC
	ptr := arena.Alloc(totalSize, align)
	if ptr == nil {
		return nil
	}

	// Copiar datos usando memcpy de C para evitar interferencia del GC
	if len(slice) > 0 {
		C.memcpy(ptr, unsafe.Pointer(&slice[0]), C.size_t(totalSize))
	}

	return &Handle[T]{
		ptr:      ptr,
		arena:    arena,
		length:   len(slice),
		elemSize: elemSize,
	}
}

// Ptr devuelve el puntero C
func (h *Handle[T]) Ptr() unsafe.Pointer {
	return h.ptr
}

// Get devuelve slice de Go que apunta a memoria de la arena
func (h *Handle[T]) Get() []T {
	if h.ptr == nil || h.length == 0 {
		return nil
	}

	// Crear slice que apunta directamente a la memoria de la arena
	return unsafe.Slice((*T)(h.ptr), h.length)
}

// CStringArray crea array de C strings en memoria libre del GC
func (a *Arena) CStringArray(goStrings []string) **C.char {
	if len(goStrings) == 0 {
		return nil
	}

	// Asignar array de punteros
	ptrArraySize := uintptr(len(goStrings)) * unsafe.Sizeof((*C.char)(nil))
	cptrArray := a.Alloc(ptrArraySize, ptrAlign)
	if cptrArray == nil {
		return nil
	}

	// Convertir cada string y almacenar puntero
	ptrs := unsafe.Slice((**C.char)(cptrArray), len(goStrings))
	for i, s := range goStrings {
		ptrs[i] = a.CString(s)
	}

	return (**C.char)(cptrArray)
}

// GoStringArray convierte array de C strings a slice de Go
func (a *Arena) GoStringArray(cstrs **C.char, length int) []string {
	if cstrs == nil || length <= 0 {
		return nil
	}

	ptrs := unsafe.Slice(cstrs, length)
	result := make([]string, length)

	for i, cstr := range ptrs {
		if cstr != nil {
			result[i] = a.GoString(cstr)
		}
	}

	return result
}

// PrintArenaStats muestra estadísticas de la arena
func (a *Arena) PrintArenaStats() {
	a.mu.Lock()
	defer a.mu.Unlock()
	fmt.Printf("Arena Stats: Chunks: %d, Total: %d bytes, Used: %d bytes\n",
		a.stats.totalChunks, a.stats.totalAlloc, a.stats.currentUsage)
}

// WithArena función de conveniencia
func WithArena[T any](f func(arena *Arena) T) T {
	a := NewArena()
	defer a.Free()
	return f(a)
}

// WithArenaContext versión con contexto
func WithArenaContext[T any](ctx context.Context, f func(ctx context.Context, arena *Arena) T) T {
	a := NewArena()
	defer a.Free()
	return f(ctx, a)
}

// Función auxiliar max
func max(a, b uintptr) uintptr {
	if a > b {
		return a
	}
	return b
}

// UnsafePointer devuelve el unsafe.Pointer del CString para compatibilidad entre paquetes CGO
func (a *Arena) UnsafeCString(s string) unsafe.Pointer {
	return unsafe.Pointer(a.CString(s))
}
