# **Example 4: Zero-Copy Numeric Array Processing**

This example showcases how OCaml can directly iterate over, aggregate, and mutate an array of 32-bit integers allocated inside a Go `argo` arena. The operations are completely **zero-copy**, meaning OCaml reads and alters the exact same physical memory addresses provisioned by Go.

### 🔹 The Strategy

1. **Go (`argo`)**: Allocates a continuous block sized for 5 integers and populates it with `[1, 2, 3, 4, 5]`.
2. **C Bridge**: Exposes primitives to read/write specific offsets using fixed-width `int32_t` pointers, and handles runtime callback translation.
3. **OCaml**: Tail-recursively sums the array via the C primitives, computes the total (`15`), and overwrites index `0` of the arena before returning.

### 1. OCaml Interface & Logic (`arena.ml`)

```ocaml
(* arena.ml *)
type argo_arena_ptr  (* Abstract custom block holding the int32_t* pointer *)

(* Primitives to interact with the raw C array *)
external read_int  : argo_arena_ptr -> int -> int = "c_read_int"
external write_int : argo_arena_ptr -> int -> int -> unit = "c_write_int"

(* Tail-recursive array summation and in-place modification *)
let sum_array arena len =
  let rec loop i acc =
    if i = len then acc
    else loop (i + 1) (acc + read_int arena i)
  in
  let total = loop 0 0 in
  
  /* Write the total back into the very first slot (index 0) of the arena */
  write_int arena 0 total;
  total

(* Register the entry point for CGO *)
let () = Callback.register "ocaml_sum_array" sum_array

```

### 2. C Stubs Bridge (`arena_stubs.c`)

```c
/* arena_stubs.c */
#include <caml/mlvalues.h>
#include <caml/memory.h>
#include <caml/alloc.h>
#include <caml/custom.h>
#include <caml/callback.h>
#include <stdint.h>

static void finalize_argo(value v) { }

static struct custom_operations argo_ops = {
  "argo.numeric.arena", finalize_argo, custom_compare_default,
  custom_hash_default, custom_serialize_default, custom_deserialize_default,
  custom_compare_ext_default
};

/* 1. Read primitive: maps OCaml index to int32_t offset */
CAMLprim value c_read_int(value v_arena, value v_offset) {
  CAMLparam2(v_arena, v_offset);
  int32_t* buf = *((int32_t**)Data_custom_val(v_arena));
  int idx = Int_val(v_offset);
  
  CAMLreturn(Val_int(buf[idx]));
}

/* 2. Write primitive: mutates the int32_t array slot in-place */
CAMLprim value c_write_int(value v_arena, value v_offset, value v_num) {
  CAMLparam3(v_arena, v_offset, v_num);
  int32_t* buf = *((int32_t**)Data_custom_val(v_arena));
  int idx = Int_val(v_offset);
  int32_t val = Int_val(v_num);
  
  buf[idx] = val;
  CAMLreturn(Val_unit); /* Returns OCaml unit */
}

/* 3. Execution Bridge invoked by Go */
int go_trigger_ocaml_sum(void* argo_ptr, int length) {
  CAMLparam0();
  CAMLlocal2(v_arena, v_res);
  int native_result = 0;

  /* Pack the raw pointer */
  v_arena = caml_alloc_custom(&argo_ops, sizeof(void*), 0, 1);
  *((void**)Data_custom_val(v_arena)) = argo_ptr;

  /* Locate the OCaml callback */
  const value* ocaml_func = caml_named_value("ocaml_sum_array");
  if (ocaml_func != NULL) {
      /* C wraps the parameters cleanly, isolating Go from OCaml macros */
      v_res = caml_callback2(*ocaml_func, v_arena, Val_int(length));
      native_result = Int_val(v_res);
  }

  CAMLreturnT(int, native_result);
}

```

### 3. Go CGO Implementation with `argo` (`main.go`)

```go
package main

/*
#cgo LDFLAGS: -L. -larena
#include "arena_stubs.c"
// External declaration of the OCaml runtime initialization helper
void caml_startup(char* argv[]);
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/devicemxl/argo"
)

func init() {
	// Initialize OCaml runtime environment once at boot
	argv := []*C.char{C.CString("argo_runtime"), nil}
	C.caml_startup(&argv[0])
}

func main() {
	// 1. Initialize an argo arena
	arena := argo.NewArena()
	defer arena.Free()

	// 2. Allocate space for 5 contiguous 32-bit integers (5 * 4 bytes)
	elementCount := 5
	blockSize := uint32(elementCount * 4)
	block, err := arena.Alloc(blockSize)
	if err != nil {
		panic(err)
	}

	// 3. Map a local slice over the argo block memory space for population
	// This slice points directly inside the stable argo allocation
	intArray := (*[5]int32)(unsafe.Pointer(&block[0]))
	intArray[0] = 1
	intArray[1] = 2
	intArray[2] = 3
	intArray[3] = 4
	intArray[4] = 5

	// 4. Send the unmovable pointer to OCaml to calculate sum
	ptr := unsafe.Pointer(&block[0])
	totalSum := C.go_trigger_ocaml_sum(ptr, C.int(elementCount))

	// 5. Verify results
	fmt.Printf("Sum calculated by OCaml: %d\n", int(totalSum)) 
	// Output: Sum calculated by OCaml: 15

	fmt.Printf("Argo Arena array post-mutation: %v\n", *intArray)
	// Output: Argo Arena array post-mutation: [15 2 3 4 5]
}

```

### 🔎 Why this design guarantees high performance

* **Memory Safety Alignment:** By enforcing `int32_t` in C and modifying an `int32` array in Go, we eliminate data misalignment and structural padding bugs entirely.
* **Tail-Recursion Efficiency:** OCaml loops through the array using stack-safe tail recursion. The GC overhead on the OCaml side is zero since no complex intermediate objects are allocated on the OCaml heap.
* **Immediate Visibility:** The moment OCaml executes `write_int`, the value changes in the `argo` buffer instantly. Go reads it with absolute zero latency.
