| Feature         | Go + Argo          | Rust + FFI        | C++ FFI         |
| --------------- | ------------------ | ----------------- | --------------- |
| Memory Control  | Scoped Arena       | Manual (but safe) | Manual (unsafe) |
| Interop Safety  | Arena hides danger | Requires `unsafe` | Fully unsafe    |
| Memory Lifetime | Automatic w/ Free  | Manual drop       | Manual free     |
| Performance     | Near-native        | Native            | Native          |
| Ergonomics      | High               | Moderate          | Low             |
