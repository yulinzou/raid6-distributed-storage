package raid6

type RAIDMath struct {
	generator int
	gfExp     [512]int
	gfLog     [256]int
	fieldSize int
}

// Initialize Galois Field with a given generator
func NewRAIDMath(generator int) *RAIDMath {
	math := &RAIDMath{
		generator: generator,
		fieldSize: 255, // GF(2^8) uses 255 field size
	}

	math.initGaloisField()
	return math
}

// Initialize Galois Field lookup tables for GF(2^8)
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

// Galois Field addition (XOR for GF(2^8))
func (rm *RAIDMath) GfAdd(a, b int) int {
	return a ^ b
}

// Galois Field multiplication
func (rm *RAIDMath) GfMul(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return rm.gfExp[rm.gfLog[a]+rm.gfLog[b]]
}

// Galois Field division
func (rm *RAIDMath) GfDiv(a, b int) int {
	if b == 0 {
		panic("Division by zero in Galois Field")
	}
	if a == 0 {
		return 0
	}
	return rm.gfExp[(rm.gfLog[a]-rm.gfLog[b]+255)%255]
}

// Galois Field exponentiation
func (rm *RAIDMath) GfExp(power int) int {
	// Ensure the power is within the valid range (mod 255 since GF(2^8) has 255 elements)
	return rm.gfExp[(power+255)%255] // Ensures non-negative index
}

// Galois Field inverse
func (rm *RAIDMath) GfInverse(a int) int {
	return rm.gfExp[255-rm.gfLog[a]]
}

// Calculate P and Q parities for the data blocks
func (rm *RAIDMath) CalculateParity(dataBlocks [][]byte) (pParity []byte, qParity []byte) {
	blockSize := len(dataBlocks[0])
	pParity = make([]byte, blockSize)
	qParity = make([]byte, blockSize)

	for i := 0; i < blockSize; i++ {
		p := 0
		q := 0
		for j := 0; j < len(dataBlocks); j++ {
			p = rm.GfAdd(p, int(dataBlocks[j][i]))
			q = rm.GfAdd(q, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i]))) // Q uses GF multiplication with generator
		}
		pParity[i] = byte(p)
		qParity[i] = byte(q)
	}

	return pParity, qParity
}

// ==== SINGLE BLOCK FAILURE RECOVERY ====

// Recover a single lost block using P parity
func (rm *RAIDMath) RecoverSingleBlockP(dataBlocks [][]byte, pParity []byte, missingIndex int) {
	blockSize := len(pParity)
	if dataBlocks[missingIndex] == nil {
		dataBlocks[missingIndex] = make([]byte, blockSize)
	}

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
}

// Recover a single lost block using Q parity
func (rm *RAIDMath) RecoverSingleBlockQ(dataBlocks [][]byte, qParity []byte, missingIndex int) {
	blockSize := len(qParity)

	// Initialize the missing block if necessary
	if dataBlocks[missingIndex] == nil {
		dataBlocks[missingIndex] = make([]byte, blockSize)
	}
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
}

// Recover P parity
func (rm *RAIDMath) RecoverPParity(dataBlocks [][]byte, pParity []byte) {
	blockSize := len(dataBlocks[0])
	if pParity == nil {
		pParity = make([]byte, blockSize)
	}

	for i := 0; i < blockSize; i++ {
		p := 0
		for j := 0; j < len(dataBlocks); j++ {
			p = rm.GfAdd(p, int(dataBlocks[j][i]))
		}

		// Update the original pParity slice in-place
		pParity[i] = byte(p)
	}
	// fmt.Println(pParity)
}

// Recover Q parity
func (rm *RAIDMath) RecoverQParity(dataBlocks [][]byte, qParity []byte) {
	blockSize := len(dataBlocks[0])
	if qParity == nil {
		qParity = make([]byte, blockSize)
	}

	for i := 0; i < blockSize; i++ {
		q := 0
		for j := 0; j < len(dataBlocks); j++ {
			q = rm.GfAdd(q, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i]))) // Q uses GF multiplication with generator
		}

		// Update the original qParity slice in-place
		qParity[i] = byte(q)
	}
}

