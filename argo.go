package argo

/*
#cgo CFLAGS: -std=c99 -O2
#cgo pkg-config:
#include "cgo_utils.h"
*/
import "C"

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	pageSize      = 4096
	minChunkSize  = 8192
	ptrAlign      = 8
	maxObjectSize = 1 << 20 // 1MB
)

// chunk representa un bloque de memoria contigua
type chunk struct {
	data   []byte  // Mantener referencia al slice para evitar que el GC lo libere
	base   uintptr // Dirección base
	size   uintptr // Tamaño total
	offset uintptr // Offset actual para asignaciones
	next   *chunk  // Siguiente chunk en la lista
}

// Arena implementa una zona de memoria para asignación manual
type Arena struct {
	mu     sync.Mutex
	chunks *chunk
	freed  uint32 // atomic flag
	stats  arenaStats
}

type arenaStats struct {
	totalAlloc   uint64
	totalChunks  uint32
	currentUsage uint64
}

// NewArena crea una nueva arena
func NewArena() *Arena {
	a := &Arena{}
	a.newChunk(minChunkSize)
	return a
}

// newChunk crea un nuevo chunk de memoria
func (a *Arena) newChunk(size uintptr) {
	// Alinear al tamaño de página
	size = (size + pageSize - 1) &^ (pageSize - 1)
	if size < minChunkSize {
		size = minChunkSize
	}

	// Crear el slice y mantener la referencia
	data := make([]byte, size)

	newChunk := &chunk{
		data:   data, // CRÍTICO: mantener referencia al slice
		base:   uintptr(unsafe.Pointer(&data[0])),
		size:   size,
		offset: 0,
		next:   a.chunks,
	}

	a.chunks = newChunk
	a.stats.totalChunks++
	a.stats.totalAlloc += uint64(size)
}

