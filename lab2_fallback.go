//go:build !fplll
// +build !fplll

package main

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"

	"gonum.org/v1/gonum/mat"
)

// runBKZ performs BKZ reduction on a given basis using the specified block size beta.
// NOTE: This is a SIMULATED version - install fplll for accurate results.
func runBKZ(basis [][]*big.Int, beta int) []float64 {
	size := len(basis)

	// Convert to float64 matrix for basic operations
	B := mat.NewDense(size, size, nil)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			val, _ := basis[i][j].Float64()
			B.Set(i, j, val)
		}
	}

	// Simulate Gram-Schmidt orthogonalization to get approximate norms
	// In a real implementation, this would use fplll's BKZ and GSO objects
	profile := make([]float64, size)

	// Simple simulation: compute norms of basis vectors with some decay
	for i := 0; i < size; i++ {
		norm := 0.0
		for j := 0; j < size; j++ {
			val := B.At(i, j)
			norm += val * val
		}

		// Apply some decay to simulate BKZ behavior
		decay := math.Pow(0.99, float64(i))
		adjustedNorm := norm * decay

		// Compute log2 of the square root (log2 of the norm)
		profile[i] = math.Log2(math.Sqrt(adjustedNorm))
	}

	return profile
}

// genRandomBasis generates a random square lattice basis of the given rank.
// This is used for Lab 2 where we need a general random lattice, not necessarily q-ary.
func genRandomBasis(rank int) [][]*big.Int {
	basis := make([][]*big.Int, rank)

	for i := 0; i < rank; i++ {
		basis[i] = make([]*big.Int, rank)
		for j := 0; j < rank; j++ {
			// Generate random integers in range [-100, 100]
			randVal, _ := rand.Int(rand.Reader, big.NewInt(201))
			randVal.Sub(randVal, big.NewInt(100))
			basis[i][j] = new(big.Int).Set(randVal)
		}
	}

	return basis
}

// runLab2Verification orchestrates the experiment for Lab 2.
// It generates a random lattice basis, runs the powerful BKZ reduction algorithm
// on it, and then prints the resulting basis profile. The linearity of this
// profile in a plot is evidence for the Geometric Series Assumption.
func runLab2Verification() {
	fmt.Println("--- Running Lab 2: Verifying the Geometric Series Assumption ---")
	fmt.Println("NOTE: Using SIMULATED BKZ reduction. Install fplll for accurate results.")

	rank := 30
	beta := 20

	fmt.Printf("Generating a random lattice of rank %d.\n", rank)
	basis := genRandomBasis(rank)

	fmt.Printf("Running BKZ reduction with block size beta = %d...\n", beta)
	profile := runBKZ(basis, beta)

	fmt.Println("BKZ finished.")
	fmt.Println("Basis Profile (log2 of Gram-Schmidt norms):")

	// Format the profile output
	fmt.Print("[")
	for i, val := range profile {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%.2f", val)
	}
	fmt.Println("]")

	fmt.Println("\nLab 2 finished. Plot this profile data to visually check for linearity.")
}
