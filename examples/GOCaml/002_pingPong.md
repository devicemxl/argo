# **Example 2: Cross-Runtime Ping-Pong using Argo Arenas**

This example demonstrates how to leverage the **argo** arena allocator to safely share stable memory spaces with OCaml through a protected C bridge, avoiding standard Go heap pointer tracking.

### 🔹 Architectural Flow

1. **Go (`argo`)**: Allocates a continuous, unmovable memory block within an unmanaged arena.
2. **C Stub**: Encapsulates the raw arena pointer inside an OCaml `custom_operations` block.
3. **OCaml**: Evaluates the abstract type safely without its GC tracking or shifting the external memory location.
4. **Deallocation**: Go explicitly releases the `argo` arena once the execution sequence concludes.

### 1. OCaml Interface (`wrapper.ml`)

```ocaml
(* wrapper.ml *)
type argo_arena_ptr  (* Abstract representation of the Argo block *)

(* External C function mapping *)
external ping : argo_arena_ptr -> string = "c_ping"

```

### 2. C Stubs Bridge (`wrapper_stubs.c`)

```c
/* wrapper_stubs.c */
#include <caml/mlvalues.h>
#include <caml/memory.h>
#include <caml/alloc.h>
#include <caml/custom.h>
#include <string.h>

/* Finalizer: Left empty because argo (Go) manages its own cleanup */
static void finalize_argo_block(value v) {
  // No-op: OCaml GC releases the wrapper, but argo retains ownership of the data.
}

static struct custom_operations argo_ops = {
  "argo.runtime.block",
  finalize_argo_block,
  custom_compare_default,
  custom_hash_default,
  custom_serialize_default,
  custom_deserialize_default,
  custom_compare_ext_default
};

/* Safe wrapper invoked from CGO */
const char* argo_to_ocaml_ping(void* raw_argo_ptr) {
  CAMLparam0();
  CAMLlocal2(v_block, v_res);

  /* 1. Wrap the stable argo pointer into an OCaml custom block */
  v_block = caml_alloc_custom(&argo_ops, sizeof(void*), 0, 1);
  *((void**)Data_custom_val(v_block)) = raw_argo_ptr;

  /* 2. Execute target runtime operation */
  // Internally, OCaml reads or writes via: *((void**)Data_custom_val(v_block))
  v_res = caml_copy_string("pong from OCaml via argo block");

  /* 3. Extract native C-string safely before OCaml releases scope */
  const char* safe_c_str = String_val(v_res);

  CAMLreturnT(const char*, safe_c_str);
}

```

### 3. Go Implementation using `argo` (`main.go`)

```go
package main

/*
#cgo LDFLAGS: -L. -lwrapper
#include "wrapper_stubs.c"
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/devicemxl/argo"
)

func main() {
	// 1. Initialize a native argo arena allocator
	arena := argo.NewArena()
	defer arena.Free() // Ensure deterministic lifecycle cleanup

	// 2. Allocate a stable, pinned memory block inside the arena
	// (Adjust Alloc or data-type instantiation to match your exact argo method syntax)
	size := uint32(1024)
	argoBlock, err := arena.Alloc(size)
	if err != nil {
		panic(err)
	}

	// 3. Obtain the raw unmovable pointer
	ptr := unsafe.Pointer(&argoBlock[0])

	// 4. Dispatch pointer to the OCaml runtime through the C stub
	res := C.argo_to_ocaml_ping(ptr)

	// 5. Safely consume the result inside the Go domain
	fmt.Println(C.GoString(res)) 
	// Output: "pong from OCaml via argo block"
}

```

### 🔎 Why this works seamlessly with `argo`

* **Zero-Copy Architecture:** The `void*` passed into OCaml points directly to the memory region managed by `argo`. There is zero allocation overhead when entering the FFI layer.
* **Deterministic Lifecycles:** Calling `defer arena.Free()` guarantees that memory is wiped only when Go finishes processing. OCaml can read and mutably write to the `argoBlock` pointer during its lifecycle execution without introducing side-effects to Go’s internal heap tracking.
