package raid6

import (
	"fmt"
)

type RAIDMath struct {
	generator int
	gfExp     [512]int
	gfLog     [256]int
	fieldSize int
}

// NewRAIDMath Initialize Galois Field with a given generator
func NewRAIDMath(generator int) *RAIDMath {
	math := &RAIDMath{
		generator: generator,
		fieldSize: 255, // GF(2^8) uses 255 field size
	}

	math.initGaloisField()
	return math
}

// initGaloisField Initialize Galois Field lookup tables for GF(2^8)
func (rm *RAIDMath) initGaloisField() {
	// Set the initial value for exponentiation table
	b := 1

	// Irreducible polynomial for GF(2^8) (x^8 + x^4 + x^3 + x^2 + 1)
	const irreducible = 0x11D

	// Build the exponentiation and logarithm tables
	for log := 0; log < 255; log++ {
		rm.gfLog[b] = log
		rm.gfExp[log] = b

		b <<= 1 // Multiply by 2 in GF(2^8)

		// If the result exceeds 8 bits, reduce using the irreducible polynomial
		if b >= 256 {
			b ^= irreducible // Reduction modulo the irreducible polynomial
		}
	}

	// Duplicate the exponentiation table to handle overflows
	for i := 255; i < 512; i++ {
		rm.gfExp[i] = rm.gfExp[i-255]
	}
}

// GfAdd Galois Field addition (XOR for GF(2^8))
func (rm *RAIDMath) GfAdd(a, b int) int {
	return a ^ b
}

// GfMul Galois Field multiplication
func (rm *RAIDMath) GfMul(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return rm.gfExp[rm.gfLog[a]+rm.gfLog[b]]
}

// GfDiv Galois Field division
func (rm *RAIDMath) GfDiv(a, b int) int {
	if b == 0 {
		panic("Division by zero in Galois Field")
	}
	if a == 0 {
		return 0
	}
	return rm.gfExp[(rm.gfLog[a]-rm.gfLog[b]+255)%255]
}

// GfExp Galois Field exponentiation
func (rm *RAIDMath) GfExp(power int) int {
	// Ensure the power is within the valid range (mod 255 since GF(2^8) has 255 elements)
	return rm.gfExp[(power+255)%255] // Ensures non-negative index
}

// GfInverse Galois Field inverse
func (rm *RAIDMath) GfInverse(a int) int {
	return rm.gfExp[255-rm.gfLog[a]]
}

// CalculateParity Calculate P and Q parities for the data blocks with pParity and qParity as *([]byte)
func (rm *RAIDMath) CalculateParity(dataBlocks [][]byte, blockSize int) ([]byte, []byte) {
	pParity := make([]byte, blockSize)
	qParity := make([]byte, blockSize)

	for i := 0; i < blockSize; i++ {
		p := 0
		q := 0
		for j := 0; j < len(dataBlocks); j++ {
			// Dereference the pointer to access the actual data block slice
			p = rm.GfAdd(p, int(dataBlocks[j][i]))
			q = rm.GfAdd(q, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i]))) // Q uses GF multiplication with generator
		}
		pParity[i] = byte(p)
		qParity[i] = byte(q)
	}

	return pParity, qParity
}

// identifyCorruptDataDisk Identify the corrupt data block using P* and Q*
func (rm *RAIDMath) identifyCorruptDataDisk(pStar, qStar int) int {
	// Compute the ratio Q* / P* to find the generator (g^z) of the corrupt disk
	ratio := rm.GfDiv(qStar, pStar)

	// Iterate over possible disks to match g^z
	for z := 0; z < 255; z++ { // 255 is the field size for GF(2^8)
		if ratio == rm.GfExp(z) {
			return z // z is the index of the corrupt data disk
		}
	}
	return -1 // If no match is found, return -1 indicating corruption could not be identified
}

// RepairCorruptedDataBlocks Identify corruption and perform the correct recovery operation
func (rm *RAIDMath) RepairCorruptedDataBlocks(dataBlocks [][]byte, pParity, qParity []byte) ([][]byte, []byte, []byte) {
	// Step 1: Recompute P* and Q* syndromes
	pStar, qStar := rm.recomputeSyndromes(dataBlocks, pParity, qParity)

	// Step 2: Identify the type of the corrupt disk
	if pStar != 0 && qStar == 0 {
		// Case 1: P parity is corrupt
		fmt.Println("P parity is corrupt, recovering...")
		pParity = rm.RecoverPParity(dataBlocks)
	} else if pStar == 0 && qStar != 0 {
		// Case 2: Q parity is corrupt
		fmt.Println("Q parity is corrupt, recovering...")
		qParity = rm.RecoverQParity(dataBlocks)
	} else if pStar != 0 {
		// Case 3: Data disk is corrupt
		fmt.Println("Data disk is corrupt, recovering...")
		corruptDisk := rm.identifyCorruptDataDisk(pStar, qStar)
		dataBlocks[corruptDisk] = rm.RecoverSingleBlockP(dataBlocks, pParity, corruptDisk)
	} else {
		fmt.Println("No corruption detected.")
	}

	return dataBlocks, pParity, qParity
}

// recomputeSyndromes Recompute P* and Q* syndromes
func (rm *RAIDMath) recomputeSyndromes(dataBlocks [][]byte, pParity, qParity []byte) (int, int) {
	pStar := 0
	qStar := 0

	// Recompute P* and Q* by summing the data blocks
	for i := 0; i < len(pParity); i++ {
		pStar = rm.GfAdd(pStar, int(pParity[i]))
		qStar = rm.GfAdd(qStar, int(qParity[i]))

		for j := 0; j < len(dataBlocks); j++ {
			if dataBlocks[j] != nil {
				pStar = rm.GfAdd(pStar, int(dataBlocks[j][i]))
				qStar = rm.GfAdd(qStar, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i])))
			}
		}
	}

	return pStar, qStar
}

