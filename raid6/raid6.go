package raid6

import (
	"errors"
	"fmt"
)

type RAID6 struct {
	Nodes []*Node
	Math  *RAIDMath
	FileNum int
}

func InitRAID6() *RAID6 {
	raid := &RAID6{
		Nodes: make([]*Node, 8), // 6 data nodes, 2 parity nodes
		Math: NewRAIDMath(2),   // Generator 2 for GF(2^8)
		FileNum: 0, // no file at the beginning
	}
	for i := 0; i < 8; i++ {
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

	blockSize := len(data) / 6 // Dividing data into 6 blocks
	for i := 0; i < 6; i++ {
		start := i * blockSize
		end := start + blockSize
		if i == 5 {
			end = len(data) // Handle last block that may have remaining data
		}

		block := InitBlock(i, fileName, data[start:end], Normal)
		r.Nodes[i].AddBlock(block)
	}

	// Calculate parity blocks for parity nodes (7th and 8th node)
	r.calculateParity(fileName)
	return nil
}


// Get data blocks from nodes 
func (r *RAID6) GetDataBlocks(Index int) (dataBlocks []*[]byte, P *[]byte, Q *[]byte){
	dataBlocks = make([]*[]byte, 6)
	j := 0
	for i := 0; i < 8; i++ {
		// fetch the no.Index file data blocks 
		if r.Nodes[i].BlockList[Index].Type == Normal {
			dataBlocks[j] = &r.Nodes[i].BlockList[0].Data
		} 
		if r.Nodes[i].BlockList[Index].Type == pParity {
			P = &r.Nodes[i].BlockList[0].Data
		}
		if r.Nodes[i].BlockList[Index].Type == qParity {
			Q = &r.Nodes[i].BlockList[0].Data
		}
		j++
		
	}
	return dataBlocks, P, Q
}


// calculateParity generates the parity blocks based on the data nodes.
func (r *RAID6) calculateParity(fileName string) {
	// parity1 := make([]byte, len(r.Nodes[0].BlockList[0].Data))
	// parity2 := make([]byte, len(r.Nodes[0].BlockList[0].Data))

	// should be revised to match the filename
	blockSize := r.Nodes[0].BlockList[0].Size
	dataBlocks, _, _ := r.GetDataBlocks(0)
	parity1, parity2 := r.Math.CalculateParity(dataBlocks, blockSize)

	// Store parity blocks
	r.Nodes[6].AddBlock(InitBlock(6, fileName, parity1, pParity))
	r.Nodes[7].AddBlock(InitBlock(7, fileName, parity2, qParity))
}

// CheckCorruption verifies if the data blocks or parity are corrupted.
func (r *RAID6) CheckCorruption() bool {
	// XOR all blocks and if it results in 0, then no corruption
	blockData := make([]byte, len(r.Nodes[0].BlockList[0].Data))

	for i := 0; i < len(blockData); i++ {
		blockData[i] = r.Nodes[0].BlockList[0].Data[i]
		for j := 1; j < 6; j++ {
			blockData[i] ^= r.Nodes[j].BlockList[0].Data[i]
		}
		blockData[i] ^= r.Nodes[6].BlockList[0].Data[i] // XOR parity
		blockData[i] ^= r.Nodes[7].BlockList[0].Data[i] // XOR second parity
	}

	for _, b := range blockData {
		if b != 0 {
			return true // Corruption detected
		}
	}

	return false // No corruption
}


// Simulate single node's failure
func (r *RAID6) NodeFailure(nodeID int) {
	r.Nodes[nodeID].GE()
}

// Simulate two nodes' failure
func (r *RAID6) TwoNodesFailure(nodeID1, nodeID2 int) {
	r.Nodes[nodeID1].GE()
	r.Nodes[nodeID2].GE()
}

// My recovery function
func (r *RAID6) RecoverSingleNode(nodeID int) {
	for i := 0; i < r.FileNum; i++ {
		// to be revised ...
		dataBlocks, P, Q := r.GetDataBlocks(i)
		if r.Nodes[nodeID].BlockList[i].Type == Normal {
			r.Math.RecoverSingleBlockP(dataBlocks, *P, nodeID)
		} else if r.Nodes[nodeID].BlockList[i].Type == pParity {
			r.Math.RecoverSingleBlockP(dataBlocks, *P, nodeID)
		} else if r.Nodes[nodeID].BlockList[i].Type == qParity {
			r.Math.RecoverSingleBlockQ(dataBlocks, *Q, nodeID)
		}
	}
}

// Need to edit ...
func (r *RAID6) RecoverData() error {
	// XOR operation across other blocks to recover corrupted data
	if r.CheckCorruption() {
		fmt.Println("Corruption detected, recovering data...")
		// Placeholder logic for recovery, as real implementation will be complex
		return nil
	} else {
		return errors.New("no corruption detected")
	}
}
