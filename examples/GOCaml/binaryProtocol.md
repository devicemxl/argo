# **Example 4: Custom Binary Protocol via Argo Arenas**

This advanced example demonstrates a full-cycle, zero-copy protocol. Go provisions an `argo` arena and formats a binary payload. OCaml parses the payload dynamically, computes a mathematical sum, writes the result back into the reserved arena slot *in-place*, and returns a status string.

### 🔹 Arena Memory Layout

The protocol relies on a contiguous byte format to prevent cross-runtime ambiguity:

* `offset 0`: `cmd_len` (uint32 little-endian)
* `offset 4`: `cmd` (bytes, variable length, no NUL terminator)
* `offset 4 + cmd_len`: `count` (uint32 little-endian)
* `offset 8 + cmd_len`: `ints` (`count` × int32 little-endian)
* `offset End`: `result` (int32) — Reserved for OCaml's computation output.

### 1. OCaml Interface & Logic (`protocol.ml`)

```ocaml
(* protocol.ml *)
type argo_arena_ptr

external read_cmd_len  : argo_arena_ptr -> int -> int = "c_read_u32"
external read_cmd      : argo_arena_ptr -> int -> int -> string = "c_read_bytes"
external read_count    : argo_arena_ptr -> int -> int = "c_read_u32"
external read_int32    : argo_arena_ptr -> int -> int = "c_read_i32"
external write_int32   : argo_arena_ptr -> int -> int -> unit = "c_write_i32"

let process_protocol arena =
  (* 1. Parse Headers *)
  let cmd_len = read_cmd_len arena 0 in
  let cmd = read_cmd arena 4 cmd_len in
  
  let count_off = 4 + cmd_len in
  let count = read_count arena count_off in
  
  (* 2. Aggregate Data *)
  let ints_off = count_off + 4 in
  let rec loop i acc =
    if i = count then acc
    else
      let v = read_int32 arena (ints_off + i * 4) in
      loop (i + 1) (acc + v)
  in
  let total = loop 0 0 in
  
  (* 3. Write output in-place *)
  let result_off = ints_off + count * 4 in
  write_int32 arena result_off total;
  
  (* 4. Return status *)
  Printf.sprintf "Status: %s_processed | Result: %d" cmd total

(* Register callback for the C bridge *)
let () = Callback.register "ocaml_process_protocol" process_protocol

```

### 2. C Stubs Bridge (`protocol_stubs.c`)

```c
/* protocol_stubs.c */
#include <caml/mlvalues.h>
#include <caml/memory.h>
#include <caml/alloc.h>
#include <caml/custom.h>
#include <caml/callback.h>
#include <stdint.h>
#include <string.h>

static void finalize_argo(value v) { } /* Argo manages lifecycle */

static struct custom_operations argo_ops = {
  "argo.protocol.arena", finalize_argo, custom_compare_default,
  custom_hash_default, custom_serialize_default, custom_deserialize_default,
  custom_compare_ext_default
};

CAMLprim value c_read_u32(value v, value offset) {
  CAMLparam2(v, offset);
  uint8_t* buf = *((uint8_t**)Data_custom_val(v));
  uint32_t x;
  memcpy(&x, buf + Int_val(offset), 4);
  CAMLreturn(Val_int((int)x));
}

CAMLprim value c_read_bytes(value v, value offset, value len) {
  CAMLparam3(v, offset, len);
  char* buf = *((char**)Data_custom_val(v));
  CAMLreturn(caml_copy_string_len(buf + Int_val(offset), Int_val(len)));
}

CAMLprim value c_read_i32(value v, value offset) {
  CAMLparam2(v, offset);
  int8_t* buf = (int8_t*)(*((void**)Data_custom_val(v)));
  int32_t x;
  memcpy(&x, buf + Int_val(offset), 4);
  CAMLreturn(Val_int((int)x));
}

CAMLprim value c_write_i32(value v, value offset, value num) {
  CAMLparam3(v, offset, num);
  int8_t* buf = (int8_t*)(*((void**)Data_custom_val(v)));
  int32_t x = (int32_t)Int_val(num);
  memcpy(buf + Int_val(offset), &x, 4);
  CAMLreturn(Val_unit); /* Must return Val_unit, not CAMLreturn0 */
}

/* Execution Bridge triggered by Go */
const char* go_trigger_protocol(void* argo_ptr) {
  CAMLparam0();
  CAMLlocal2(v_arena, v_res);

  v_arena = caml_alloc_custom(&argo_ops, sizeof(void*), 0, 1);
  *((void**)Data_custom_val(v_arena)) = argo_ptr;

  const value* ocaml_func = caml_named_value("ocaml_process_protocol");
  if (ocaml_func != NULL) {
      v_res = caml_callback(*ocaml_func, v_arena);
      CAMLreturnT(const char*, String_val(v_res));
  }
  
  CAMLreturnT(const char*, "Error: OCaml callback not found");
}

void init_ocaml_runtime() {
    char* argv[] = {"argo_ocaml", NULL};
    caml_startup(argv);
}

```

