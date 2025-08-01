//go:build !fplll
// +build !fplll

package main

import (
	"math/big"
)

// createFplllMatrix is a fallback stub when fplll is not available
// This should never be called in the fallback builds
func createFplllMatrix(basis [][]*big.Int) interface{} {
	// This function should never be called in fallback mode
	// since the fallback implementations don't use fplll
	panic("createFplllMatrix called in fallback mode - this should not happen")
}
