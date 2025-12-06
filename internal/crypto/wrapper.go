package crypto

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lpaycrypto
#include <stdlib.h>
#include <paycrypto.h>
*/
import "C"
import (
	"unsafe"
)

// ComputeHMAC uses the C++ shared library to compute HMAC-SHA256
// This is a demonstration of bridging Go and C++ for "heavy" operations
func ComputeHMAC(key, data string) string {
	cKey := C.CString(key)
	cData := C.CString(data)
	// output hex sha256 is 64 chars + null terminator
	cOutput := C.malloc(65)

	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cData))
	defer C.free(unsafe.Pointer(cOutput))

	C.ComputeHMAC_SHA256(cKey, cData, (*C.char)(cOutput))

	return C.GoString((*C.char)(cOutput))
}
