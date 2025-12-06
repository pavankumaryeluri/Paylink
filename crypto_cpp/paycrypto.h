#pragma once

#ifdef __cplusplus
extern "C" {
#endif

// Example function: Batch HMAC-SHA256 signature generation/verification
// In a real scenario, this would handle batch operations to minimize cgo calls
// For this blueprint, we implement a simple single-item helper to prove the point

typedef struct {
    const char* key;
    const char* data;
    char* output_hex; // Pre-allocated buffer by caller, size needs to be 64+1
} HmacRequest;

void ComputeHMAC_SHA256(const char* key, const char* data, char* output_hex);

// Batch processing simulation
// count: number of items
// requests: array of HmacRequest
void BatchComputeHMAC(int count, HmacRequest* requests);

#ifdef __cplusplus
}
#endif
