# 📄 **cgo\_utils.h**

The `cgo_utils.h` header file serves as the public interface for the C utility functions implemented in `cgo_utils.c`. It declares all the functions, constants, and basic types that are intended to be accessed from Go code via CGO, providing necessary prototypes and ensuring proper compilation and linking.

-----

## 🧱 **1. General Purpose**

This file exists to formally define the API of the C helper functions. It acts as the contract between the Go and C parts of the application, ensuring type compatibility and correct function signatures when Go calls C functions. Without this header, CGO would not know how to correctly link and pass data to the C implementations.

-----

## 🧩 **2. Key Components**

  * **Header Guards**: `#ifndef CGO_UTILS_H`, `#define CGO_UTILS_H`, `#endif`.
  * **Standard Includes**: `stdio.h`, `stdlib.h`, `string.h`, `stdint.h`, `stdbool.h`.
  * **C++ Compatibility**: `#ifdef __cplusplus extern "C" { #endif` block.
  * **Function Declarations**:
      * String Processing: `process_string`, `concat_strings`, `to_uppercase`, `to_lowercase`, `free_string_array`.
      * Array Processing: `process_array`, `sum_array`, `max_array`, `min_array`, `sort_array`.
      * Data Generation: `generate_data`, `generate_sequence`, `generate_random_array`.
      * Mathematical Calculations: `fast_pow`, `sqrt_newton`, `factorial`.
      * Bit Processing: `count_bits`, `reverse_bits`.
      * Simple Cryptography: `caesar_cipher`, `caesar_decipher`.
      * Validation: `is_valid_email`, `is_valid_url`.
      * Conversion: `bytes_to_hex`, `hex_to_bytes`.
      * Benchmarking: `measure_time`.
      * Logging: `log_message`, `log_formatted_message`.

-----

## 📐 **3. Architecture / Internal Design**

`cgo_utils.h` is a flat header file, meaning it contains a direct list of function prototypes grouped by category with comments for clarity. It does not define complex data structures or internal architectural details; its sole purpose is to expose the C API.

### 🧭 **Responsibilities**

  * **API Definition**: Clearly defines the names, parameters, and return types of all C functions available for external (Go) consumption.
  * **Type Safety**: Ensures that Go, when calling C functions, uses the correct data types, preventing CGO compilation errors related to type mismatches.
  * **Modularity**: Allows the `cgo_utils.c` implementation details to remain separate, providing a clean interface.
  * **Compilation Orchestration**: Instructs the C compiler and CGO about the available functions and their signatures.

-----

## 🔧 **4. Principal Structures or Traits**

This file primarily declares functions. No custom `struct`s or `trait`s are defined here; it relies on standard C types (`char*`, `int`, `size_t`, `double`, `uint32_t`, `uint64_t`, `bool`).

-----

## 🎛️ **5. Enumerations / Important Flags**

This file includes `stdbool.h` which defines `bool`, `true`, and `false`. No custom enumerations are defined within `cgo_utils.h`.

-----

## ⚙️ **6. Traits / Interfaces**

As a C header, it defines function prototypes which serve as the "interface" for the C code.

### **Function Prototypes (e.g., `char* process_string(const char* input, int multiplier);`)**

```c
// === String Processing Functions ===
char* process_string(const char* input, int multiplier);
// ...
```

#### 📌 Purpose

These prototypes define the contract for how Go functions can interact with the underlying C implementations. They specify the exact arguments required and the type of value that will be returned.

#### 🔍 Main Methods

  * `char* process_string(const char* input, int multiplier);`: Repeats an input string.
  * `void process_array(int* arr, size_t len, int factor);`: Modifies an integer array in place.
  * `int sum_array(const int* arr, size_t len);`: Calculates the sum of array elements.
  * `char* caesar_cipher(const char* text, int shift);`: Performs Caesar cipher encryption.
  * `bool is_valid_email(const char* email);`: Basic email format validation.
  * `double measure_time(void (*func)(void*), void* data);`: Benchmarks a given C function.

-----

## 🏗️ **7. Factories or Creators**

This header declares functions that act as creators (e.g., `generate_data`, `generate_sequence`, `generate_random_array`), but it does not define a factory pattern in the object-oriented sense.

-----

## 🧯 **8. Error Handling**

Error handling is implicitly defined by the function signatures. For instance, functions returning `char*` may return `NULL` to indicate an error (as defined in `cgo_utils.c`). The header merely declares these return types, implying the need for error checking on the Go side.

-----

## 🤝 **9. Key Integrations / Dependencies**

  * **`cgo_utils.c`**: This is the corresponding implementation file that provides the actual code for the functions declared here.
  * **Go Programs (`main.go`, `argo.go`)**: Go files that use CGO will `#include "cgo_utils.h"` via `/* #include "cgo_utils.h" */` CGO directives, enabling them to call the declared C functions.

> **Example**:
> This header file is a critical dependency for any Go file that uses CGO to interact with the functions implemented in `cgo_utils.c`, such as `main.go` and `argo.go`.

-----

## 🧪 **10. Expected Usage Patterns**

### 🟢 **Basic Usage (Go Import C)**

```go
// In a Go file:
/*
#include "cgo_utils.h"
*/
import "C"

// ... inside a Go function
cStr := C.CString("hello")
C.to_uppercase(cStr) // Call C function
C.free(unsafe.Pointer(cStr)) // Free C-allocated memory
```

### ⚙️ **Advanced Usage (With Argo Arena)**

```go
// In a Go file (within an argo.WithArena context):
// Assuming 'arena' is the argo.Arena instance
cInput := arena.CString("example") // Arena-managed C string
cResult := C.process_string((*C.char)(unsafe.Pointer(cInput)), C.int(2))
// The Go code handles freeing cResult if it's from C.malloc outside the arena
```

-----

## 🚀 **11. Future Extensibility**

### 🧩 Possible New Functionalities

  * **Adding new C functions**: Simply declare new function prototypes in this header, then implement them in `cgo_utils.c`.
  * **Introducing C data structures**: Define new `struct` types in this header if complex data needs to be passed directly between Go and C.
  * **Exposing C constants or macros**: Define C constants or macros that Go code might need to use.

-----

## 📈 **12. Impact on Performance**

  * **Minimal Overhead**: As a header file, `cgo_utils.h` itself has virtually no runtime performance impact.
  * **Enables Performance**: Its role is to enable the use of performance-optimized C code by Go, thus indirectly contributing to the overall application performance by providing the necessary interface for faster operations.

-----

## ✅ **13. Conclusion**

The `cgo_utils.h` file is fundamental for establishing a clear and reliable interface between the Go and C components of the project. By precisely defining the signatures of the C utility functions, it ensures correct data exchange and compilation, making it possible for Go to leverage the performance benefits of C while maintaining a robust and maintainable codebase. It is the cornerstone for CGO interoperability within this project.