// ==== SINGLE BLOCK FAILURE RECOVERY ====

// RecoverSingleBlockP Recover a single lost block using P parity
func (rm *RAIDMath) RecoverSingleBlockP(dataBlocks [][]byte, pParity []byte, missingIndex int) []byte {
	blockSize := len(pParity)
	dataBlocks[missingIndex] = make([]byte, blockSize)

	for i := 0; i < blockSize; i++ {
		p := int(pParity[i])

		// XOR all available blocks, excluding the missing one
		for j := 0; j < len(dataBlocks); j++ {
			if j != missingIndex {
				p = rm.GfAdd(p, int(dataBlocks[j][i]))
			}
		}

		// Write recovered block directly into the dataBlocks slice
		dataBlocks[missingIndex][i] = byte(p)
	}

	return dataBlocks[missingIndex]
}

// RecoverSingleBlockQ Recover a single lost block using Q parity
func (rm *RAIDMath) RecoverSingleBlockQ(dataBlocks [][]byte, qParity []byte, missingIndex int) []byte {
	blockSize := len(qParity)

	// Initialize the missing block if necessary
	dataBlocks[missingIndex] = make([]byte, blockSize)

	for i := 0; i < blockSize; i++ {
		q := int(qParity[i])

		// Subtract the contributions of all available blocks using Galois Field multiplication
		for j := 0; j < len(dataBlocks); j++ {
			if j != missingIndex {
				q = rm.GfAdd(q, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i])))
			}
		}

		// Recover the missing block by dividing by g^missingIndex (the generator raised to the missing block index)
		dataBlocks[missingIndex][i] = byte(rm.GfDiv(q, rm.GfExp(missingIndex)))
	}

	return dataBlocks[missingIndex]
}

// RecoverPParity Recover P parity with dataBlocks
func (rm *RAIDMath) RecoverPParity(dataBlocks [][]byte) (pParity []byte) {
	blockSize := len(dataBlocks[0])
	pParity = make([]byte, blockSize)

	for i := 0; i < blockSize; i++ {
		p := 0
		for j := 0; j < len(dataBlocks); j++ {
			p = rm.GfAdd(p, int(dataBlocks[j][i]))
		}

		pParity[i] = byte(p)
	}

	return pParity
}

// RecoverQParity Recover Q parity with dataBlocks
func (rm *RAIDMath) RecoverQParity(dataBlocks [][]byte) (qParity []byte) {
	blockSize := len(dataBlocks[0])
	qParity = make([]byte, blockSize)
	for i := 0; i < blockSize; i++ {
		q := 0
		for j := 0; j < len(dataBlocks); j++ {
			q = rm.GfAdd(q, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i]))) // Q uses GF multiplication with generator
		}

		qParity[i] = byte(q)
	}

	return qParity
}

// RecoverTwoDataBlocks Recover two lost blocks using P and Q parities with pParity and qParity
func (rm *RAIDMath) RecoverTwoDataBlocks(dataBlocks [][]byte, pParity, qParity []byte, missingIndex1, missingIndex2 int) ([]byte, []byte) {
	blockSize := len(pParity)

	// Initialize missing blocks if necessary
	dataBlocks[missingIndex1] = make([]byte, blockSize)
	dataBlocks[missingIndex2] = make([]byte, blockSize)

	for i := 0; i < blockSize; i++ {
		// Get P and Q parities for the current byte
		p := int((pParity)[i])
		q := int((qParity)[i])

		// Calculate the sum of known data blocks for P and Q
		for j := 0; j < len(dataBlocks); j++ {
			if j != missingIndex1 && j != missingIndex2 {
				p = rm.GfAdd(p, int(dataBlocks[j][i]))
				q = rm.GfAdd(q, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i])))
			}
		}

		// Solve for D2 first
		x := rm.GfExp(missingIndex1) // g^missingIndex1
		y := rm.GfExp(missingIndex2) // g^missingIndex2
		diff := rm.GfAdd(y, x)       // g^missingIndex2 - g^missingIndex1

		d2 := rm.GfDiv(rm.GfAdd(q, rm.GfMul(x, p)), diff)
		d1 := rm.GfAdd(p, d2)

		// Write the recovered blocks
		dataBlocks[missingIndex1][i] = byte(d1)
		dataBlocks[missingIndex2][i] = byte(d2)
	}
	return dataBlocks[missingIndex1], dataBlocks[missingIndex2]
}

// RecoverPQParities Recover P and Q parities with pParity and qParity as *([]byte)
func (rm *RAIDMath) RecoverPQParities(dataBlocks [][]byte) (pParity []byte, qParity []byte) {
	blockSize := len(dataBlocks[0])
	pParity = make([]byte, blockSize)
	qParity = make([]byte, blockSize)

	// Recalculate both P and Q parities from scratch
	for i := 0; i < blockSize; i++ {
		p := 0
		q := 0

		// Sum over all the data blocks to recalculate P and Q parities
		for j := 0; j < len(dataBlocks); j++ {
			p = rm.GfAdd(p, int(dataBlocks[j][i]))
			q = rm.GfAdd(q, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i])))
		}

		// Update the P and Q parities in-place
		pParity[i] = byte(p)
		qParity[i] = byte(q)
	}

	return pParity, qParity
}