// ==== MULTIPLE BLOCK FAILURE RECOVERY ====

// Recover two lost blocks using P and Q parities
func (rm *RAIDMath) RecoverTwoDataBlocks(dataBlocks [][]byte, pParity, qParity []byte, missingIndex1, missingIndex2 int) {
	blockSize := len(pParity)

	// Initialize missing blocks if necessary
	if dataBlocks[missingIndex1] == nil {
		dataBlocks[missingIndex1] = make([]byte, blockSize)
	}
	if dataBlocks[missingIndex2] == nil {
		dataBlocks[missingIndex2] = make([]byte, blockSize)
	}

	for i := 0; i < blockSize; i++ {
		// Get P and Q parities for the current byte
		p := int(pParity[i])
		q := int(qParity[i])

		// Calculate the sum of known data blocks for P and Q
		for j := 0; j < len(dataBlocks); j++ {
			if j != missingIndex1 && j != missingIndex2 {
				p = rm.GfAdd(p, int(dataBlocks[j][i]))
				q = rm.GfAdd(q, rm.GfMul(rm.GfExp(j), int(dataBlocks[j][i])))
			}
		}

		// Now, we need to solve the system for the two unknown blocks
		// Let D1 be the block at missingIndex1 and D2 be the block at missingIndex2

		x := rm.GfExp(missingIndex1) // g^missingIndex1
		y := rm.GfExp(missingIndex2) // g^missingIndex2

		// Solve for D2 first
		// D2 = (Q - g^missingIndex1 * P) / (g^missingIndex2 - g^missingIndex1)
		diff := rm.GfAdd(y, x) // g^missingIndex2 - g^missingIndex1
		d2 := rm.GfDiv(rm.GfAdd(q, rm.GfMul(x, p)), diff)

		// Solve for D1
		// D1 = P XOR D2
		d1 := rm.GfAdd(p, d2)

		// Write the recovered blocks
		dataBlocks[missingIndex1][i] = byte(d1)
		dataBlocks[missingIndex2][i] = byte(d2)
	}
}

// Recover single lost block and P parity
func (rm *RAIDMath) RecoverDataBlockAndPParity(dataBlocks [][]byte, pParity, qParity []byte, missingDataIndex int) {
	blockSize := len(qParity)
	if pParity == nil {
		pParity = make([]byte, blockSize)
	}
	if dataBlocks[missingDataIndex] == nil {
		dataBlocks[missingDataIndex] = make([]byte, blockSize)
	}

	rm.RecoverSingleBlockQ(dataBlocks, qParity, missingDataIndex)
	rm.RecoverPParity(dataBlocks, pParity)
}

// Recover single lost block and Q parity
func (rm *RAIDMath) RecoverDataBlockAndQParity(dataBlocks [][]byte, pParity, qParity []byte, missingDataIndex int) {
	blockSize := len(pParity)
	if qParity == nil {
		qParity = make([]byte, blockSize)
	}
	if dataBlocks[missingDataIndex] == nil {
		dataBlocks[missingDataIndex] = make([]byte, blockSize)
	}

	rm.RecoverSingleBlockP(dataBlocks, pParity, missingDataIndex)

	// Recalculate and update Q parity using the recovered data block
	rm.RecoverQParity(dataBlocks, qParity)
}

// Recover P and Q parities
func (rm *RAIDMath) RecoverPQParities(dataBlocks [][]byte, pParity, qParity []byte) {
	blockSize := len(dataBlocks[0])

	if pParity == nil && qParity == nil {
		pParity = make([]byte, blockSize)
		qParity = make([]byte, blockSize)
	}

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
}
