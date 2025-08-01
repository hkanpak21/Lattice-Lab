# To-Do List: Final `fplll` Integration

**Objective:** Replace the simulated `svpOracle` and `runBKZ` functions in `lab1.go` and `lab2.go` with full `fplll` implementations via `cgo`. This will produce accurate, scientifically valid results.

## Part 1: Create a Common `cgo` Helper File

To avoid code duplication, we'll create a new file for common `cgo` functions, like converting our Go basis into the `fplll` matrix format.

1.  **Create `fplll_helpers.go`:**
    Create a new file in your project directory named `fplll_helpers.go`.

2.  **Add `cgo` Preamble and Imports to `fplll_helpers.go`:**
    Paste the following code at the top of the new file. This sets up `cgo` and includes the necessary Go packages.

    ```go
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
    ```
    *Note the addition of `-lgmp` to the linker flags, which is required by `fplll`.*

3.  **Implement the Basis Conversion Helper:**
    Add the following function to `fplll_helpers.go`. This function will be called by both labs to convert the Go basis into a C `fplll` matrix, handling all the complex memory management.

    -   **Instruction:** Implement the function `createFplllMatrix(basis [][]*big.Int) *C.fplll_int_matrix`.

    -   **Add this comment and code:**
        ```go
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
        ```

## Part 2: Finalize `lab1.go` (SVP Oracle)

Now, let's replace the simulated `svpOracle` with the real one.

1.  **Uncomment `cgo` Preamble in `lab1.go`:**
    Find the `cgo` block at the top of `lab1.go` and uncomment it so it's active. Add the `-lgmp` flag.

    ```go
    // Change this:
    // /*
    // #cgo LDFLAGS: -lfplll
    // #include <fplll.h>
    // */
    // import "C"
    
    // To this:
    /*
    #cgo LDFLAGS: -lfplll -lgmp
    #include <fplll.h>
    */
    import "C"
    ```

2.  **Replace `svpOracle` Function:**
    Delete the *entire* existing `svpOracle` function in `lab1.go` and replace it with this real implementation.

    -   **Instruction:** Replace the function `svpOracle` with the new version.

    -   **Add this comment and code:**
        ```go
        // svpOracle finds the shortest non-zero vector in the lattice within a given radius.
        // It acts as a wrapper around the fplll C library, accessed via cgo.
        // The process involves:
        // 1. Converting the Go integer basis into fplll's integer matrix format.
        // 2. Running LLL reduction to preprocess the basis, which is essential for enumeration.
        // 3. Running the enumeration algorithm to find the vector with the smallest norm.
        // It returns the squared norm of the vector found.
        func svpOracle(basis [][]*big.Int, radius float64) float64 {
            // 1. Convert Go basis to fplll C matrix
            cBasis := createFplllMatrix(basis)
            defer C.fplll_int_matrix_free(cBasis) // Ensure memory is freed!

            // 2. Create GSO object from the matrix for LLL and Enumeration
            gso := C.fplll_gso_init(cBasis, C.FPLLL_GSO_DEFAULT)
            defer C.fplll_gso_free(gso) // Ensure GSO is freed!
            C.gso_update(gso)

            // 3. Run LLL reduction to preprocess the basis
            C.lll_reduction(gso)

            // 4. Run enumeration to find the shortest vector
            var svp_sol C.fplll_svp_solution
            squaredRadius := C.double(radius * radius)
            
            // Call enumeration. The '1' requests a single solution (the shortest).
            C.enumeration(gso, &svp_sol, 0, C.int(len(basis)), squaredRadius, 0, C.ENUM_MODE_FIND_SHORTEST, 1)

            // 5. Extract and return the squared norm from the solution
            return float64(svp_sol.norm)
        }
        ```

## Part 3: Finalize `lab2.go` (BKZ Reduction)

Next, we will replace the simulated `runBKZ` function.

1.  **Uncomment `cgo` Preamble in `lab2.go`:**
    Do the same as in `lab1.go`: find and uncomment the `cgo` block, adding the `-lgmp` flag.

    ```go
    /*
    #cgo LDFLAGS: -lfplll -lgmp
    #include <fplll.h>
    */
    import "C"
    ```

