package raid6

import (
	"bytes"
	"errors"

	// "fmt"

	// "fmt"

	"math/rand"
	"time"
)

type RAID6 struct {
	Nodes   []*Node
	Math    *RAIDMath
	FileNum int
	DiskNum int
}

func InitRAID6() *RAID6 {
	raid := &RAID6{
		DiskNum: 8,
		Nodes:   make([]*Node, 8), // 6 data nodes, 2 parity nodes
		Math:    NewRAIDMath(2),   // Generator 2 for GF(2^8)
		FileNum: 0,                // no file at the beginning
	}
	for i := 0; i < raid.DiskNum; i++ {
		raid.Nodes[i] = InitNode(i)
	}

	return raid
}

// WriteFile splits input data into blocks and writes them to data nodes.
// It also calculates parity blocks.
func (r *RAID6) WriteFile(fileName string, data []byte) error {
	if len(data) == 0 {
		return errors.New("file data is empty")
	}

	// Randomly select two indices for P and Q parity
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) // Seed the random number generator
	parityIndices := rnd.Perm(r.DiskNum)[:2]               // Randomly pick two unique indices from the range of FileNum

	// Assign P and Q parity blocks to the randomly selected indices
	pIndex := parityIndices[0]
	qIndex := parityIndices[1]
	// fmt.Println("P parity index:", pIndex, "Q parity index:", qIndex)

	// Number of data disks (excluding the parity disks)
	numDataBlocks := r.DiskNum - 2                               // 2 disks are for P and Q parity
	blockSize := (len(data) + numDataBlocks - 1) / numDataBlocks // Block size with rounding up for padding

	nodeID := 0
	// Write data blocks into the first (numDataBlocks) nodes
	for i := 0; i < numDataBlocks; i++ {
		for nodeID == pIndex || nodeID == qIndex {
			nodeID++
		}

		start := i * blockSize
		end := start + blockSize
		if end > len(data) {
			end = len(data) // Adjust if it's the last block and smaller than blockSize
		}

		// Allocate a new slice for the block data
		blockData := make([]byte, blockSize)
		if start > end {
			// Block already filled with zeros
			// Pass

		} else {
			copy(blockData, data[start:end]) // Copy data into the new block, with padding if necessary
		}

		// Initialize the block and assign it to the corresponding node
		block := InitBlock(i, fileName, &blockData, Normal, blockSize)
		r.Nodes[nodeID].AddBlock(block)
		nodeID++
	}
	r.Nodes[pIndex].AddBlock(InitBlock(-1, fileName, nil, pParity, blockSize))
	r.Nodes[qIndex].AddBlock(InitBlock(-2, fileName, nil, qParity, blockSize))

	fileID := r.FileNum
	r.calculateParity(fileID, blockSize, pIndex, qIndex)
	r.FileNum++

	return nil
}

// Read the file data from the RAID 6 by index
func (r *RAID6) ReadFile(Index int) ([]byte, error) {
	// Get the data blocks, P parity, and Q parity for the current file (index i)
	dataBlocks, _, _ := r.GetDataBlocks(Index)

	if !r.CheckStatus() {
		return nil, errors.New("node failure")
	}

	// Concatenate the data blocks to recover the original file data
	fileData := make([]byte, 0)

	for i := 0; i < len(dataBlocks); i++ {
		if dataBlocks[i] != nil {
			fileData = append(fileData, *dataBlocks[i]...)
		}
	}
	fileData = bytes.TrimRight(fileData, "\x00") // Remove padding

	return fileData, nil
}

// Check if all nodes are active
func (r *RAID6) CheckStatus() bool {
	for i := 0; i < r.DiskNum; i++ {
		if !r.Nodes[i].status {
			return false
		}
	}
	return true
}

