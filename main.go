package main

import "fmt"

// main is the entry point of the program. It executes the verification
// experiments for both Lab 1 and Lab 2 in sequence and prints the
// results to standard output in a formatted log.
func main() {
	fmt.Println("=== Lattice Heuristics Lab Implementation ===")
	fmt.Println()

	// Run Lab 1: Gaussian Heuristic Verification
	runLab1Verification()

	fmt.Println()

	// Run Lab 2: Geometric Series Assumption Verification
	runLab2Verification()

	fmt.Println()
	fmt.Println("=== All experiments completed ===")
}
