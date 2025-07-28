package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
)

type InputData struct {
	N      int
	K      int
	Shares []Share
}

type Share struct {
	X, Y *big.Int
}

func lagrangeInterpolateAtZero(points []Share) *big.Int {
	secret := new(big.Float).SetPrec(0)

	for i := 0; i < len(points); i++ {
		xi := new(big.Float).SetPrec(0).SetInt(points[i].X)
		yi := new(big.Float).SetPrec(0).SetInt(points[i].Y)
		term := new(big.Float).SetPrec(0).Set(yi)

		for j := 0; j < len(points); j++ {
			if i == j {
				continue
			}
			xj := new(big.Float).SetPrec(0).SetInt(points[j].X)
			numerator := new(big.Float).SetPrec(0).Neg(xj)
			denominator := new(big.Float).SetPrec(0).Sub(xi, xj)
			fraction := new(big.Float).SetPrec(0).Quo(numerator, denominator)
			term.Mul(term, fraction)
		}
		secret.Add(secret, term)
	}

	result, _ := secret.Int(nil)
	return result
}

func main() {
	testFiles := []string{"shares1.json", "shares2.json"}
	fmt.Println("--- Shamir's Secret Sharing Solver (Go Version) ---")

	for _, filename := range testFiles {
		fileBytes, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", filename, err)
			continue
		}

		var rawData map[string]interface{}
		if err := json.Unmarshal(fileBytes, &rawData); err != nil {
			fmt.Printf("Error parsing JSON from %s: %v\n", filename, err)
			continue
		}

		var input InputData
		for key, val := range rawData {
			if key == "keys" {
				keysMap := val.(map[string]interface{})
				input.N = int(keysMap["n"].(float64))
				input.K = int(keysMap["k"].(float64))
			} else {
				shareMap := val.(map[string]interface{})
				x, _ := new(big.Int).SetString(key, 10)
				base, _ := strconv.Atoi(shareMap["base"].(string))
				valueStr := shareMap["value"].(string)
				y, _ := new(big.Int).SetString(valueStr, base)
				input.Shares = append(input.Shares, Share{X: x, Y: y})
			}
		}

		sort.Slice(input.Shares, func(i, j int) bool {
			return input.Shares[i].X.Cmp(input.Shares[j].X) < 0
		})

		subsetForSolving := input.Shares[:input.K]
		secret := lagrangeInterpolateAtZero(subsetForSolving)
		fmt.Printf("Secret for %s: %s\n", filename, secret.String())
	}
}

// core logic from my notes :
// Parsing JSON data, where share values are encoded in various number bases.
// It uses the `math/big` package to handle arbitrarily large integers, preventing overflow.
// go has inbuilt functionality for it
// then we sort the parsed shares by their x coordinate so we get a consistent solution'
// Applying Lagrange Interpolation on the first 'k' shares to find the polynomial's constant term (f(0)), which is the secret
// Performing all interpolation calculations with unlimited-precision floating-point numbers (`big.Float`) to ensure mathematical accuracy and prevent rounding errors.
