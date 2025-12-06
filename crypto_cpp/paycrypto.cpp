#include "paycrypto.h"
#include <string>
#include <vector>
#include <openssl/hmac.h>
#include <openssl/sha.h>
#include <iomanip>
#include <sstream>
#include <cstring>

// Helper C++ function
std::string hmac_sha256(const std::string& key, const std::string& data) {
    unsigned char hash[EVP_MAX_MD_SIZE];
    unsigned int length = 0;
    
    HMAC(EVP_sha256(), key.c_str(), key.length(), 
         (unsigned char*)data.c_str(), data.length(), hash, &length);
         
    std::stringstream ss;
    for(unsigned int i = 0; i < length; i++) {
        ss << std::hex << std::setw(2) << std::setfill('0') << (int)hash[i];
    }
    return ss.str();
}

extern "C" {

void ComputeHMAC_SHA256(const char* key, const char* data, char* output_hex) {
    if (!key || !data || !output_hex) return;
    std::string res = hmac_sha256(std::string(key), std::string(data));
    std::strcpy(output_hex, res.c_str());
}

void BatchComputeHMAC(int count, HmacRequest* requests) {
    if (count <= 0 || !requests) return;
    
    // In a real heavy scenario, this loop runs in C++ "native" land
    // possibly parallelized with OpenMP or just avoiding cgo switching overhead
    // for thousands of items.
    for (int i = 0; i < count; i++) {
         if (requests[i].key && requests[i].data && requests[i].output_hex) {
             std::string res = hmac_sha256(requests[i].key, requests[i].data);
             std::strcpy(requests[i].output_hex, res.c_str());
         }
    }
}

}
