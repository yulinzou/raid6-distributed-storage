package main

import (
    "errors"
    "fmt"
)

type Raid6 struct {
    Nodes []*Node
	Math *Math
}

func InitRaid6() *Raid6 {
    raid := &Raid6{
        Nodes: make([]*Node, 8), // 6 data nodes, 2 parity nodes
    }
    for i := 0; i < 8; i++ {
        raid.Nodes[i] = InitNode(i)
    }
    return raid
}

// WriteFile splits input data into blocks and writes them to data nodes.
// It also calculates parity blocks.
func (r *Raid6) WriteFile(fileName string, data []byte) error {
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

// calculateParity generates the parity blocks based on the data nodes.
func (r *Raid6) calculateParity(fileName string) {
    parity1 := make([]byte, len(r.Nodes[0].BlockList[0].Data))
    parity2 := make([]byte, len(r.Nodes[0].BlockList[0].Data))

    for i := 0; i < len(r.Nodes[0].BlockList[0].Data); i++ {
        // calculate parity by math algorithms
        // parity1[i] = r.Math ...
        // parity2[i] = r.Math ...
    }

    // Store parity blocks
    r.Nodes[6].AddBlock(InitBlock(6, fileName, parity1, Parity))
    r.Nodes[7].AddBlock(InitBlock(7, fileName, parity2, Parity))
}

// CheckCorruption verifies if the data blocks or parity are corrupted.
func (r *Raid6) CheckCorruption() bool {
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

// Need to edit ...
func (r *Raid6) RecoverData() error {
    // XOR operation across other blocks to recover corrupted data
    if r.CheckCorruption() {
        fmt.Println("Corruption detected, recovering data...")
        // Placeholder logic for recovery, as real implementation will be complex
        return nil
    } else {
        return errors.New("no corruption detected")
    }
}
