# ¿Qué hace este benchmark?

1. **Método Tradicional**: Usa `C.CString()` que aloca cada string en el heap de Go (sujeto al GC)
2. **Método Arena**: Usa `arena.CString()` que aloca en memoria C fuera del alcance del GC

## Puntos clave del test:

- **Función C simulada**: `process_data()` que procesa strings (agrega prefijo y convierte a mayúsculas)
- **Múltiples tamaños**: Tests con 10K, 50K, y 100K iteraciones
- **Presión de memoria**: Test adicional que simula alta presión en el heap
- **Medición de memoria**: Monitorea el uso del heap de Go antes/después
- **GC forzado**: Simula condiciones reales con GC periódico

## Lo que deberías ver:

1. **Arena más rápida** cuando hay muchas allocaciones pequeñas
2. **Menor presión en Go heap** con Arena (menos diferencia de memoria)
3. **Mejor rendimiento bajo presión** especialmente en el test de presión de memoria

Para usar este código:

1. Guárdalo en un archivo `main.go` 
2. Asegúrate de que tu paquete `argo` esté disponible
3. Ejecuta: `go run main.go`

**El benchmark debería mostrar que la Arena es significativamente más rápida**, especialmente cuando tienes muchas allocaciones de strings que luego pasas a funciones C. 🚀
