//go:build cgo && fplll
// +build cgo,fplll

package main

/*
#cgo LDFLAGS: -lfplll -lgmp
#include <fplll.h>
#include <gmp.h>
#include <stdlib.h> // For C.CString and C.free

// Helper function to set a value in an fplll_int_matrix from a string
void set_matrix_entry_from_str(fplll_int_matrix A, int i, int j, const char* s) {
    mpz_set_str(A->Z[i][j], s, 10);
}
*/
import "C"
import (
	"math/big"
	"unsafe"
)

// createFplllMatrix converts a Go basis (2D slice of *big.Int) into a C fplll_int_matrix.
// It handles memory allocation for the C matrix and conversion of each big integer.
// The caller is responsible for freeing the returned C matrix using fplll_int_matrix_free.
func createFplllMatrix(basis [][]*big.Int) *C.fplll_int_matrix {
	rows := C.int(len(basis))
	cols := C.int(len(basis[0]))

	// 1. Allocate the fplll integer matrix in C
	cBasis := C.fplll_int_matrix_init(rows, cols)

	// 2. Iterate through the Go matrix and set values in the C matrix
	for i := 0; i < len(basis); i++ {
		for j := 0; j < len(basis[i]); j++ {
			// Convert big.Int to a C string
			valStr := basis[i][j].String()
			cStr := C.CString(valStr)

			// Set the value in the C matrix using our C helper function
			C.set_matrix_entry_from_str(cBasis, C.int(i), C.int(j), cStr)

			// Free the temporary C string
			C.free(unsafe.Pointer(cStr))
		}
	}

	return cBasis
}
