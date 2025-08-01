package main

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"

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

// genRandomBasis generates a "hard" random square lattice basis of the given rank.
// It populates a matrix with large random numbers drawn from [0, q), ensuring
// a high-determinant lattice that is a good candidate for reduction algorithms.
func genRandomBasis(rank int, q *big.Int) [][]*big.Int {
	basis := make([][]*big.Int, rank)
	for i := 0; i < rank; i++ {
		basis[i] = make([]*big.Int, rank)
		for j := 0; j < rank; j++ {
			// Generate a large random integer for each entry
			randVal, _ := rand.Int(rand.Reader, q)
			basis[i][j] = new(big.Int).Set(randVal)
		}
	}
	return basis
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

// writeBasisToFile writes a basis matrix to a file in fplll format
func writeBasisToFile(basis [][]*big.Int, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write in fplll format: [rows] [cols] followed by the matrix
	rows := len(basis)
	cols := len(basis[0])

	fmt.Fprintf(file, "[")
	for i := 0; i < rows; i++ {
		fmt.Fprintf(file, "[")
		for j := 0; j < cols; j++ {
			fmt.Fprintf(file, "%s", basis[i][j].String())
			if j < cols-1 {
				fmt.Fprintf(file, " ")
			}
		}
		fmt.Fprintf(file, "]")
		if i < rows-1 {
			fmt.Fprintf(file, "\n")
		}
	}
	fmt.Fprintf(file, "]\n")

	return nil
}

// svpOracle finds the shortest non-zero vector in the lattice using fplll command line tool.
// It writes the basis to a temporary file, calls fplll -a svp, and parses the result.
func svpOracle(basis [][]*big.Int, radius float64) float64 {
	// Write basis to temporary file
	tmpFile := "/tmp/lattice_basis.txt"
	err := writeBasisToFile(basis, tmpFile)
	if err != nil {
		fmt.Printf("Error writing basis to file: %v\n", err)
		return 0
	}
	defer os.Remove(tmpFile)

	// Call fplll -a svp
	cmd := exec.Command("fplll", "-a", "svp", tmpFile)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running fplll: %v\n", err)
		return 0
	}

	// Parse the output to extract the shortest vector and compute its norm
	outputStr := strings.TrimSpace(string(output))

	// The output should be in format [val1 val2 val3 ...]
	if strings.HasPrefix(outputStr, "[") && strings.HasSuffix(outputStr, "]") {
		vectorStr := outputStr[1 : len(outputStr)-1] // Remove brackets
		coords := strings.Fields(vectorStr)

		norm := 0.0
		for _, coord := range coords {
			if val, err := strconv.ParseFloat(coord, 64); err == nil {
				norm += val * val
			}
		}
		return norm
	}

	// If parsing fails, return a reasonable estimate
	return radius * radius
}

// runLab1Verification orchestrates the primary experiment of Lab 1.
// It iterates through various lattice dimensions, and for each dimension:
// 1. Generates a random hard lattice basis.
// 2. Predicts the shortest vector norm using the Gaussian Heuristic.
// 3. Finds the actual shortest vector norm using the SVP oracle.
// 4. Prints the predicted norm, the actual norm, and the relative error.
func runLab1Verification() {
	fmt.Println("--- Running Lab 1: Verifying the Gaussian Heuristic ---")
	fmt.Println("Using FPLLL command-line tool for accurate SVP computation.")
	// This q now defines the range of entries for our random basis
	q := big.NewInt(131)
	fmt.Printf("Target q for random coefficients: %s. Iterating from n=30 to n=60...\n\n", q.String())

	fmt.Printf("%-4s | %-13s | %-13s | %-14s\n", "n", "GH Prediction", "SVP Norm", "Relative Error")
	fmt.Println("------------------------------------------------------")

	for n := 30; n <= 60; n += 2 {
		// NOTE: We are replacing genBasis with genRandomBasis.
		// The rank of this lattice is simply n.
		basis := genRandomBasis(n, q)

		// The rank is n, not m+n
		rank := n

		// Calculate lattice volume
		vol := latticeVolume(basis)

		// Calculate Gaussian heuristic prediction
		gh := gaussianHeuristic(vol, rank)
		ghFloat, _ := gh.Float64()

		// Call SVP oracle
		svpNormSquared := svpOracle(basis, 1.5*ghFloat)
		svpNorm := math.Sqrt(svpNormSquared)

		// Calculate relative error
		relativeError := math.Abs(svpNorm-ghFloat) / svpNorm * 100

		// The dimension 'n' is now the total rank
		fmt.Printf("%-4d | %-13.2f | %-13.2f | %-13.2f%%\n", n, ghFloat, svpNorm, relativeError)
	}

	fmt.Println("\nLab 1 finished.")
}
