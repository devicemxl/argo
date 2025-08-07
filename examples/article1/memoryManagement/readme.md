Here's a comprehensive test for the `newChunkC` function! This demonstration shows:

## 🔍 What it demonstrates:

1. **Core Functionality**: How `newChunkC` allocates memory using C's `malloc`
2. **GC Independence**: Proves the allocated memory is invisible to Go's garbage collector
3. **Memory Management**: Shows how chunks are linked and managed
4. **Data Integrity**: Verifies data survives aggressive garbage collection
5. **Error Handling**: Tests behavior with oversized allocations

## 🎯 Key Features Tested:

- **C malloc allocation** - Memory allocated outside Go's heap
- **Memory zeroing** - Clean initialization with `memset`
- **Chunk linking** - How multiple chunks form a linked list
- **Offset tracking** - Simple allocation within chunks
- **GC immunity** - Data persists through garbage collection cycles

## 📊 What you'll see:

- Go heap statistics before/after operations
- Actual memory addresses and chunk structures  
- Proof that C memory doesn't affect Go's heap counters
- Data integrity verification after multiple GC cycles

The test creates a simplified version of the Arena's chunk system and demonstrates why allocating via C malloc is so powerful - **the memory is completely invisible to Go's garbage collector**, which eliminates GC pressure and provides predictable performance.

Run it to see exactly how your `newChunkC` function works under the hood! 🚀
