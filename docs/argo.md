## 📄 **argo.go**

The file `argo.go` provides a zero-GC (Garbage Collector-free) memory arena system implemented in Go using raw memory allocation via C (with `malloc` and `free`). It enables efficient, manual memory control ideal for interop with C, temporary data lifetimes, or avoiding GC pauses in high-performance scenarios.

---

```go
// 📄 argo.go
//
// A zero-GC memory arena written in Go using C allocations.
// Provides manual memory management, C interop, and utilities for strings and arrays.
// Inspired by techniques in systems programming for predictable memory behavior.

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
```

---

## 🧱 **1. General Purpose**

This file implements a manual memory allocator ("Arena") that avoids interaction with the Go garbage collector. It is especially useful when building bindings with C, or when deterministic memory lifetime and performance are required.

---

## 🧩 **2. Key Components**

* `Arena`: main structure for memory allocation
* `chunk`: internal memory block used by the arena
* `Handle[T]`: typed container that wraps arena memory
* `NewArena`, `Free`, `Alloc`: core memory lifecycle
* `CString`, `GoString`, `CStringArray`: string interop helpers
* `WithArena`, `WithArenaContext`: scoped arena usage
* `PrintArenaStats`: debug tool for memory reporting

---

## 📐 **3. Internal Architecture**

The arena holds a linked list of `chunk`s, each representing a block of memory allocated via `C.malloc`. It keeps track of current usage, supports alignment, and prevents memory leaks via `Free()`.

---

## 🔧 **4. Core Structures**

### **`chunk`**

```go
type chunk struct {
	base   uintptr // base memory pointer (C malloc'ed)
	size   uintptr // total allocated size
	offset uintptr // current allocation offset
	next   *chunk  // linked list to the next chunk
}
```

#### 📌 Purpose

Encapsulates a memory block allocated outside Go's GC using `malloc`.

#### 🎯 Uses

* Holds all memory slices allocated through the arena.
* Linked list of memory pages for scalable growth.

---

### **`Arena`**

```go
type Arena struct {
	mu     sync.Mutex // protects concurrent access
	chunks *chunk     // head of the chunk list
	freed  uint32     // indicates if memory was released
	stats  arenaStats // memory tracking
}
```

#### 📌 Purpose

Manual memory manager that operates outside Go's GC.

#### 🧠 Use Cases

* Temporary buffers for interop (e.g., C APIs)
* Performance-critical code where GC must be avoided

#### 🧪 Key Methods

* `NewArena()`
* `Alloc(size, align)`
* `Free()`
* `CString(s string)`
* `PrintArenaStats()`

---

## 🎛️ **5. Enums and Flags**

### **`arenaStats`**

```go
type arenaStats struct {
	totalAlloc   uint64 // total memory ever allocated
	totalChunks  uint32 // number of chunks created
	currentUsage uint64 // current active usage
}
```

#### 🧩 Purpose

Stores memory usage metrics for debugging and profiling.

---

## ⚙️ **6. Traits / Interfaces**

N/A — uses concrete types instead of interfaces for performance and simplicity.

---

## 🏗️ **7. Factory Pattern**

### **`NewArena`**

```go
func NewArena(opts ...ArenaOption) *Arena
```

#### 🧠 Purpose

Creates and initializes a new memory arena.

#### 🔍 Heuristics

* Uses minimum chunk size (`8192 bytes`)
* Sets a finalizer to ensure memory gets freed on GC if forgotten

---

## 🧯 **8. Error Handling**

Memory allocation is explicitly panicked on failure:

```go
if cptr == nil {
    panic("arena: failed to allocate memory")
}
```

Additionally, safe-guards exist to prevent double-free or use-after-free via the `freed` atomic flag.

---

## 🤝 **9. Dependencies**

* `C.malloc`, `C.free`, `C.memset`, `C.memcpy`, `C.strlen`: for low-level memory operations
* `unsafe`: for working with raw pointers
* `runtime.SetFinalizer`: integrates arena with Go's GC lifecycle

---

## 🧪 **10. Usage Patterns**

### 🟢 Basic Usage

```go
a := argo.NewArena()
defer a.Free()
ptr := a.Alloc(64, 8)
```

### ⚙️ Advanced: C String Interop

```go
cstr := a.CString("hello")
fmt.Println(a.GoString(cstr))
```

### 🔄 Handle API

```go
handle := argo.NewHandle(a, []int{1, 2, 3})
fmt.Println(handle.Get()) // returns []int from arena
```

---

## 🚀 **11. Future Extensions**

* Chunk reuse and compaction
* Advanced alignment (SIMD support)
* Scoped allocations (sub-arenas)
* Zero-copy interfaces with I/O

---

## 📈 **12. Performance Considerations**

* All allocations are page-aligned and GC-free
* Memory is only freed when `Arena.Free()` is called
* Eliminates GC pauses in allocation-heavy paths
* Zero-copy string support improves FFI performance

---

## ✅ **13. Conclusion**

This file implements a minimal yet powerful arena allocator in Go using raw C memory. It's an essential piece for efficient interop, temporary memory buffers, or deterministic memory handling in low-level systems. Keeping it modular and leak-free is crucial for long-term reliability.
