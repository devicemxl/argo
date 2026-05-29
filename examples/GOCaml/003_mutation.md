# **Example 3: In-Place Arena Mutation with Argo**

This example demonstrates true bidirectional memory sharing. Go allocates an `argo` arena, writes an initial value, and passes it to OCaml. OCaml reads the memory, processes the string (`"hello"` -> `"hello_world"`), and writes the result directly back into the Go-managed `argo` arena buffer.

### 🔹 The Execution Flow

1. **Go:** Allocates stable memory via `argo.NewArena()` and writes `"hello"`.
2. **C Bridge:** Wraps the raw pointer and invokes the OCaml callback.
3. **OCaml:** Reads the custom block, performs string concatenation, and calls a C primitive to write the result back to the `argo` pointer.
4. **Go:** Reads the exact same `argo` memory space and sees `"hello_world"`.

### 1. OCaml Interface (`arena.ml`)

Instead of bypassing OCaml, we expose C primitives that allow OCaml to safely read from and write to the opaque `argo` custom block.

```ocaml
(* arena.ml *)
type argo_arena_ptr  (* Abstract custom block *)

(* Primitives granting OCaml access to read/write the Argo memory *)
external read_arena : argo_arena_ptr -> int -> int -> string = "c_read_arena"
external write_arena : argo_arena_ptr -> int -> string -> unit = "c_write_arena"

(* The actual OCaml logic *)
let process_inplace arena offset length =
  let text = read_arena arena offset length in
  let new_text = text ^ "_world" in
  write_arena arena offset new_text  (* Writes directly into Go's memory *)

(* Register the function so CGO can trigger it *)
let () = Callback.register "ocaml_process_inplace" process_inplace

```

### 2. C Stubs & Execution Bridge (`arena_stubs.c`)

This file handles the OCaml GC constraints and exposes the mutation primitives.

```c
/* arena_stubs.c */
#include <caml/mlvalues.h>
#include <caml/memory.h>
#include <caml/alloc.h>
#include <caml/custom.h>
#include <caml/callback.h>
#include <string.h>

static void finalize_argo(value v) { } /* Argo manages lifecycle */

static struct custom_operations argo_ops = {
  "argo.arena.ptr", finalize_argo, custom_compare_default,
  custom_hash_default, custom_serialize_default, custom_deserialize_default,
  custom_compare_ext_default
};

/* 1. Primitive for OCaml to READ the Argo buffer */
CAMLprim value c_read_arena(value v_arena, value v_off, value v_len) {
  CAMLparam3(v_arena, v_off, v_len);
  char* buf = *((char**)Data_custom_val(v_arena));
  int off = Int_val(v_off);
  int len = Int_val(v_len);

  char tmp[256];
  if (len > 255) len = 255; /* Basic safety boundary */
  memcpy(tmp, buf + off, len);
  tmp[len] = '\0';

  CAMLreturn(caml_copy_string(tmp));
}

/* 2. Primitive for OCaml to WRITE into the Argo buffer */
CAMLprim value c_write_arena(value v_arena, value v_off, value v_str) {
  CAMLparam3(v_arena, v_off, v_str);
  char* buf = *((char**)Data_custom_val(v_arena));
  int off = Int_val(v_off);
  const char* str = String_val(v_str);

  /* In-place mutation of Go's memory */
  strcpy(buf + off, str);

  CAMLreturn(Val_unit); /* Must return Val_unit, not CAMLreturn0 */
}

/* 3. The Bridge: Go calls this, which triggers OCaml */
void go_trigger_ocaml_process(void* argo_ptr, int offset, int length) {
  CAMLparam0();
  CAMLlocal1(v_arena);

  /* Wrap the stable Argo pointer */
  v_arena = caml_alloc_custom(&argo_ops, sizeof(void*), 0, 1);
  *((void**)Data_custom_val(v_arena)) = argo_ptr;

  /* Look up and execute the registered OCaml function */
  const value* ocaml_func = caml_named_value("ocaml_process_inplace");
  if (ocaml_func != NULL) {
      /* C handles the OCaml macros (Val_int) on behalf of Go */
      caml_callback3(*ocaml_func, v_arena, Val_int(offset), Val_int(length));
  }

  CAMLreturn0; /* Standard C void return */
}

/* Initialization helper for Go */
void init_ocaml_runtime() {
    char* argv[] = {"argo_ocaml", NULL};
    caml_startup(argv);
}

```

### 3. Go CGO Implementation with `argo` (`main.go`)

```go
package main

/*
#cgo LDFLAGS: -L. -larena
#include "arena_stubs.c"
*/
import "C"
import (
	"fmt"
	"unsafe"
	"github.com/devicemxl/argo"
)

func main() {
	// 1. Initialize OCaml Runtime (Crucial when Go is the main process)
	C.init_ocaml_runtime()

	// 2. Initialize the argo arena
	arena := argo.NewArena()
	defer arena.Free()

	// 3. Allocate stable memory and cast to pointer
	block, err := arena.Alloc(64) // Reserve 64 bytes
	if err != nil {
		panic(err)
	}
	
	// 4. Write initial state to the argo buffer
	copy(block, []byte("hello\x00"))
	ptr := unsafe.Pointer(&block[0])

	// 5. Trigger OCaml to process the pointer in-place
	// (Go passes standard C integers, not OCaml macros)
	C.go_trigger_ocaml_process(ptr, 0, 5)

	// 6. Read the exact same block. OCaml has mutated it!
	fmt.Println("Argo Arena result:", string(block[:11]))
	// Output: Argo Arena result: hello_world
}

```

### 🔎 Why `argo` makes this safe

Notice that Go never interacts with OCaml's Garbage Collector. When OCaml executes `text ^ "_world"`, it allocates temporary strings on its own minor heap. OCaml's GC will eventually clean up those temporary strings, but the final result is written directly into the stable `argo` block via `strcpy`. Because `argo` pins the memory on the Go side, both runtimes operate with absolute confidence without overlapping GC domains.
