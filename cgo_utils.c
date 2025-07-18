#include "cgo_utils.h"
#include <time.h>
#include <ctype.h>
#include <stdarg.h>
#include <math.h>

// === Implementación de funciones de strings ===

char* process_string(const char* input, int multiplier) {
    if (!input || multiplier <= 0) return NULL;
    
    size_t len = strlen(input);
    char* result = malloc(len * multiplier + 1);
    if (!result) return NULL;
    
    result[0] = '\0';
    for (int i = 0; i < multiplier; i++) {
        strcat(result, input);
    }
    
    return result;
}

char* concat_strings(const char* str1, const char* str2) {
    if (!str1 || !str2) return NULL;
    
    size_t len1 = strlen(str1);
    size_t len2 = strlen(str2);
    
    char* result = malloc(len1 + len2 + 1);
    if (!result) return NULL;
    
    strcpy(result, str1);
    strcat(result, str2);
    
    return result;
}

char* to_uppercase(const char* input) {
    if (!input) return NULL;
    
    size_t len = strlen(input);
    char* result = malloc(len + 1);
    if (!result) return NULL;
    
    for (size_t i = 0; i < len; i++) {
        result[i] = toupper(input[i]);
    }
    result[len] = '\0';
    
    return result;
}

char* to_lowercase(const char* input) {
    if (!input) return NULL;
    
    size_t len = strlen(input);
    char* result = malloc(len + 1);
    if (!result) return NULL;
    
    for (size_t i = 0; i < len; i++) {
        result[i] = tolower(input[i]);
    }
    result[len] = '\0';
    
    return result;
}

// === Implementación de funciones de arrays ===

void process_array(int* arr, size_t len, int factor) {
    if (!arr) return;
    
    for (size_t i = 0; i < len; i++) {
        arr[i] *= factor;
    }
}

int sum_array(const int* arr, size_t len) {
    if (!arr) return 0;
    
    int sum = 0;
    for (size_t i = 0; i < len; i++) {
        sum += arr[i];
    }
    return sum;
}

int max_array(const int* arr, size_t len) {
    if (!arr || len == 0) return 0;
    
    int max = arr[0];
    for (size_t i = 1; i < len; i++) {
        if (arr[i] > max) max = arr[i];
    }
    return max;
}

int min_array(const int* arr, size_t len) {
    if (!arr || len == 0) return 0;
    
    int min = arr[0];
    for (size_t i = 1; i < len; i++) {
        if (arr[i] < min) min = arr[i];
    }
    return min;
}

void sort_array(int* arr, size_t len) {
    if (!arr || len < 2) return;
    
    // Bubble sort simple
    for (size_t i = 0; i < len - 1; i++) {
        for (size_t j = 0; j < len - i - 1; j++) {
            if (arr[j] > arr[j + 1]) {
                int temp = arr[j];
                arr[j] = arr[j + 1];
                arr[j + 1] = temp;
            }
        }
    }
}

// === Implementación de funciones de generación ===

char* generate_data(size_t size) {
    if (size == 0) return NULL;
    
    char* data = malloc(size);
    if (!data) return NULL;
    
    for (size_t i = 0; i < size; i++) {
        data[i] = 'A' + (i % 26);
    }
    
    return data;
}

int* generate_sequence(int start, int count) {
    if (count <= 0) return NULL;
    
    int* arr = malloc(count * sizeof(int));
    if (!arr) return NULL;
    
    for (int i = 0; i < count; i++) {
        arr[i] = start + i;
    }
    
    return arr;
}

int* generate_random_array(size_t count, int min, int max) {
    if (count == 0 || min >= max) return NULL;
    
    int* arr = malloc(count * sizeof(int));
    if (!arr) return NULL;
    
    srand(time(NULL));
    for (size_t i = 0; i < count; i++) {
        arr[i] = min + rand() % (max - min);
    }
    
    return arr;
}

// === Implementación de funciones de memoria ===

void safe_memcpy(void* dest, const void* src, size_t n) {
    if (!dest || !src || n == 0) return;
    memcpy(dest, src, n);
}

void safe_memset(void* ptr, int value, size_t n) {
    if (!ptr || n == 0) return;
    memset(ptr, value, n);
}

int safe_memcmp(const void* ptr1, const void* ptr2, size_t n) {
    if (!ptr1 || !ptr2 || n == 0) return 0;
    return memcmp(ptr1, ptr2, n);
}

// === Implementación de funciones de archivo ===

char* read_file(const char* filename, size_t* size) {
    if (!filename || !size) return NULL;
    
    FILE* file = fopen(filename, "rb");
    if (!file) return NULL;
    
    fseek(file, 0, SEEK_END);
    *size = ftell(file);
    fseek(file, 0, SEEK_SET);
    
    char* buffer = malloc(*size + 1);
    if (!buffer) {
        fclose(file);
        return NULL;
    }
    
    fread(buffer, 1, *size, file);
    buffer[*size] = '\0';
    
    fclose(file);
    return buffer;
}

bool write_file(const char* filename, const char* data, size_t size) {
    if (!filename || !data || size == 0) return false;
    
    FILE* file = fopen(filename, "wb");
    if (!file) return false;
    
    size_t written = fwrite(data, 1, size, file);
    fclose(file);
    
    return written == size;
}

// === Implementación de funciones de utilidad ===