// Get data blocks from nodes
func (r *RAID6) GetDataBlocks(Index int) (dataBlocks []*[]byte, P *[]byte, Q *[]byte) {
	dataBlocks = make([]*[]byte, r.DiskNum-2)

	for i := 0; i < r.DiskNum; i++ {
		// Ensure the index exists in BlockList
		if r.Nodes[i].BlockList[Index] == nil || Index >= len(r.Nodes[i].BlockList) {
			continue // Skip if the index is out of bounds or the block is nil
		} else if r.Nodes[i].BlockList[Index].BlockID >= 0 {
			dataBlocks[r.Nodes[i].BlockList[Index].BlockID] = r.Nodes[i].BlockList[Index].Data
		} else if r.Nodes[i].BlockList[Index].BlockID == -1 {
			P = r.Nodes[i].BlockList[Index].Data
		} else if r.Nodes[i].BlockList[Index].BlockID == -2 {
			Q = r.Nodes[i].BlockList[Index].Data
		}
		// fmt.Println("Node ", i, "BlockID ", r.Nodes[i].BlockList[Index].BlockID, "Data ", *r.Nodes[i].BlockList[Index].Data)

	}
	return dataBlocks, P, Q
}

// calculateParity generates the parity blocks based on the data nodes.
func (r *RAID6) calculateParity(fileId int, blockSize int, pIndex, qIndex int) {
	// should be revised to match the filename
	parity1 := make([]byte, blockSize)
	parity2 := make([]byte, blockSize)

	// fmt.Println("file number: ", fileId)
	dataBlocks, _, _ := r.GetDataBlocks(fileId)

	// Ensure dataBlocks is non-nil before proceeding
	if dataBlocks == nil {
		panic("dataBlocks is nil")
	}

	// Calculate P and Q parities
	r.Math.CalculateParity(dataBlocks, blockSize, &parity1, &parity2)

	// Store parity blocks
	r.Nodes[pIndex].BlockList[r.FileNum].Data = &parity1
	r.Nodes[qIndex].BlockList[r.FileNum].Data = &parity2
}

// CheckCorruption verifies if the data blocks or parity are corrupted.
// func (r *RAID6) CheckCorruption() bool {
// 	// XOR all blocks and if it results in 0, then no corruption
// 	blockData := make([]byte, len(r.Nodes[0].BlockList[0].Data))

// 	for i := 0; i < len(blockData); i++ {
// 		blockData[i] = r.Nodes[0].BlockList[0].Data[i]
// 		for j := 1; j < 6; j++ {
// 			blockData[i] ^= r.Nodes[j].BlockList[0].Data[i]
// 		}
// 		blockData[i] ^= r.Nodes[6].BlockList[0].Data[i] // XOR parity
// 		blockData[i] ^= r.Nodes[7].BlockList[0].Data[i] // XOR second parity
// 	}

// 	for _, b := range blockData {
// 		if b != 0 {
// 			return true // Corruption detected
// 		}
// 	}

// 	return false // No corruption
// }

// Simulate single node's failure
func (r *RAID6) NodeFailure(nodeID int) {
	r.Nodes[nodeID].GE()
}

// Simulate two nodes' failure
func (r *RAID6) TwoNodesFailure(nodeID1, nodeID2 int) {
	r.Nodes[nodeID1].GE()
	r.Nodes[nodeID2].GE()
}


// Recover single file with single node failure
func (r *RAID6) RecoverFile(nodeID int, fileID int) {
	dataBlocks, P, Q := r.GetDataBlocks(fileID)
	blockIndex := r.Nodes[nodeID].BlockList[fileID].BlockID

	if r.Nodes[nodeID].BlockList[fileID].BlockID >= 0 {
		r.Math.RecoverSingleBlockP(dataBlocks, P, blockIndex)
		r.Nodes[nodeID].BlockList[fileID].Data = dataBlocks[blockIndex]
	} else if r.Nodes[nodeID].BlockList[fileID].BlockID == -1 {
		r.Math.RecoverPParity(dataBlocks, P)
		r.Nodes[nodeID].BlockList[fileID].Data = P
	} else if r.Nodes[nodeID].BlockList[fileID].BlockID == -2 {
		r.Math.RecoverQParity(dataBlocks, Q)
		r.Nodes[nodeID].BlockList[fileID].Data = Q
	}
}

// My recovery function
func (r *RAID6) RecoverSingleNode(nodeID int) {
	// For the ith file
	for i := 0; i < r.FileNum; i++ {
		r.RecoverFile(nodeID, i)
	}
	r.Nodes[nodeID].status = true
}

