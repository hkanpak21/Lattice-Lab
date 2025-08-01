# To-Do List: Lattice Heuristics Lab in Go

**Objective:** Implement the "Verifying the Gaussian Heuristic" (Lab 1) and "Verifying the Geometric Series Assumption" (Lab 2) exercises in the Go programming language. The implementation will rely on the `gonum/mat` library for matrix operations and will use `cgo` to interface with the C-based `fplll` library for advanced lattice algorithms (LLL, Enumeration, BKZ), as there is no mature native Go equivalent.

## Part 0: Environment Setup

1.  **Install Go:** Ensure you have a recent version of the Go programming language installed and that your `GOPATH` and `GOROOT` environment variables are correctly configured.

2.  **Install a C Compiler:** A C compiler like GCC is required for `cgo`.
    *   On Debian/Ubuntu: `sudo apt-get update && sudo apt-get install build-essential`
    *   On macOS: Install Xcode Command Line Tools: `xcode-select --install`
    *   On Windows: Install TDM-GCC or MinGW-w64.

3.  **Install `fplll` Library:** Install the `fplll` library and its dependencies, as it will be the backend for our lattice operations.
    ```bash
    # Install dependencies (on Debian/Ubuntu)
    sudo apt-get install libgmp-dev libmpfr-dev

    # Clone and install fplll
    git clone https://github.com/fplll/fplll.git
    cd fplll
    ./autogen.sh
    ./configure
    make
    sudo make install
    ```
    This will install the necessary header files and the static/dynamic library (`libfplll.so` or `libfplll.a`) in a system-wide location (like `/usr/local/lib`).

4.  **Initialize Go Project:** Create a new directory for the project and initialize a Go module.
    ```bash
    mkdir go-lattice-labs
    cd go-lattice-labs
    go mod init lattice-labs
    go get gonum.org/v1/gonum/mat
    ```

## Part 1: Implementing Lab 1 - The Gaussian Heuristic

Create a file named `lab1.go`. You will implement the following functions inside it.

### 1.1. Helper: Matrix Conversion

-   **Instruction:** Create a helper function to convert a `gonum/mat.Dense` matrix (which uses `float64`) to a 2D slice of `*big.Int`. This is because our lattice basis will consist of large integers, but `fplll`'s C interface will need a format we can easily pass. Start by representing the basis as `[][]*big.Int`.

-   **Add this comment:**
    ```go
    // toBigIntMatrix converts a gonum.org/v1/mat.Dense matrix of float64s
    // into a 2D slice of *big.Int. This is a utility for when we need to
    // handle matrices with integer coefficients that may exceed the capacity of int64.
    ```

### 1.2. Generating Matrix A and the Lattice Basis

-   **Instruction:** Implement a Go function `genBasis(n, m, q)` that performs two steps:
    1.  Generates a random `m x n` matrix `A` with entries in $\mathbb{Z}_q$. Use `crypto/rand` to generate random numbers for cryptographic security.
    2.  Constructs the `(m+n) x (m+n)` integer basis matrix `B` for the lattice $\Lambda_q(\mathbf{A})$ as described in the lab:
        $
        \mathbf{B} = \begin{pmatrix} q \mathbf{I}_m & \mathbf{A} \\ \mathbf{0} & \mathbf{I}_n \end{pmatrix}
        $
    The final basis `B` should be a 2D slice of `*big.Int` (`[][]*big.Int`).

-   **Add this comment:**
    ```go
    // genBasis creates the basis for a q-ary lattice.
    // 1. It generates a random m x n matrix A with entries in the integers modulo q.
    // 2. It constructs the (m+n) x (m+n) basis B as:
    //    [[q*I_m, A], [0, I_n]]
    // The resulting basis is returned as a 2D slice of *big.Int.
    ```

### 1.3. Lattice Volume Calculation

-   **Instruction:** Implement a function `latticeVolume(basis [][]*big.Int)`. This function will:
    1.  Convert the `[][]*big.Int` basis into a `gonum/mat.Dense` matrix of `float64`. **Note:** This is a simplification. For true precision, all calculations should use `*big.Float`, but for this lab, `float64` is acceptable.
    2.  Calculate $\mathbf{B}\mathbf{B}^T$.
    3.  Compute the determinant of the result.
    4.  Return the square root of the determinant.

-   **Add this comment:**
    ```go
    // latticeVolume calculates the volume of the lattice spanned by the given basis.
    // The volume is defined as the square root of the determinant of B * B^T.
    // The basis is temporarily converted to float64 for use with the gonum/mat library.
    ```

### 1.4. Gaussian Heuristic Calculation

-   **Instruction:** Implement the function `gaussianHeuristic(vol *big.Float, rank int)`. This function should calculate the expected norm of the shortest vector using the formula:
    $ \text{GH}(\mathcal{L}) = \sqrt{\frac{n}{2\pi e}} \left( \text{vol}(\mathcal{L}) \right)^{1/n} $
    Use the `math` package for standard operations and `math/big` for handling potentially large volumes.

-   **Add this comment:**
    ```go
    // gaussianHeuristic computes the predicted length of the shortest non-zero vector
    // in a lattice of a given rank and volume, based on the Gaussian Heuristic formula.
    ```

### 1.5. The SVP Oracle (via `cgo` and `fplll`)

-   **Instruction:** This is the most complex step. Create a function `svpOracle(basis [][]*big.Int, radius float64)`. This function will use `cgo` to call `fplll`.
    1.  Add the `cgo` preamble to the top of your `lab1.go` file:
        ```go
        /*
        #cgo LDFLAGS: -lfplll
        #include <fplll.h>
        */
        import "C"
        ```
    2.  Inside the function, convert your Go `[][]*big.Int` basis into a C-style `fplll_int_matrix`.
    3.  Call `fplll`'s LLL reduction function (`lll_reduction`).
    4.  Call `fplll`'s enumeration function (`enumeration`) with the LLL-reduced basis and the given `radius`.
    5.  The function should return the squared norm of the shortest vector found.

