# **Cross-Runtime Memory Management (Go + C + OCaml)**

The single most critical challenge in `argo` is managing memory across two entirely independent Garbage Collectors (GCs): **Go's GC** and **OCaml's GC**. Because the C FFI layer is blind to both, we enforce a strict separation of concerns to prevent segmentation faults and memory corruption.

```
┌──────────────┐                 ┌───────────┐                 ┌──────────────┐
│  Go Runtime  │ ──(Via CGO)──>  │  C Bridge │ ──(Via FFI)──>  │ OCaml (Why3) │
│  [Go's GC]   │                 │  Layer    │                 │  [OCaml GC]  │
└──────────────┘                 └───────────┘                 └──────────────┘

```

## 🧠 How the Runtimes Behave

### **1. OCaml Memory Management**

* **The Compaction Problem:** The OCaml GC aggressively moves objects within its *minor heap* during collection cycles to optimize memory.
* **The Fix (Tracking):** Any pointer exposed to C/OCaml must be registered as a local root using OCaml’s GC macros (`CAMLparam`, `CAMLlocal`, `CAMLreturn`). This ensures that if the GC moves a value, the reference is updated automatically.
* **External Memory:** If OCaml receives a raw pointer from Go, it treats it as an opaque **Custom Block**. The OCaml GC tracks the wrapper, but it will **never** move, scan, or free the underlying Go memory.

### **2. Go Memory Management & Arenas**

* **The Pointer Stability Problem:** Standard Go-allocated pointers are tricky to pass safely through CGO due to strict pointer-passing rules and potential GC tracking overhead.
* **The Fix (Arenas):** `argo` bypasses the standard Go heap for cross-runtime data by using manual **Arena Allocators**.
* **Stability:** Memory allocated inside an `argo` arena is pinned and stable; Go's GC will never move it, making it completely safe to pass down to C and OCaml.

## 🔄 Data Flow & Lifecycle Rules

1. **Go → OCaml:** Go provisions data inside an unmovable `argo` arena and passes the stable pointer via CGO. OCaml wraps it into a `value` custom block to use it safely during execution.
2. **OCaml → Go:** OCaml processes the data and returns results (like strings or integers). Go catches them via CGO and **immediately copies them** into its own domain (Go heap or arena) before the OCaml GC alters the context.
3. **Deallocation:** The OCaml GC never frees Go arena blocks. The Go side retains absolute ownership and determines when the arena is dropped.

> ⚠️ **Golden Rule:** OCaml execution using an arena pointer must entirely conclude *before* that arena is manually freed on the Go side.

## 📊 Architectural Trade-offs

| Approach | Pros | Cons |
| --- | --- | --- |
| **Go Arena Allocator** | Absolute pointer stability, bypasses standard CGO pointer rules. | Requires strict manual lifecycle tracking to avoid leaks. |
| **OCaml Custom Blocks** | Clean integration with OCaml's type system and GC tracking. | Requires boilerplate C stubs wrapped in OCaml `CAML` macros. |
| **IPC Alternative (JSON/Sockets)** | Complete runtime isolation; zero shared memory risks. | Heavy serialization overhead; significantly lower performance. |
