//go:build !cgo

package coinsetffi

import "errors"

func Inspect(_ []byte, _ bool, _ bool) ([]byte, error) {
	return nil, errors.New("coinset ffi requires cgo-enabled build")
}

func ClvmDecompile(_ string, _ bool) ([]byte, error) {
	return nil, errors.New("coinset ffi requires cgo-enabled build")
}

func ClvmCompile(_ string, _ bool) ([]byte, error) {
	return nil, errors.New("coinset ffi requires cgo-enabled build")
}

func ClvmRun(_ string, _ string, _ uint64, _ bool, _ bool) ([]byte, error) {
	return nil, errors.New("coinset ffi requires cgo-enabled build")
}

func ClvmTreeHash(_ string, _ bool) ([]byte, error) {
	return nil, errors.New("coinset ffi requires cgo-enabled build")
}

func ClvmUncurry(_ string, _ bool) ([]byte, error) {
	return nil, errors.New("coinset ffi requires cgo-enabled build")
}

func ClvmCurry(_ string, _ []string, _ bool) ([]byte, error) {
	return nil, errors.New("coinset ffi requires cgo-enabled build")
}
