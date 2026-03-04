//go:build cgo

package coinsetffi

/*
#include <stdint.h>
#include <stdlib.h>

#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/../../cgo-lib/aarch64-apple-darwin
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../../cgo-lib/x86_64-unknown-linux-gnu
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/../../cgo-lib/x86_64-pc-windows-gnu
#cgo LDFLAGS: -lcoinset
#cgo linux LDFLAGS: -lm
#cgo windows LDFLAGS: -lbcrypt -lntdll -lws2_32 -luserenv

char* coinset_inspect(const uint8_t* input_ptr, size_t input_len, uint32_t flags);
char* coinset_clvm_decompile(const uint8_t* input_ptr, size_t input_len, uint32_t flags);
char* coinset_clvm_compile(const uint8_t* input_ptr, size_t input_len, uint32_t flags);
char* coinset_clvm_run(const uint8_t* program_ptr, size_t program_len, const uint8_t* env_ptr, size_t env_len, uint64_t max_cost, uint32_t flags);
char* coinset_clvm_tree_hash(const uint8_t* input_ptr, size_t input_len, uint32_t flags);
char* coinset_clvm_uncurry(const uint8_t* input_ptr, size_t input_len, uint32_t flags);
char* coinset_clvm_curry(const uint8_t* mod_ptr, size_t mod_len, const uint8_t* args_json_ptr, size_t args_json_len, uint32_t flags);
void coinset_free(char* s);
const char* coinset_version(void);
*/
import "C"

import (
	"encoding/json"
	"errors"
	"unsafe"
)

const (
	flagPretty         = 1 << 0
	flagConditionsOnly = 1 << 1
	flagIncludeCost    = 1 << 2
)

func Inspect(input []byte, pretty bool, conditionsOnly bool) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("empty input")
	}
	var flags C.uint32_t
	if pretty {
		flags |= flagPretty
	}
	if conditionsOnly {
		flags |= flagConditionsOnly
	}

	ptr := (*C.uint8_t)(unsafe.Pointer(&input[0]))
	out := C.coinset_inspect(ptr, C.size_t(len(input)), flags)
	return readAndFree(out)
}

func ClvmDecompile(hexBytes string, pretty bool) ([]byte, error) {
	in := []byte(hexBytes)
	if len(in) == 0 {
		return nil, errors.New("empty input")
	}
	var flags C.uint32_t
	if pretty {
		flags |= flagPretty
	}
	ptr := (*C.uint8_t)(unsafe.Pointer(&in[0]))
	out := C.coinset_clvm_decompile(ptr, C.size_t(len(in)), flags)
	return readAndFree(out)
}

func ClvmCompile(program string, pretty bool) ([]byte, error) {
	in := []byte(program)
	if len(in) == 0 {
		return nil, errors.New("empty input")
	}
	var flags C.uint32_t
	if pretty {
		flags |= flagPretty
	}
	ptr := (*C.uint8_t)(unsafe.Pointer(&in[0]))
	out := C.coinset_clvm_compile(ptr, C.size_t(len(in)), flags)
	return readAndFree(out)
}

func ClvmRun(program string, env string, maxCost uint64, includeCost bool, pretty bool) ([]byte, error) {
	p := []byte(program)
	if len(p) == 0 {
		return nil, errors.New("empty program")
	}
	e := []byte(env)
	if len(e) == 0 {
		e = []byte("()")
	}
	var flags C.uint32_t
	if pretty {
		flags |= flagPretty
	}
	if includeCost {
		flags |= flagIncludeCost
	}

	pPtr := (*C.uint8_t)(unsafe.Pointer(&p[0]))
	ePtr := (*C.uint8_t)(unsafe.Pointer(&e[0]))
	out := C.coinset_clvm_run(pPtr, C.size_t(len(p)), ePtr, C.size_t(len(e)), C.uint64_t(maxCost), flags)
	return readAndFree(out)
}

func ClvmTreeHash(input string, pretty bool) ([]byte, error) {
	in := []byte(input)
	if len(in) == 0 {
		return nil, errors.New("empty input")
	}
	var flags C.uint32_t
	if pretty {
		flags |= flagPretty
	}
	ptr := (*C.uint8_t)(unsafe.Pointer(&in[0]))
	out := C.coinset_clvm_tree_hash(ptr, C.size_t(len(in)), flags)
	return readAndFree(out)
}

func ClvmUncurry(input string, pretty bool) ([]byte, error) {
	in := []byte(input)
	if len(in) == 0 {
		return nil, errors.New("empty input")
	}
	var flags C.uint32_t
	if pretty {
		flags |= flagPretty
	}
	ptr := (*C.uint8_t)(unsafe.Pointer(&in[0]))
	out := C.coinset_clvm_uncurry(ptr, C.size_t(len(in)), flags)
	return readAndFree(out)
}

type CurryArg struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type curryPayload struct {
	ModIsHash bool       `json:"mod_is_hash"`
	Args      []CurryArg `json:"args"`
}

func ClvmCurry(modInput string, modIsHash bool, args []CurryArg, pretty bool) ([]byte, error) {
	m := []byte(modInput)
	if len(m) == 0 {
		return nil, errors.New("empty mod")
	}
	payload := curryPayload{ModIsHash: modIsHash, Args: args}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var flags C.uint32_t
	if pretty {
		flags |= flagPretty
	}
	mPtr := (*C.uint8_t)(unsafe.Pointer(&m[0]))
	pPtr := (*C.uint8_t)(unsafe.Pointer(&payloadJSON[0]))
	out := C.coinset_clvm_curry(mPtr, C.size_t(len(m)), pPtr, C.size_t(len(payloadJSON)), flags)
	return readAndFree(out)
}

func readAndFree(out *C.char) ([]byte, error) {
	if out == nil {
		return nil, errors.New("ffi returned null")
	}
	defer C.coinset_free(out)
	return []byte(C.GoString(out)), nil
}