uint32_t simple_hash(const char* str) {
    if (!str) return 0;
    
    uint32_t hash = 5381;
    int c;
    
    while ((c = *str++)) {
        hash = ((hash << 5) + hash) + c; // hash * 33 + c
    }
    
    return hash;
}

bool is_valid_number(const char* str) {
    if (!str || *str == '\0') return false;
    
    if (*str == '-' || *str == '+') str++;
    if (*str == '\0') return false;
    
    while (*str) {
        if (!isdigit(*str)) return false;
        str++;
    }
    
    return true;
}

bool safe_str_to_int(const char* str, int* result) {
    if (!str || !result) return false;
    
    if (!is_valid_number(str)) return false;
    
    *result = atoi(str);
    return true;
}

// === Implementación de funciones de arrays de strings ===

char** process_string_array(char** strings, size_t count, char* (*processor)(const char*)) {
    if (!strings || count == 0 || !processor) return NULL;
    
    char** result = malloc(count * sizeof(char*));
    if (!result) return NULL;
    
    for (size_t i = 0; i < count; i++) {
        result[i] = processor(strings[i]);
    }
    
    return result;
}

void free_string_array(char** strings, size_t count) {
    if (!strings) return;
    
    for (size_t i = 0; i < count; i++) {
        free(strings[i]);
    }
    free(strings);
}

// === Implementación de funciones matemáticas ===

double fast_pow(double base, int exp) {
    if (exp == 0) return 1.0;
    if (exp == 1) return base;
    
    double result = 1.0;
    int abs_exp = abs(exp);
    
    while (abs_exp > 0) {
        if (abs_exp & 1) {
            result *= base;
        }
        base *= base;
        abs_exp >>= 1;
    }
    
    return exp < 0 ? 1.0 / result : result;
}

double sqrt_newton(double x) {
    if (x < 0) return -1; // Error
    if (x == 0) return 0;
    
    double guess = x / 2.0;
    double prev_guess;
    
    do {
        prev_guess = guess;
        guess = (guess + x / guess) / 2.0;
    } while (fabs(guess - prev_guess) > 1e-10);
    
    return guess;
}

uint64_t factorial(int n) {
    if (n < 0) return 0;
    if (n <= 1) return 1;
    
    uint64_t result = 1;
    for (int i = 2; i <= n; i++) {
        result *= i;
    }
    return result;
}

// === Implementación de funciones de bits ===

int count_bits(uint32_t n) {
    int count = 0;
    while (n) {
        count += n & 1;
        n >>= 1;
    }
    return count;
}

uint32_t reverse_bits(uint32_t n) {
    uint32_t result = 0;
    for (int i = 0; i < 32; i++) {
        result = (result << 1) | (n & 1);
        n >>= 1;
    }
    return result;
}

// === Implementación de funciones de criptografía ===

// Definición de caesar_cipher
char* caesar_cipher(const char* text, int shift) {
    if (!text) return NULL;

    size_t len = strlen(text);
    char* result = malloc(len + 1);
    if (!result) return NULL;

    // Normalizar shift para que esté en el rango 0-25
    shift = shift % 26;
    if (shift < 0) {
        shift += 26;
    }

    for (size_t i = 0; i < len; i++) {
        char c = text[i];
        if (c >= 'A' && c <= 'Z') {
            result[i] = 'A' + (c - 'A' + shift) % 26;
        } else if (c >= 'a' && c <= 'z') {
            result[i] = 'a' + (c - 'a' + shift) % 26;
        } else {
            result[i] = c; // Mantener caracteres no alfabéticos tal cual
        }
    }
    result[len] = '\0';
    return result;
}

// Definición de caesar_decipher (incluida en tu código)
char* caesar_decipher(const char* text, int shift) {
    return caesar_cipher(text, -shift);
}

// === Implementación de funciones de validación ===

bool is_valid_email(const char* email) {
    if (!email) return false;
    
    const char* at = strchr(email, '@');
    if (!at) return false;
    
    const char* dot = strchr(at, '.');
    if (!dot) return false;
    
    return dot > at + 1 && strlen(dot) > 1;
}

bool is_valid_url(const char* url) {
    if (!url) return false;
    
    return strncmp(url, "http://", 7) == 0 || 
           strncmp(url, "https://", 8) == 0;
}

// === Implementación de funciones de conversión ===

char* bytes_to_hex(const unsigned char* bytes, size_t len) {
    if (!bytes || len == 0) return NULL;
    
    char* hex = malloc(len * 2 + 1);
    if (!hex) return NULL;
    
    for (size_t i = 0; i < len; i++) {
        sprintf(hex + i * 2, "%02x", bytes[i]);
    }
    
    return hex;
}

// Convierte string hexadecimal a bytes
unsigned char* hex_to_bytes(const char* hex, size_t* len) {
    if (!hex || strlen(hex) % 2 != 0) {
        if (len) { // Se asegura de que 'len' no sea NULL antes de desreferenciarlo
            *len = 0;
        }
        return NULL;
    }

    size_t hex_len = strlen(hex);
    *len = hex_len / 2;

    unsigned char* bytes = malloc(*len);
    if (!bytes) {
        if (len) { // Se asegura de que 'len' no sea NULL antes de desreferenciarlo
            *len = 0;
        }
        return NULL;
    }

    for (size_t i = 0; i < *len; i++) {
        sscanf(hex + i * 2, "%2hhx", &bytes[i]);
    }
    return bytes; // ¡Añadir este retorno!
}