2.  **Replace `runBKZ` Function:**
    Delete the *entire* existing `runBKZ` function in `lab2.go` and replace it with this real implementation.

    -   **Instruction:** Replace the function `runBKZ` with the new version.

    -   **Add this comment and code:**
        ```go
        // runBKZ performs BKZ reduction on a given basis using the specified block size beta.
        // This function is a wrapper around the fplll C library.
        // The process is:
        // 1. Convert the Go basis to an fplll matrix.
        // 2. Call the BKZ reduction algorithm.
        // 3. After reduction, extract the squared norms of the Gram-Schmidt vectors.
        // 4. Compute the final profile as the log base 2 of the norms (log(||b_i*||)).
        // The resulting profile is returned for later analysis/plotting.
        func runBKZ(basis [][]*big.Int, beta int) []float64 {
            rank := len(basis)

            // 1. Convert Go basis to fplll C matrix
            cBasis := createFplllMatrix(basis)
            defer C.fplll_int_matrix_free(cBasis)

            // 2. Create GSO object
            gso := C.fplll_gso_init(cBasis, C.FPLLL_GSO_DEFAULT)
            defer C.fplll_gso_free(gso)
            C.gso_update(gso)

            // 3. Set up BKZ parameters
            params := C.fplll_bkz_param_init()
            defer C.fplll_bkz_param_free(params)
            params.block_size = C.int(beta)
            // Use default strategies for the algorithm
            C.fplll_bkz_param_set_strategies(params, C.int(beta), C.BKZ_DEFAULT_STRATEGY)


            // 4. Run BKZ reduction
            C.bkz_reduction(gso, params)
            C.gso_update(gso) // Update GSO object with the reduced basis info

            // 5. Extract the profile (log2 of Gram-Schmidt vector norms)
            profile := make([]float64, rank)
            for i := 0; i < rank; i++ {
                // gso_get_r_d returns the squared norm ||b_i*||^2
                squaredNorm := C.gso_get_r_d(gso, C.int(i), C.int(i))
                norm := math.Sqrt(float64(squaredNorm))
                profile[i] = math.Log2(norm)
            }

            return profile
        }
        ```

## Part 4: Final Build and Verification

You have now completed the full implementation. The final step is to build and run the code to see the accurate results.

1.  **Build the Executable:**
    Open your terminal in the project directory and run the build command. `cgo` will link against `fplll` and `gmp`.
    ```bash
    go build -o lattice-labs
    ```

2.  **Run the Experiment:**
    Execute the compiled program.
    ```bash
    ./lattice-labs
    ```

3.  **Verify the Output:**
    The output should now be dramatically different from the simulation. Check for the following signs of success:

    -   **For Lab 1:** The "Relative Error" should be much smaller, typically under 15-20%. The predicted and actual SVP norms should be reasonably close.
    -   **For Lab 2:** The "Basis Profile" should show a clear, monotonically decreasing, and approximately linear trend.

    **Example of Expected Final Output:**
    ```text
    === Lattice Heuristics Lab Implementation ===

    --- Running Lab 1: Verifying the Gaussian Heuristic ---
    Target q: 131. Iterating from n=30 to n=60...

    n    | GH Prediction | SVP Norm      | Relative Error
    ------------------------------------------------------
    30   | 21.45         | 22.18         | 3.29%
    32   | 22.16         | 21.95         | 0.96%
    34   | 22.86         | 23.51         | 2.76%
    ...  | ...           | ...           | ...
    58   | 26.54         | 25.89         | 2.51%
    60   | 27.18         | 28.01         | 2.96%

    Lab 1 finished.

    --- Running Lab 2: Verifying the Geometric Series Assumption ---
    Generating a random lattice of rank 30.
    Running BKZ reduction with block size beta = 20...
    BKZ finished.
    Basis Profile (log2 of Gram-Schmidt norms):
    [5.01, 4.88, 4.75, 4.63, 4.50, 4.38, 4.26, 4.14, 4.02, 3.90, 3.78, 3.66, 3.55, 3.43, 3.32, 3.20, 3.09, 2.98, 2.86, 2.75, 2.64, 2.53, 2.41, 2.30, 2.19, 2.08, 1.96, 1.85, 1.74, 1.62]

    Lab 2 finished. Plot this profile data to visually check for linearity.

    === All experiments completed ===
    ```