// Assume that nodeID1 < nodeID2
func (r *RAID6) RecoverDoubleNodes(nodeID1, nodeID2 int) {
	// For the ith file
	for i := 0; i < r.FileNum; i++ {
		// Get the data blocks, P parity, and Q parity for the current file (index i)
		dataBlocks, P, Q := r.GetDataBlocks(i)

		// Block types of the failed blocks
		blockIndex1 := r.Nodes[nodeID1].BlockList[i].BlockID
		blockIndex2 := r.Nodes[nodeID2].BlockList[i].BlockID

		if blockIndex1 < blockIndex2 { // Swap nodeID1 and nodeID2 to keep blockIndex1 <= blockIndex2
			nodeID1, nodeID2 = nodeID2, nodeID1
			blockIndex1, blockIndex2 = blockIndex2, blockIndex1
		}

		// Recovery logic based on the block types of the two failed nodes
		if blockIndex1 >= 0 && blockIndex2 >= 0 {
			// Both blocks are normal data blocks, recover them using both P and Q parities
			r.Math.RecoverTwoDataBlocks(dataBlocks, P, Q, blockIndex1, blockIndex2)
			r.Nodes[nodeID1].BlockList[i].Data = dataBlocks[blockIndex1]
			r.Nodes[nodeID2].BlockList[i].Data = dataBlocks[blockIndex2]
		} else if blockIndex1 >= 0 && blockIndex2 == -1 {
			// Recover normal data block and recalculate P parity
			r.Math.RecoverSingleBlockQ(dataBlocks, Q, blockIndex1) // Recover normal data block using Q
			r.Math.RecoverPParity(dataBlocks, P)                   // Recalculate P parity
			r.Nodes[nodeID1].BlockList[i].Data = dataBlocks[blockIndex1]
			r.Nodes[nodeID2].BlockList[i].Data = P // Reassign the recalculated P parity
		} else if blockIndex1 >= 0 && blockIndex2 == -2 {
			// Recover normal data block and recalculate Q parity
			r.Math.RecoverSingleBlockP(dataBlocks, P, blockIndex1) // Recover normal data block using P
			r.Math.RecoverQParity(dataBlocks, Q)                   // Recalculate Q parity
			r.Nodes[nodeID1].BlockList[i].Data = dataBlocks[blockIndex1]
			r.Nodes[nodeID2].BlockList[i].Data = Q // Reassign the recalculated Q parity
		} else if blockIndex1 == -1 && blockIndex2 == -2 {
			// Both P and Q parities are missing, recalculate both
			r.Math.RecoverPQParities(dataBlocks, P, Q)
			r.Nodes[nodeID1].BlockList[i].Data = P // Reassign the recalculated P parity
			r.Nodes[nodeID2].BlockList[i].Data = Q // Reassign the recalculated Q parity
		}
	}
	r.Nodes[nodeID1].status = true
	r.Nodes[nodeID2].status = true
}

func (r *RAID6) UpdateData(fileName string, newData []byte) error {
	if len(newData) == 0 {
		return errors.New("file data is empty")
	}

	for i := 0; i < r.FileNum; i++ {
		if r.Nodes[0].BlockList[i].FileName == fileName {
			oldData, _, _ := r.GetDataBlocks(i)
			blockSize := r.Nodes[0].BlockList[i].Size
			for j := 0; j < len(oldData); j++ {
				start := j * blockSize
				end := start + blockSize
				if end > len(newData) {
					end = len(newData)
				}

				if !bytes.Equal(newData[start:end], *oldData[j]) {
					blockData := make([]byte, blockSize)
					copy(blockData, newData[start:end])
					r.Nodes[j].BlockList[i].Data = &blockData
				}
			}
			for pIndex := 0; pIndex < r.DiskNum; pIndex++ {
				if r.Nodes[pIndex].BlockList[i].BlockID == -1 {
					for qIndex := 0; qIndex < r.DiskNum; qIndex++ {
						if r.Nodes[pIndex].BlockList[i].BlockID == -2 {
							r.calculateParity(i, blockSize, pIndex, qIndex)
						}
					}
				}
			}
		}

	}
	return nil
}

// Need to edit ...
// func (r *RAID6) RecoverData() error {
// 	// XOR operation across other blocks to recover corrupted data
// 	if r.CheckCorruption() {
// 		fmt.Println("Corruption detected, recovering data...")
// 		// Placeholder logic for recovery, as real implementation will be complex
// 		return nil
// 	} else {
// 		return errors.New("no corruption detected")
// 	}
// }
