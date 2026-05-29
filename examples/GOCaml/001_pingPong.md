# **Example 1: Safe Ping-Pong with Go Arenas**

This example demonstrates the minimum working implementation of passing a Go Arena pointer into the OCaml runtime safely. We achieve this by wrapping the external Go pointer inside an OCaml **Custom Block**.

### 🔹 The Strategy

1. **Go** allocates an arena and passes the raw memory pointer to C.
2. **C** safely wraps this pointer in an OCaml Custom Block within a protected GC scope.
3. **OCaml** processes the block (opaque to its GC) and returns an OCaml string.
4. **C** extracts a standard C-string from the OCaml value and hands it back to Go.

---

### 1. OCaml Interface (`wrapper.ml`)

On the OCaml side, we define an abstract type. OCaml doesn't need to know it's a Go pointer; it just treats it as an opaque block of data.

```ocaml
(* wrapper.ml *)
type arena_ptr  (* Abstract custom block type *)

(* External C function declaration *)
external ping : arena_ptr -> string = "c_ping"

```

---

### 2. C Stubs & GC Wrappers (`wrapper_stubs.c`)

This is the bridge. We must strictly use `CAML` macros to ensure the OCaml GC tracks our variables while we handle the custom block allocation and string conversion.

```c
/* wrapper_stubs.c */
#include <caml/mlvalues.h>
#include <caml/memory.h>
#include <caml/alloc.h>
#include <caml/custom.h>
#include <string.h>

/* Optional Finalizer: Does nothing, Go retains ownership of the arena */
static void finalize_arena(value v) {
  // No-op: Go is responsible for freeing the arena memory.
}

/* Custom operations definition for the block */
static struct custom_operations arena_ops = {
  "argo.go.arena",
  finalize_arena,
  custom_compare_default,
  custom_hash_default,
  custom_serialize_default,
  custom_deserialize_default,
  custom_compare_ext_default
};

/* The actual OCaml primitive */
CAMLprim value c_ping(value v_arena) {
  CAMLparam1(v_arena);
  
  /* Extract the raw Go pointer from the custom block */
  void* ptr = *((void**)Data_custom_val(v_arena));
  
  /* ... Read or write to the Go arena using 'ptr' here ... */

  /* Return an OCaml string value */
  const char* msg = "pong from OCaml";
  CAMLreturn(caml_copy_string(msg));
}

/* * Safe Wrapper for Go 
 * Go should never hold raw OCaml 'value' types. This wrapper 
 * handles the OCaml GC scope safely before returning a standard char*.
 */
const char* go_to_ocaml_ping(void* go_arena_ptr) {
  CAMLparam0();
  CAMLlocal2(v_arena, v_res);

  /* 1. Wrap the raw Go pointer in an OCaml custom block */
  v_arena = caml_alloc_custom(&arena_ops, sizeof(void*), 0, 1);
  *((void**)Data_custom_val(v_arena)) = go_arena_ptr;

  /* 2. Execute the ping function */
  v_res = c_ping(v_arena);

  /* 3. Extract standard C string from OCaml string value to prevent segfaults in Go */
  const char* safe_c_str = String_val(v_res);

  CAMLreturnT(const char*, safe_c_str);
}

```

---

### 3. Go CGO Implementation (`main.go`)

Go handles the actual memory allocation. It passes the raw pointer down to the C wrapper, keeping standard Go GC mechanics isolated from OCaml.

```go
/*
#cgo LDFLAGS: -L. -lwrapper
#include "wrapper_stubs.c"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	// 1. Allocate stable arena in Go
	arena := make([]byte, 1024)
	ptr := unsafe.Pointer(&arena[0])

	// 2. Pass raw pointer to C wrapper. 
	// C will handle the OCaml custom block and GC scopes internally.
	res := C.go_to_ocaml_ping(ptr)

	// 3. Convert the returned safe C string to a Go string
	fmt.Println(C.GoString(res)) // Output: pong from OCaml
	
	// Arena lifecycle remains under Go's control here.
}

```

### 🔎 Key Takeaways

* **No Leaking `value`:** Go passes `void*` and receives `char*`. The OCaml `value` types (`v_arena`, `v_res`) only exist within the safely bounded `CAMLlocal` scope inside C.
* **Header Safety:** Using `String_val()` is mandatory. Passing `caml_copy_string` directly to `C.GoString` reads the OCaml memory header as text, causing undefined behavior.
* **Opaque Tracking:** The OCaml GC registers the custom block wrapper, but ignores the internal Go pointer, preserving the arena's integrity.
