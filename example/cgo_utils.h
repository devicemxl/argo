#ifndef CGO_UTILS_H
#define CGO_UTILS_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// === Funciones de procesamiento de strings ===

// Procesa un string repitiéndolo 'multiplier' veces
char* process_string(const char* input, int multiplier);

// Concatena dos strings
char* concat_strings(const char* str1, const char* str2);

// Convierte string a mayúsculas
char* to_uppercase(const char* input);

// Convierte string a minúsculas
char* to_lowercase(const char* input);

// === Funciones de procesamiento de arrays ===

// Multiplica todos los elementos de un array por un factor
void process_array(int* arr, size_t len, int factor);

// Suma todos los elementos de un array
int sum_array(const int* arr, size_t len);

// Encuentra el valor máximo en un array
int max_array(const int* arr, size_t len);

// Encuentra el valor mínimo en un array
int min_array(const int* arr, size_t len);

// Ordena un array en orden ascendente
void sort_array(int* arr, size_t len);

// === Funciones de generación de datos ===

// Genera datos aleatorios
char* generate_data(size_t size);

// Genera array de números secuenciales
int* generate_sequence(int start, int count);

// Genera array de números aleatorios
int* generate_random_array(size_t count, int min, int max);

// === Funciones de manipulación de memoria ===

// Copia datos de manera segura
void safe_memcpy(void* dest, const void* src, size_t n);

// Rellena memoria con un valor
void safe_memset(void* ptr, int value, size_t n);

// Compara dos bloques de memoria
int safe_memcmp(const void* ptr1, const void* ptr2, size_t n);

// === Funciones de procesamiento de archivos ===

// Lee un archivo completo en memoria
char* read_file(const char* filename, size_t* size);

// Escribe datos a un archivo
bool write_file(const char* filename, const char* data, size_t size);

// === Funciones de utilidad ===

// Calcula hash simple de una cadena
uint32_t simple_hash(const char* str);

// Verifica si una cadena es un número válido
bool is_valid_number(const char* str);

// Convierte string a entero de manera segura
bool safe_str_to_int(const char* str, int* result);

// === Funciones de procesamiento de arrays de strings ===

// Procesa un array de strings aplicando una función
char** process_string_array(char** strings, size_t count, char* (*processor)(const char*));

// Libera un array de strings
void free_string_array(char** strings, size_t count);

// === Funciones de cálculo matemático ===

// Calcula potencia de manera eficiente
double fast_pow(double base, int exp);

// Calcula raíz cuadrada usando método de Newton
double sqrt_newton(double x);

// Calcula factorial
uint64_t factorial(int n);

// === Funciones de procesamiento de bits ===

// Cuenta bits establecidos en un número
int count_bits(uint32_t n);

// Invierte bits de un número
uint32_t reverse_bits(uint32_t n);

// === Funciones de criptografía simple ===

// Cifrado César simple
char* caesar_cipher(const char* text, int shift);

// Descifrado César
char* caesar_decipher(const char* text, int shift);

// === Funciones de validación ===

// Valida formato de email (básico)
bool is_valid_email(const char* email);

// Valida formato de URL (básico)
bool is_valid_url(const char* url);

// === Funciones de conversión ===

// Convierte bytes a string hexadecimal
char* bytes_to_hex(const unsigned char* bytes, size_t len);

// Convierte string hexadecimal a bytes
unsigned char* hex_to_bytes(const char* hex, size_t* len);

// === Funciones de benchmarking ===

// Mide tiempo de ejecución de una función
double measure_time(void (*func)(void*), void* data);

// === Funciones de logging ===

// Log con timestamp
void log_message(const char* level, const char* message);

// Log con formato
void log_formatted(const char* level, const char* format, ...);

#ifdef __cplusplus
}
#endif

#endif // CGO_UTILS_H