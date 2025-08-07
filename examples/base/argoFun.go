package main

/*
#include "cgo_utils.h"
*/
import "C"
import (
	"unsafe"

	"github.com/devicemxl/argo"
)

// === Funciones de procesamiento de strings ===

func ProcessStringWithArena(arena *argo.Arena, input string, multiplier int) string {
	cstr := arena.UnsafeCString(input)
	cresult := C.process_string((*C.char)(cstr), C.int(multiplier))
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}

func ConcatStringsWithArena(arena *argo.Arena, str1, str2 string) string {
	cstr1 := arena.UnsafeCString(str1) // Usar UnsafeCString en lugar de CString
	cstr2 := arena.UnsafeCString(str2) // Usar UnsafeCString en lugar de CString
	cresult := C.concat_strings((*C.char)(cstr1), (*C.char)(cstr2))
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}

func ToUppercaseWithArena(arena *argo.Arena, input string) string {
	cstr := arena.CString(input)
	cresult := C.to_uppercase((*C.char)(cstr))
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}

func ToLowercaseWithArena(arena *argo.Arena, input string) string {
	cstr := arena.CString(input)
	cresult := C.to_lowercase((*C.char)(cstr))
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}

// === Funciones de procesamiento de arrays ===

// Nota: Esta función modifica el array in-place, no retorna nuevo array
func ProcessArrayInPlace(arena *argo.Arena, numbers []int32, factor int) {
	if len(numbers) == 0 {
		return
	}
	handle := argo.NewHandle(arena, numbers)
	C.process_array((*C.int)(handle.Ptr()), C.size_t(len(numbers)), C.int(factor))
	// Los cambios están directamente en el handle, usar handle.Get() para obtener el slice modificado
}

func SumArrayWithArena(arena *argo.Arena, numbers []int32) int {
	if len(numbers) == 0 {
		return 0
	}
	handle := argo.NewHandle(arena, numbers)
	result := C.sum_array((*C.int)(handle.Ptr()), C.size_t(len(numbers)))
	return int(result)
}

func MaxArrayWithArena(arena *argo.Arena, numbers []int32) int {
	if len(numbers) == 0 {
		return 0
	}
	handle := argo.NewHandle(arena, numbers)
	result := C.max_array((*C.int)(handle.Ptr()), C.size_t(len(numbers)))
	return int(result)
}

func MinArrayWithArena(arena *argo.Arena, numbers []int32) int {
	if len(numbers) == 0 {
		return 0
	}
	handle := argo.NewHandle(arena, numbers)
	result := C.min_array((*C.int)(handle.Ptr()), C.size_t(len(numbers)))
	return int(result)
}

func SortArrayInPlace(arena *argo.Arena, numbers []int32) {
	if len(numbers) <= 1 {
		return
	}
	handle := argo.NewHandle(arena, numbers)
	C.sort_array((*C.int)(handle.Ptr()), C.size_t(len(numbers)))
}

// === Funciones de generación de datos ===

func GenerateDataWithArena(arena *argo.Arena, size int) string {
	cresult := C.generate_data(C.size_t(size))
	if cresult == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}

func GenerateSequenceWithArena(arena *argo.Arena, start, count int) []int32 {
	cresult := C.generate_sequence(C.int(start), C.int(count))
	if cresult == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(cresult))

	// Copiar resultado a arena
	handle := argo.NewHandle(arena, make([]int32, count))
	C.safe_memcpy(unsafe.Pointer(handle.Ptr()), unsafe.Pointer(cresult), C.size_t(count*4))
	return handle.Get()
}

func GenerateRandomArrayWithArena(arena *argo.Arena, count, min, max int) []int32 {
	cresult := C.generate_random_array(C.size_t(count), C.int(min), C.int(max))
	if cresult == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(cresult))

	// Copiar resultado a arena
	handle := argo.NewHandle(arena, make([]int32, count))
	C.safe_memcpy(unsafe.Pointer(handle.Ptr()), unsafe.Pointer(cresult), C.size_t(count*4))
	return handle.Get()
}

// === Funciones matemáticas ===

func FastPow(base float64, exp int) float64 {
	result := C.fast_pow(C.double(base), C.int(exp))
	return float64(result)
}

func SqrtNewton(x float64) float64 {
	result := C.sqrt_newton(C.double(x))
	return float64(result)
}

func Factorial(n int) uint64 {
	result := C.factorial(C.int(n))
	return uint64(result)
}

// === Funciones de criptografía ===

func CaesarCipherWithArena(arena *argo.Arena, text string, shift int) string {
	cstr := arena.CString(text)
	cresult := C.caesar_cipher(((*C.char)(cstr)), C.int(shift))
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}

func CaesarDecipherWithArena(arena *argo.Arena, text string, shift int) string {
	cstr := arena.CString(text)
	cresult := C.caesar_decipher(((*C.char)(cstr)), C.int(shift))
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}

// === Funciones de validación ===

func IsValidEmail(email string) bool {
	cemail := C.CString(email)
	defer C.free(unsafe.Pointer(cemail))
	result := C.is_valid_email(cemail)
	return bool(result)
}

func IsValidURL(url string) bool {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	result := C.is_valid_url(curl)
	return bool(result)
}

// === Funciones de conversión ===

func BytesToHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	cresult := C.bytes_to_hex((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
	defer C.free(unsafe.Pointer(cresult))
	return C.GoString(cresult)
}

func HexToBytes(hex string) []byte {
	chex := C.CString(hex)
	defer C.free(unsafe.Pointer(chex))

	var size C.size_t
	cresult := C.hex_to_bytes(chex, &size)
	if cresult == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(cresult))

	// Copiar a slice de Go
	result := make([]byte, int(size))
	C.safe_memcpy(unsafe.Pointer(&result[0]), unsafe.Pointer(cresult), size)
	return result
}

// === Funciones de utilidad ===

func SimpleHash(str string) uint32 {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	result := C.simple_hash(cstr)
	return uint32(result)
}

func IsValidNumber(str string) bool {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	result := C.is_valid_number(cstr)
	return bool(result)
}

func SafeStrToInt(str string) (int, bool) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	var result C.int
	success := C.safe_str_to_int(cstr, &result)
	return int(result), bool(success)
}

// === Funciones de bits ===

func CountBits(n uint32) int {
	result := C.count_bits(C.uint(n))
	return int(result)
}

func ReverseBits(n uint32) uint32 {
	result := C.reverse_bits(C.uint(n))
	return uint32(result)
}
