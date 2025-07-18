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
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Constantes basadas en la implementación interna de Go
const (
	// Tamaño de página del sistema (típicamente 4KB)
	pageSize = 4096

	// Tamaño mínimo de chunk
	minChunkSize = 8192

	// Alineación de memoria
	ptrAlign = 8

	// Tamaño máximo de objeto individual
	maxObjectSize = 1 << 20 // 1MB
)

// chunk representa un bloque de memoria contigua
type chunk struct {
	base   uintptr
	size   uintptr
	offset uintptr
	next   *chunk
}

// Arena mejorada basada en la implementación interna
type Arena struct {
	mu     sync.Mutex
	chunks *chunk
	freed  uint32 // atomic
	stats  arenaStats
}

// arenaStats mantiene estadísticas de uso
type arenaStats struct {
	totalAlloc   uint64
	totalChunks  uint32
	currentUsage uint64
}

// ArenaOption define una función para aplicar opciones a una Arena.
type ArenaOption func(*Arena)

// NewArena crea una nueva arena con configuración optimizada
func NewArena(opts ...ArenaOption) *Arena {
	a := &Arena{}

	// Aplicar opciones
	for _, opt := range opts {
		opt(a)
	}

	// Crear primer chunk
	a.newChunk(minChunkSize)

	return a
}

// newChunk asigna un nuevo chunk de memoria para la arena
func (a *Arena) newChunk(size uintptr) {
	size = (size + pageSize - 1) &^ (pageSize - 1) // Alinea al tamaño de página
	if size < minChunkSize {
		size = minChunkSize
	}

	buf := make([]byte, size) // Asigna memoria en el heap de Go
	newChunk := &chunk{
		base:   uintptr(unsafe.Pointer(&buf[0])),
		size:   size,
		offset: 0,
	}

	// Agrega el nuevo chunk al frente de la lista
	newChunk.next = a.chunks
	a.chunks = newChunk

	a.stats.totalChunks++
	a.stats.totalAlloc += uint64(size) // Contabiliza el tamaño total del chunk
}

// Free libera toda la memoria asignada por la arena.
// Debe llamarse explícitamente cuando la arena ya no se necesita.
func (a *Arena) Free() {
	if atomic.LoadUint32(&a.freed) == 1 {
		return // Ya liberado
	}

	atomic.StoreUint32(&a.freed, 1) // Marca la arena como liberada

	a.mu.Lock()
	defer a.mu.Unlock()

	// Desreferencia los chunks para que el GC de Go pueda reclamar la memoria
	current := a.chunks
	for current != nil {
		// No necesitamos liberar explícitamente con C.free porque make([]byte, ...)
		// asigna en el heap de Go. El GC de Go se encargará.
		// Solo aseguramos que las referencias se rompan.
		current.base = 0
		current.size = 0
		current.offset = 0
		next := current.next
		current.next = nil // Rompe la cadena para evitar retención cíclica
		current = next
	}
	a.chunks = nil
	a.stats = arenaStats{} // Resetea las estadísticas
}

