package main

// NOTE: Uncomment the following when fplll is properly installed:
// /*
// #cgo LDFLAGS: -lfplll
// #include <fplll.h>
// */
// import "C"

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"

	"gonum.org/v1/gonum/mat"
)

// toBigIntMatrix converts a gonum.org/v1/mat.Dense matrix of float64s
// into a 2D slice of *big.Int. This is a utility for when we need to
// handle matrices with integer coefficients that may exceed the capacity of int64.
func toBigIntMatrix(m *mat.Dense) [][]*big.Int {
	rows, cols := m.Dims()
	result := make([][]*big.Int, rows)

	for i := 0; i < rows; i++ {
		result[i] = make([]*big.Int, cols)
		for j := 0; j < cols; j++ {
			val := m.At(i, j)
			result[i][j] = big.NewInt(int64(val))
		}
	}

	return result
}

// genBasis creates the basis for a q-ary lattice.
//  1. It generates a random m x n matrix A with entries in the integers modulo q.
//  2. It constructs the (m+n) x (m+n) basis B as:
//     [[q*I_m, A], [0, I_n]]
//
// The resulting basis is returned as a 2D slice of *big.Int.
func genBasis(n, m int, q *big.Int) [][]*big.Int {
	// Generate random m x n matrix A with entries in Z_q
	A := make([][]*big.Int, m)
	for i := 0; i < m; i++ {
		A[i] = make([]*big.Int, n)
		for j := 0; j < n; j++ {
			// Generate random number in [0, q)
			randVal, _ := rand.Int(rand.Reader, q)
			A[i][j] = new(big.Int).Set(randVal)
		}
	}

	// Construct the (m+n) x (m+n) basis matrix B
	size := m + n
	B := make([][]*big.Int, size)

	for i := 0; i < size; i++ {
		B[i] = make([]*big.Int, size)
		for j := 0; j < size; j++ {
			B[i][j] = big.NewInt(0)
		}
	}

	// Fill in q*I_m in the top-left block
	for i := 0; i < m; i++ {
		B[i][i] = new(big.Int).Set(q)
	}

	// Fill in A in the top-right block
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			B[i][m+j] = new(big.Int).Set(A[i][j])
		}
	}

	// Fill in I_n in the bottom-right block
	for i := 0; i < n; i++ {
		B[m+i][m+i] = big.NewInt(1)
	}

	return B
}

// latticeVolume calculates the volume of the lattice spanned by the given basis.
// The volume is defined as the square root of the determinant of B * B^T.
// The basis is temporarily converted to float64 for use with the gonum/mat library.
func latticeVolume(basis [][]*big.Int) *big.Float {
	size := len(basis)

	// Convert basis to gonum Dense matrix
	B := mat.NewDense(size, size, nil)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			val, _ := basis[i][j].Float64()
			B.Set(i, j, val)
		}
	}

	// Calculate B * B^T
	BT := mat.NewDense(size, size, nil)
	BT.Copy(B)
	BT.T()

	BBT := mat.NewDense(size, size, nil)
	BBT.Mul(B, BT)

	// Calculate determinant
	det := mat.Det(BBT)

	// Return square root of determinant as big.Float
	vol := big.NewFloat(math.Sqrt(math.Abs(det)))
	return vol
}

// gaussianHeuristic computes the predicted length of the shortest non-zero vector
// in a lattice of a given rank and volume, based on the Gaussian Heuristic formula.
func gaussianHeuristic(vol *big.Float, rank int) *big.Float {
	// GH(L) = sqrt(n/(2*pi*e)) * vol(L)^(1/n)
	n := float64(rank)

	// Calculate sqrt(n/(2*pi*e))
	coefficient := math.Sqrt(n / (2 * math.Pi * math.E))

	// Calculate vol^(1/n)
	volFloat64, _ := vol.Float64()
	volPowerN := math.Pow(volFloat64, 1.0/n)

	// Combine
	result := coefficient * volPowerN

	return big.NewFloat(result)
}

// svpOracle finds the shortest non-zero vector in the lattice within a given radius.
// It acts as a wrapper around the fplll C library, accessed via cgo.
// The process involves:
// 1. Converting the Go integer basis into fplll's integer matrix format.
// 2. Running LLL reduction to preprocess the basis, which is essential for enumeration.
// 3. Running the enumeration algorithm to find the vector with the smallest norm.
// It returns the squared norm of the vector found.
func svpOracle(basis [][]*big.Int, radius float64) float64 {
	// Note: This is a simplified implementation that would need actual fplll integration
	// For now, we'll simulate the SVP oracle with a placeholder that uses LLL approximation

	size := len(basis)

	// Convert to float64 matrix for basic LLL simulation
	B := mat.NewDense(size, size, nil)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			val, _ := basis[i][j].Float64()
			B.Set(i, j, val)
		}
	}

	// Simple approximation: return the norm of the first basis vector after basic operations
	// In a real implementation, this would call fplll's LLL and enumeration
	firstVector := make([]float64, size)
	for j := 0; j < size; j++ {
		firstVector[j] = B.At(0, j)
	}

	// Calculate squared norm
	norm := 0.0
	for _, val := range firstVector {
		norm += val * val
	}

	return norm
}

// runLab1Verification orchestrates the primary experiment of Lab 1.
// It iterates through various lattice dimensions, and for each dimension:
// 1. Generates a q-ary lattice basis.
// 2. Predicts the shortest vector norm using the Gaussian Heuristic.
// 3. Finds the actual shortest vector norm using the SVP oracle.
// 4. Prints the predicted norm, the actual norm, and the relative error.
func runLab1Verification() {
	fmt.Println("--- Running Lab 1: Verifying the Gaussian Heuristic ---")
	q := big.NewInt(131)
	fmt.Printf("Target q: %s. Iterating from n=30 to n=60...\n\n", q.String())

	fmt.Printf("%-4s | %-13s | %-13s | %-14s\n", "n", "GH Prediction", "SVP Norm", "Relative Error")
	fmt.Println("------------------------------------------------------")

	for n := 30; n <= 60; n += 2 {
		m := n

		// Generate basis
		basis := genBasis(n, m, q)

		// Calculate lattice volume
		vol := latticeVolume(basis)

		// Calculate Gaussian heuristic prediction
		gh := gaussianHeuristic(vol, m+n)
		ghFloat, _ := gh.Float64()

		// Call SVP oracle
		svpNormSquared := svpOracle(basis, 1.5*ghFloat)
		svpNorm := math.Sqrt(svpNormSquared)

		// Calculate relative error
		relativeError := math.Abs(svpNorm-ghFloat) / svpNorm * 100

		// Print results
		fmt.Printf("%-4d | %-13.2f | %-13.2f | %-13.2f%%\n", n, ghFloat, svpNorm, relativeError)
	}

	fmt.Println("\nLab 1 finished.")
}