-   **Add this comment:**
    ```go
    // svpOracle finds the shortest non-zero vector in the lattice within a given radius.
    // It acts as a wrapper around the fplll C library, accessed via cgo.
    // The process involves:
    // 1. Converting the Go integer basis into fplll's integer matrix format.
    // 2. Running LLL reduction to preprocess the basis, which is essential for enumeration.
    // 3. Running the enumeration algorithm to find the vector with the smallest norm.
    // It returns the squared norm of the vector found.
    ```

### 1.6. Putting Lab 1 Together

-   **Instruction:** Create a function `runLab1Verification()` that orchestrates the test.
    1.  Loop `n` from 30 to 60, as in the example. Set `m = n`. Choose a prime `q`, e.g., 131.
    2.  Inside the loop, call `genBasis`, `latticeVolume`, and `gaussianHeuristic`.
    3.  Call `svpOracle` with a radius of `1.5 * gh_prediction`.
    4.  Calculate the relative error: `abs(sqrt(svp_norm) - gh_prediction) / sqrt(svp_norm)`.
    5.  Print the results for each `n` in a formatted table.

-   **Add this comment:**
    ```go
    // runLab1Verification orchestrates the primary experiment of Lab 1.
    // It iterates through various lattice dimensions, and for each dimension:
    // 1. Generates a q-ary lattice basis.
    // 2. Predicts the shortest vector norm using the Gaussian Heuristic.
    // 3. Finds the actual shortest vector norm using the SVP oracle.
    // 4. Prints the predicted norm, the actual norm, and the relative error.
    ```

## Part 2: Implementing Lab 2 - The Geometric Series Assumption

Create a new file `lab2.go`.

### 2.1. The BKZ Function (via `cgo` and `fplll`)

-   **Instruction:** Create a function `runBKZ(basis [][]*big.Int, beta int)`. This function will also use `cgo` to call `fplll`.
    1.  Add the same `cgo` preamble as in `lab1.go`.
    2.  The function will convert the Go basis into an `fplll_int_matrix`.
    3.  Call `fplll`'s BKZ reduction function (`bkz_reduction`).
    4.  After reduction, access the Gram-Schmidt norms (`r(i,i)`) from the internal GSO object. `fplll`'s API allows access to this.
    5.  Calculate the profile: `profile[i] = log2(sqrt(gso_norms[i]))`.
    6.  Return the profile as a slice of `float64`.

-   **Add this comment:**
    ```go
    // runBKZ performs BKZ reduction on a given basis using the specified block size beta.
    // This function is a wrapper around the fplll C library.
    // The process is:
    // 1. Convert the Go basis to an fplll matrix.
    // 2. Call the BKZ reduction algorithm.
    // 3. After reduction, extract the squared norms of the Gram-Schmidt vectors.
    // 4. Compute the final profile as the log base 2 of the norms (log(||b_i*||)).
    // The resulting profile is returned for later analysis/plotting.
    ```

### 2.2. Putting Lab 2 Together

-   **Instruction:** Create a function `runLab2Verification()`.
    1.  Set parameters, e.g., `rank = 30`, `beta = 20`.
    2.  Generate a random basis using a helper function (you can adapt `genBasis` or create a new `genRandomBasis` that just creates a square matrix with random entries).
    3.  Call `runBKZ` with this basis.
    4.  Print the resulting basis profile to the console.

-   **Add this comment:**
    ```go
    // runLab2Verification orchestrates the experiment for Lab 2.
    // It generates a random lattice basis, runs the powerful BKZ reduction algorithm
    // on it, and then prints the resulting basis profile. The linearity of this
    // profile in a plot is evidence for the Geometric Series Assumption.
    ```

## Part 3: Final Execution and Demonstration

Create a file `main.go`.

### 3.1. Main Function

-   **Instruction:** Write the `main` function to call the orchestrator functions from the other files.

-   **Add this comment:**
    ```go
    // main is the entry point of the program. It executes the verification
    // experiments for both Lab 1 and Lab 2 in sequence and prints the
    // results to standard output in a formatted log.
    ```

### 3.2. Generate Execution Log

-   **Instruction:** Run your Go program and capture the output. The final output in your README or documentation should look like this structured log.

-   **Example Execution Log:**
    ```text
    --- Running Lab 1: Verifying the Gaussian Heuristic ---
    Target q: 131. Iterating from n=30 to n=60...

    n    | GH Prediction | SVP Norm      | Relative Error
    ------------------------------------------------------
    30   | 362.15        | 358.92        | 0.89%
    32   | 364.40        | 368.01        | 0.99%
    34   | 366.55        | 361.75        | 1.31%
    ...  | ...           | ...           | ...
    58   | 380.05        | 385.44        | 1.41%
    60   | 381.51        | 379.99        | 0.40%

    Lab 1 finished.

    --- Running Lab 2: Verifying the Geometric Series Assumption ---
    Generating a random lattice of rank 30.
    Running BKZ reduction with block size beta = 20...
    BKZ finished.
    Basis Profile (log2 of Gram-Schmidt norms):
    [9.88, 9.65, 9.42, 9.21, 9.01, 8.80, 8.61, 8.40, 8.21, 8.03, 7.84, 7.66, 7.48, 7.30, 7.12, 6.94, 6.77, 6.59, 6.42, 6.25, 6.08, 5.91, 5.74, 5.57, 5.40, 5.23, 5.06, 4.89, 4.72, 4.55]

    Lab 2 finished. Plot this profile data to visually check for linearity.
    ```