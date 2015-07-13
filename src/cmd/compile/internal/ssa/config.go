// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

type Config struct {
	arch       string // "amd64", etc.
	ptrSize    int64  // 4 or 8
	Uintptr    Type   // pointer arithmetic type
	Int        Type
	lowerBlock func(*Block) bool          // lowering function
	lowerValue func(*Value, *Config) bool // lowering function
	fe         Frontend                   // callbacks into compiler frontend

	// TODO: more stuff.  Compiler flags of interest, ...
}

type Frontend interface {
	// StringSym returns a symbol pointing to the given string.
	// Strings are laid out in read-only memory with one word of pointer,
	// one word of length, then the contents of the string.
	StringSym(string) interface{} // returns *gc.Sym

	// Log logs a message from the compiler.
	Logf(string, ...interface{})

	// Fatal reports a compiler error and exits.
	Fatalf(string, ...interface{})

	// Unimplemented reports that the function cannot be compiled.
	// It will be removed once SSA work is complete.
	Unimplementedf(msg string, args ...interface{})
}

// NewConfig returns a new configuration object for the given architecture.
func NewConfig(arch string, fe Frontend) *Config {
	c := &Config{arch: arch, fe: fe}
	switch arch {
	case "amd64":
		c.ptrSize = 8
		c.lowerBlock = rewriteBlockAMD64
		c.lowerValue = rewriteValueAMD64
	case "386":
		c.ptrSize = 4
		c.lowerBlock = rewriteBlockAMD64
		c.lowerValue = rewriteValueAMD64 // TODO(khr): full 32-bit support
	default:
		fe.Unimplementedf("arch %s not implemented", arch)
	}

	// cache the frequently-used types in the config
	c.Uintptr = TypeUInt32
	c.Int = TypeInt32
	if c.ptrSize == 8 {
		c.Uintptr = TypeUInt64
		c.Int = TypeInt64
	}

	return c
}

// NewFunc returns a new, empty function object
func (c *Config) NewFunc() *Func {
	// TODO(khr): should this function take name, type, etc. as arguments?
	return &Func{Config: c}
}

func (c *Config) Logf(msg string, args ...interface{})           { c.fe.Logf(msg, args...) }
func (c *Config) Fatalf(msg string, args ...interface{})         { c.fe.Fatalf(msg, args...) }
func (c *Config) Unimplementedf(msg string, args ...interface{}) { c.fe.Unimplementedf(msg, args...) }

// TODO(khr): do we really need a separate Config, or can we just
// store all its fields inside a Func?