### 3. Go CGO Implementation (`main.go`)

```go
package main

/*
#cgo LDFLAGS: -L. -lprotocol
#include "protocol_stubs.c"
*/
import "C"
import (
	"fmt"
	"unsafe"
	"github.com/devicemxl/argo"
)

func u32ToBytesLE(x uint32) []byte {
	return []byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24)}
}

func i32ToBytesLE(x int32) []byte {
	return []byte{byte(x), byte(x >> 8), byte(x >> 16), byte(x >> 24)}
}

func main() {
	C.init_ocaml_runtime()

	// 1. Initialize Argo Arena
	arena := argo.NewArena()
	defer arena.Free()

	cmd := []byte("sum")
	cmdLen := uint32(len(cmd))
	ints := []int32{1, 2, 3}
	count := uint32(len(ints))

	// 2. Calculate dynamic layout size
	totalSize := 4 + int(cmdLen) + 4 + (4 * len(ints)) + 4
	block, err := arena.Alloc(uint32(totalSize))
	if err != nil {
		panic(err)
	}

	// 3. Serialize payload into Argo block
	copy(block[0:4], u32ToBytesLE(cmdLen))
	copy(block[4:4+len(cmd)], cmd)
	
	offCount := 4 + len(cmd)
	copy(block[offCount:offCount+4], u32ToBytesLE(count))
	
	offInts := offCount + 4
	for i, v := range ints {
		copy(block[offInts+i*4 : offInts+i*4+4], i32ToBytesLE(v))
	}
	
	resultOff := offInts + 4*len(ints)
	copy(block[resultOff:resultOff+4], i32ToBytesLE(0))

	// 4. Dispatch pointer to OCaml runtime
	ptr := unsafe.Pointer(&block[0])
	resStr := C.go_trigger_protocol(ptr)

	// 5. Evaluate Results
	fmt.Println("OCaml Reply:", C.GoString(resStr))

	// Decode in-place mutation
	b := block[resultOff : resultOff+4]
	total := int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
	fmt.Println("Result in Argo Arena:", total) // Output: 6
}

```

### 🛡️ Security & Architecture Notes

* **GC Roots & Scope:** Complex stubs must register `CAMLparam` and `CAMLlocal` correctly to prevent the OCaml GC from sweeping values mid-execution.
* **Alignment & Endianness:** By explicitly enforcing Little-Endian conversion (`u32ToBytesLE`) and fixed-width types (`uint32_t`), the protocol ensures total hardware and compiler independence.
* **Argo Stability Guarantee:** Because `argo` prevents Go's GC from moving the memory slice, `void* ptr` remains absolutely stable while the C/OCaml thread operates on it.
* **Out-of-Bounds (OOB) Safety:** In production, add offset validation checks in the C stubs (`if (offset + length > MAX_ARENA_SIZE) return Error;`) before executing `memcpy` to prevent buffer overflow exploits.
* **Concurrency:** If multiple Go routines share the same `argo` block, thread access must be synchronized manually (via mutexes), as the C/OCaml runtimes are oblivious to Go's concurrency model.
