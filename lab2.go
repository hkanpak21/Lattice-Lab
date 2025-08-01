package main

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// runBKZ performs BKZ reduction on a given basis using the fplll command line tool.
// It writes the basis to a temporary file, calls fplll -a bkz, and parses the reduced basis
// to compute the Gram-Schmidt profile using Go's matrix operations.
func runBKZ(basis [][]*big.Int, beta int) []float64 {
	rank := len(basis)

	// Write basis to temporary file
	tmpFile := "/tmp/lattice_basis_bkz.txt"
	err := writeBasisToFile(basis, tmpFile)
	if err != nil {
		fmt.Printf("Error writing basis to file: %v\n", err)
		return make([]float64, rank)
	}
	defer os.Remove(tmpFile)

	// Call fplll -a bkz with specified block size
	cmd := exec.Command("fplll", "-a", "bkz", "-b", strconv.Itoa(beta), tmpFile)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running fplll BKZ: %v\n", err)
		return make([]float64, rank)
	}

	// Parse the reduced basis from output
	reducedBasis := parseMatrixOutput(string(output))

	if reducedBasis == nil || len(reducedBasis) != rank {
		fmt.Printf("Error parsing BKZ output\n")
		return make([]float64, rank)
	}

	// Compute Gram-Schmidt profile
	profile := computeGramSchmidtProfile(reducedBasis)

	return profile
}

// parseMatrixOutput parses the matrix output from fplll
func parseMatrixOutput(output string) [][]*big.Int {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return nil
	}

	var matrix [][]*big.Int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || (!strings.HasPrefix(line, "[") && !strings.Contains(line, "[")) {
			continue
		}

		// Remove brackets and parse row
		line = strings.Trim(line, "[]")
		if line == "" {
			continue
		}

		coords := strings.Fields(line)
		if len(coords) == 0 {
			continue
		}

		var row []*big.Int
		for _, coord := range coords {
			val := new(big.Int)
			if _, ok := val.SetString(coord, 10); ok {
				row = append(row, val)
			}
		}

		if len(row) > 0 {
			matrix = append(matrix, row)
		}
	}

	return matrix
}

// computeGramSchmidtProfile computes the log2 of Gram-Schmidt vector norms
func computeGramSchmidtProfile(basis [][]*big.Int) []float64 {
	n := len(basis)
	if n == 0 {
		return nil
	}

	// Convert to float64 for numerical stability
	B := make([][]float64, n)
	for i := 0; i < n; i++ {
		B[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if j < len(basis[i]) {
				val, _ := basis[i][j].Float64()
				B[i][j] = val
			}
		}
	}

	// Perform Gram-Schmidt orthogonalization
	profile := make([]float64, n)
	orthoBasis := make([][]float64, n)

	for i := 0; i < n; i++ {
		// Copy current vector
		orthoBasis[i] = make([]float64, n)
		copy(orthoBasis[i], B[i])

		// Orthogonalize against previous vectors
		for k := 0; k < i; k++ {
			dot := 0.0
			orthogNormSq := 0.0

			for j := 0; j < n; j++ {
				dot += B[i][j] * orthoBasis[k][j]
				orthogNormSq += orthoBasis[k][j] * orthoBasis[k][j]
			}

			if orthogNormSq > 1e-10 {
				coeff := dot / orthogNormSq
				for j := 0; j < n; j++ {
					orthoBasis[i][j] -= coeff * orthoBasis[k][j]
				}
			}
		}

		// Compute norm and take log2
		norm := 0.0
		for j := 0; j < n; j++ {
			norm += orthoBasis[i][j] * orthoBasis[i][j]
		}

		if norm > 1e-10 {
			profile[i] = math.Log2(math.Sqrt(norm))
		} else {
			profile[i] = -50 // Very small norm
		}
	}

	return profile
}

// runLab2Verification orchestrates the experiment for Lab 2.
// It generates a random lattice basis, runs the powerful BKZ reduction algorithm
// on it, and then prints the resulting basis profile. The linearity of this
// profile in a plot is evidence for the Geometric Series Assumption.
func runLab2Verification() {
	fmt.Println("--- Running Lab 2: Verifying the Geometric Series Assumption ---")
	fmt.Println("Using FPLLL command-line tool for accurate BKZ reduction.")

	rank := 30
	beta := 20
	// Use a large prime for the coefficient range to ensure a "hard" lattice
	q := big.NewInt(100003) // A reasonably large prime

	fmt.Printf("Generating a random lattice of rank %d with coefficients up to %s.\n", rank, q.String())
	// Pass 'q' to the new generator
	basis := genRandomBasis(rank, q)

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
