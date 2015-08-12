//go:cgo_ldflag "-g"
//go:cgo_ldflag "-O2"
// Created by cgo - DO NOT EDIT

package main

import "unsafe"

import _ "runtime/cgo"

import "syscall"

var _ syscall.Errno
func _Cgo_ptr(ptr unsafe.Pointer) unsafe.Pointer { return ptr }

//go:linkname _Cgo_always_false runtime.cgoAlwaysFalse
var _Cgo_always_false bool
//go:linkname _Cgo_use runtime.cgoUse
func _Cgo_use(interface{})
type _Ctype_int int32

type _Ctype_void [0]byte

//go:linkname _cgo_runtime_cgocall runtime.cgocall
func _cgo_runtime_cgocall(unsafe.Pointer, uintptr) int32

//go:linkname _cgo_runtime_cmalloc runtime.cmalloc
func _cgo_runtime_cmalloc(uintptr) unsafe.Pointer

//go:linkname _cgo_runtime_cgocallback runtime.cgocallback
func _cgo_runtime_cgocallback(unsafe.Pointer, unsafe.Pointer, uintptr)

//go:cgo_import_static _cgo_5b040d13e328_Cfunc_getint
//go:linkname __cgofn__cgo_5b040d13e328_Cfunc_getint _cgo_5b040d13e328_Cfunc_getint
var __cgofn__cgo_5b040d13e328_Cfunc_getint byte
var _cgo_5b040d13e328_Cfunc_getint = unsafe.Pointer(&__cgofn__cgo_5b040d13e328_Cfunc_getint)

func _Cfunc_getint() (r1 _Ctype_int) {
	_cgo_runtime_cgocall(_cgo_5b040d13e328_Cfunc_getint, uintptr(unsafe.Pointer(&r1)))
	if _Cgo_always_false {
	}
	return
}