// Alloc asigna un bloque de memoria alineado dentro de la arena.
func (a *Arena) Alloc(size uintptr, align uintptr) unsafe.Pointer {
	if atomic.LoadUint32(&a.freed) == 1 {
		panic("arena: use after free")
	}

	if size == 0 {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	current := a.chunks
	if current == nil {
		// Esto no debería pasar si NewArena siempre crea el primer chunk
		// o si la arena no fue liberada prematuramente.
		// Crear uno nuevo como fallback.
		a.newChunk(size)
		current = a.chunks
	}

	// Alinea el offset
	alignedOffset := (current.offset + align - 1) &^ (align - 1)

	// Verifica si el chunk actual tiene suficiente espacio
	if alignedOffset+size > current.size {
		// No hay suficiente espacio, crea un nuevo chunk
		// El tamaño del nuevo chunk debe ser al menos el minChunkSize o lo suficientemente grande para la asignación actual
		newChunkSize := uintptr(math.Max(float64(minChunkSize), float64(size+align)))
		a.newChunk(newChunkSize)
		current = a.chunks // El nuevo chunk siempre se añade al frente
		alignedOffset = (current.offset + align - 1) &^ (align - 1)

		if alignedOffset+size > current.size {
			// Si todavía no hay suficiente espacio (ej. size es mayor que maxObjectSize),
			// esto indicaría un problema, pero para simplificar, se asume que newChunkSize lo maneja.
			return nil // Esto podría ser un pánico si prefieres un error irrecuperable
		}
	}

	ptr := current.base + alignedOffset
	current.offset = alignedOffset + size

	// Actualiza estadísticas
	a.stats.currentUsage += uint64(size)

	return unsafe.Pointer(ptr)
}

// CString copia un string de Go a memoria de C gestionada por la arena.
func (a *Arena) CString(s string) unsafe.Pointer {
	if len(s) == 0 {
		// Devuelve un puntero a una cadena vacía válida en C
		// o NULL, dependiendo del comportamiento deseado.
		// Para Arena, es mejor devolver un puntero a un byte nulo asignado.
		nullBytePtr := a.Alloc(1, 1) // Asigna 1 byte para el terminador nulo
		*(*C.char)(nullBytePtr) = 0  // Asegura que sea un null terminator
		return nullBytePtr
	}

	size := uintptr(len(s)) + 1 // +1 para el terminador nulo de C
	cptr := a.Alloc(size, 1)
	if cptr == nil {
		return nil
	}
	copy(unsafe.Slice((*byte)(cptr), size), s)
	*(*C.char)(unsafe.Pointer(uintptr(cptr) + uintptr(len(s)))) = 0 // Terminador nulo

	return cptr
}

// GoString convierte un puntero a char de C (gestionado por la arena) a un string de Go.
func (a *Arena) GoString(cptr *byte) string {
	if cptr == nil {
		return ""
	}
	// Asumimos que cptr apunta a una cadena terminada en nulo
	length := 0
	for p := uintptr(unsafe.Pointer(cptr)); ; p++ {
		if *(*C.char)(unsafe.Pointer(p)) == 0 {
			break
		}
		length++
	}
	return string(unsafe.Slice(cptr, length))
}

// GoBytes copia bytes de memoria de C (gestionada por la arena) a un slice de bytes de Go.
func (a *Arena) GoBytes(cptr *byte, length int) []byte {
	if cptr == nil || length <= 0 {
		return nil
	}

	goBytesPtr := a.Alloc(uintptr(length), 1) // 1-byte alignment is sufficient for bytes
	if goBytesPtr == nil {
		return nil
	}

	C.memcpy(goBytesPtr, unsafe.Pointer(cptr), C.size_t(length))

	return unsafe.Slice((*byte)(goBytesPtr), length)
}

// NewHandle crea un Handle para un array de Go, copiándolo a la arena.
func NewHandle[T any](arena *Arena, slice []T) *Handle[T] {
	if len(slice) == 0 {
		return nil
	}

	typ := reflect.TypeOf(slice).Elem()
	elemSize := typ.Size()
	totalSize := uintptr(len(slice)) * elemSize
	align := typ.Align()

	ptr := arena.Alloc(totalSize, uintptr(align))
	if ptr == nil {
		return nil
	}

	// Copiar los datos del slice de Go al bloque de memoria de la arena
	srcSliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	dstSliceHeader := &reflect.SliceHeader{
		Data: uintptr(ptr),
		Len:  len(slice),
		Cap:  len(slice),
	}
	C.memcpy(unsafe.Pointer(dstSliceHeader.Data), unsafe.Pointer(srcSliceHeader.Data), C.size_t(totalSize))

	return &Handle[T]{
		ptr:     ptr,
		arena:   arena, // Mantener referencia a la arena para asegurar que no se libere prematuramente
		elemTyp: typ,
		len:     len(slice),
	}
}

// Handle representa un puntero a un bloque de memoria de la arena que contiene un array de Go.
type Handle[T any] struct {
	ptr     unsafe.Pointer
	arena   *Arena // Mantener referencia a la arena
	elemTyp reflect.Type
	len     int
}

// Ptr devuelve el puntero C al inicio del array en la arena.
func (h *Handle[T]) Ptr() unsafe.Pointer {
	return h.ptr
}

// Get devuelve el array de Go desde la memoria de la arena.
func (h *Handle[T]) Get() []T {
	if h.ptr == nil {
		return nil
	}

	sliceHeader := &reflect.SliceHeader{
		Data: uintptr(h.ptr),
		Len:  h.len,
		Cap:  h.len,
	}
	return *(*[]T)(unsafe.Pointer(sliceHeader))
}

// CStringArray convierte un slice de strings de Go a un array de C strings en la arena.
func (a *Arena) CStringArray(goStrings []string) unsafe.Pointer {
	if len(goStrings) == 0 {
		return nil
	}

	// 1. Asignar memoria en la arena para el array de punteros char*
	// Cada puntero char* es de tamaño unsafe.Sizeof((*C.char)(nil))
	ptrArraySize := uintptr(len(goStrings)) * unsafe.Sizeof((*C.char)(nil))
	cptrArray := a.Alloc(ptrArraySize, ptrAlign)
	if cptrArray == nil {
		return nil
	}

	// 2. Iterar sobre los strings de Go, convertirlos a CString en la arena, y
	// almacenar sus punteros en el array de punteros
	for i, s := range goStrings {
		cstr := a.CString(s) // Convertir Go string a C string en la arena
		// Almacenar el puntero del C string en la posición correcta del array de punteros
		*(*unsafe.Pointer)(unsafe.Pointer(uintptr(cptrArray) + uintptr(i)*unsafe.Sizeof((*C.char)(nil)))) = cstr
	}

	return cptrArray
}

// GoStringArray convierte un array de C strings (char**) a un slice de strings de Go.
// Asume que los C strings individuales y el array de punteros están gestionados por la arena.
func (a *Arena) GoStringArray(cstrs **C.char, length int) []string {
	if cstrs == nil || length <= 0 {
		return nil
	}

	ptrs := (*[1 << 30]*C.char)(unsafe.Pointer(cstrs))[:length:length]
	result := make([]string, length)

	for i, cstr := range ptrs {
		if cstr != nil {
			result[i] = a.GoString((*byte)(unsafe.Pointer(cstr)))
		}
	}

	return result
}

// Estos ejemplos asumen que tienes cgo_utils.h con las funciones correspondientes

// ProcessStringWithArena procesa un string usando C y arena
func ProcessStringWithArena(input string, multiplier int) string {
	return WithArena(func(arena *Arena) string {
		cstr := arena.CString(input)

		// Llamar función C externa
		cresult := C.process_string((*C.char)(unsafe.Pointer(cstr)), C.int(multiplier))
		defer C.free(unsafe.Pointer(cresult))

		return arena.GoString((*byte)(unsafe.Pointer(cresult)))
	})
}

// ProcessArrayWithArena procesa un array usando C y arena
// Cambia a []int32
func ProcessArrayWithArena(numbers []int32, factor int) []int32 {
	return WithArena(func(arena *Arena) []int32 {
		handle := NewHandle(arena, numbers)

		// Llamar función C externa
		// Asegúrate de que C.int sea compatible con int32
		C.process_array((*C.int)(handle.Ptr()), C.size_t(len(numbers)), C.int(factor))

		return handle.Get()
	})
}

// GenerateDataWithArena genera datos en C y los copia a un slice de Go usando arena
func GenerateDataWithArena(size int) []byte {
	return WithArena(func(arena *Arena) []byte {
		cdata := C.generate_data(C.size_t(size))
		defer C.free(unsafe.Pointer(cdata)) // Liberar la memoria C original

		return arena.GoBytes((*byte)(unsafe.Pointer(cdata)), size)
	})
}

// PrintArenaStats imprime las estadísticas actuales de la arena
func (a *Arena) PrintArenaStats() {
	a.mu.Lock()
	defer a.mu.Unlock()
	fmt.Printf("Arena Stats: Total Chunks: %d, Total Allocated: %d bytes, Current Usage: %d bytes\n",
		a.stats.totalChunks, a.stats.totalAlloc, a.stats.currentUsage)
}

// WithArena es una función de conveniencia para usar la arena.
// Crea una arena, ejecuta la función f, y luego libera la arena.
func WithArena[T any](f func(arena *Arena) T) T {
	a := NewArena()
	defer a.Free()
	return f(a)
}

// WithArenaContext es similar a WithArena pero con un contexto de cancelación.
func WithArenaContext[T any](ctx context.Context, f func(ctx context.Context, arena *Arena) T) T {
	a := NewArena()
	defer a.Free()

	// Monitorear el contexto en un goroutine separado si es necesario para tareas prolongadas
	// o simplemente pasar el contexto a la función f.
	return f(ctx, a)
}