// Free libera toda la memoria de la arena
func (a *Arena) Free() {
	if !atomic.CompareAndSwapUint32(&a.freed, 0, 1) {
		return // Ya liberado
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Limpiar todos los chunks
	current := a.chunks
	for current != nil {
		next := current.next
		// Limpiar referencias
		current.data = nil
		current.base = 0
		current.size = 0
		current.offset = 0
		current.next = nil
		current = next
	}

	a.chunks = nil
	a.stats = arenaStats{}
}

// New crea un nuevo objeto de tipo T en la arena (conservando el nombre original)
func New[T any](a *Arena) *T {
	if atomic.LoadUint32(&a.freed) == 1 {
		panic("arena: use after free")
	}

	var zero T
	typ := reflect.TypeOf(zero)
	size := typ.Size()
	align := uintptr(typ.Align())

	ptr := a.Alloc(size, align)
	if ptr == nil {
		panic("arena: allocation failed")
	}

	// Inicializar con valor zero
	result := (*T)(ptr)
	*result = zero

	return result
}

// MakeSlice crea un slice de tipo T en la arena (conservando el nombre original)
func MakeSlice[T any](a *Arena, len, cap int) []T {
	if atomic.LoadUint32(&a.freed) == 1 {
		panic("arena: use after free")
	}

	if len < 0 || cap < 0 || len > cap {
		panic("arena: invalid slice dimensions")
	}

	if cap == 0 {
		return nil
	}

	var zero T
	elemSize := reflect.TypeOf(zero).Size()
	totalSize := uintptr(cap) * elemSize

	ptr := a.Alloc(totalSize, ptrAlign)
	if ptr == nil {
		panic("arena: allocation failed")
	}

	// Crear el slice header
	slice := (*reflect.SliceHeader)(unsafe.Pointer(&[]T{}))
	slice.Data = uintptr(ptr)
	slice.Len = len
	slice.Cap = cap

	return *(*[]T)(unsafe.Pointer(slice))
}

// Alloc es la función pública de asignación (conservando el nombre original)
func (a *Arena) Alloc(size, align uintptr) unsafe.Pointer {
	if size == 0 {
		return nil
	}

	if size > maxObjectSize {
		panic("arena: object too large")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	current := a.chunks
	if current == nil {
		panic("arena: no chunks available")
	}

	// Alinear el offset
	alignedOffset := (current.offset + align - 1) &^ (align - 1)

	// Verificar si hay suficiente espacio
	if alignedOffset+size > current.size {
		// Crear nuevo chunk
		newSize := uintptr(minChunkSize)
		if size+align > uintptr(minChunkSize) {
			newSize = size + align + pageSize
		}
		a.newChunk(newSize)
		current = a.chunks
		alignedOffset = (current.offset + align - 1) &^ (align - 1)
	}

	ptr := unsafe.Pointer(current.base + alignedOffset)
	current.offset = alignedOffset + size
	a.stats.currentUsage += uint64(size)

	return ptr
}

// Clone hace una copia del valor fuera de la arena (similar a la API oficial)
func Clone[T any](value T) T {
	// Para tipos simples, simplemente devolver el valor
	// Para tipos con punteros, necesitaríamos una implementación más compleja
	return value
}

// Handle representa un puntero manejado a datos en la arena
type Handle[T any] struct {
	ptr   unsafe.Pointer
	arena *Arena
	len   int
}

// NewHandle crea un nuevo handle para un array en la arena
func NewHandle[T any](a *Arena, data []T) *Handle[T] {
	if atomic.LoadUint32(&a.freed) == 1 {
		panic("arena: use after free")
	}

	if len(data) == 0 {
		return &Handle[T]{
			ptr:   nil,
			arena: a,
			len:   0,
		}
	}

	// Crear slice en la arena
	arenaSlice := MakeSlice[T](a, len(data), len(data))

	// Copiar datos
	copy(arenaSlice, data)

	return &Handle[T]{
		ptr:   unsafe.Pointer(&arenaSlice[0]),
		arena: a,
		len:   len(data),
	}
}

// Ptr devuelve el puntero unsafe
func (h *Handle[T]) Ptr() unsafe.Pointer {
	return h.ptr
}

// Get devuelve el slice de Go
func (h *Handle[T]) Get() []T {
	if h.ptr == nil {
		return nil
	}

	sliceHeader := reflect.SliceHeader{
		Data: uintptr(h.ptr),
		Len:  h.len,
		Cap:  h.len,
	}

	return *(*[]T)(unsafe.Pointer(&sliceHeader))
}

// Stats devuelve las estadísticas de la arena
func (a *Arena) Stats() arenaStats {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stats
}

// PrintArenaStats imprime las estadísticas (conservando el nombre original)
func (a *Arena) PrintArenaStats() {
	stats := a.Stats()
	fmt.Printf("Arena Stats: Total Chunks: %d, Total Allocated: %d bytes, Current Usage: %d bytes\n",
		stats.totalChunks, stats.totalAlloc, stats.currentUsage)
}

// WithArena ejecuta una función con una arena que se libera automáticamente
func WithArena[T any](f func(*Arena) T) T {
	a := NewArena()
	defer a.Free()

	// Mantener la arena viva durante la ejecución
	runtime.KeepAlive(a)

	return f(a)
}

// WithArenaContext similar a WithArena pero con contexto
func WithArenaContext[T any](ctx context.Context, f func(context.Context, *Arena) T) T {
	a := NewArena()
	defer a.Free()

	runtime.KeepAlive(a)

	return f(ctx, a)
}

// CString converts a Go string to a C string (char*).
// The returned C string is allocated in the Go heap and MUST be freed by C.free.
func CString(s string) *C.char {
	return C.CString(s)
}

// GoString converts a C string (char*) to a Go string.
// It assumes the C string is null-terminated.
func GoString(cptr *C.char) string {
	if cptr == nil {
		return ""
	}
	return C.GoString(cptr)
}

// GoBytes converts a C byte array (void* or char*) of a given length to a Go byte slice.
func GoBytes(cptr *C.char, length int) []byte {
	if cptr == nil || length <= 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(cptr), C.int(length))
}

// CStringArray converts a Go string slice to a C array of char* (char**).
// Each C string and the array of pointers itself are allocated in the C heap
// and MUST be freed by calling C.free on each string and then on the array pointer.
//
// Alternatively, and preferably in the context of an Arena, you could implement a
// version that allocates these strings within the Arena itself, but for general CGO
// interoperability, this standard approach is more common.
// Given your Arena, a better approach might be to allocate each C string within the arena,
// and then create the `char**` array also in the arena.
// For now, I'll provide the standard C.CString approach which allocates in the Go heap.
func CStringArray(goStrings []string) **C.char {
	if len(goStrings) == 0 {
		return nil
	}

	// Allocate a C array of char*
	// Using C.malloc for the array of pointers, so it can be passed to C functions
	// The individual strings converted by C.CString are allocated by Go runtime
	// and need to be freed individually using C.free.
	cArray := C.malloc(C.size_t(len(goStrings)) * C.sizeof_char_ptr)
	if cArray == nil {
		panic("Failed to allocate C array for strings")
	}

	// Cast to char**
	cStrings := (**C.char)(cArray)

	// Populate the C array with C strings
	for i, s := range goStrings {
		// Allocate individual C string
		// NOTE: C.CString allocates memory that must be freed with C.free
		// This is outside the arena management.
		cs := C.CString(s)
		// Set the pointer in the C array
		*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStrings)) + uintptr(i)*unsafe.Sizeof((*C.char)(nil)))) = cs
	}

	return cStrings
}

// FreeCStringArray is a helper to free the memory allocated by CStringArray
func FreeCStringArray(cStrings **C.char, count int) {
	if cStrings == nil {
		return
	}
	for i := 0; i < count; i++ {
		strPtr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStrings)) + uintptr(i)*unsafe.Sizeof((*C.char)(nil))))
		C.free(unsafe.Pointer(strPtr))
	}
	C.free(unsafe.Pointer(cStrings))
}
