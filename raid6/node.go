package raid6

type BlockType int

const (
    Normal BlockType = iota
    Parity
)

type Block struct {
    Type     BlockType
    FileName string
    Data     []byte
    BlockID  int
    Size     int
}

type Node struct {
    NodeID    int
    BlockList []*Block
}

func InitBlock(blockID int, fileName string, data []byte, blockType BlockType) *Block {
    return &Block{
        Type:     blockType,
        FileName: fileName,
        Data:     data,
        BlockID:  blockID,
        Size:     len(data),
    }
}

func InitNode(nodeID int) *Node {
    return &Node{
        NodeID:    nodeID,
        BlockList: []*Block{},
    }
}

func (n *Node) AddBlock(block *Block) {
    n.BlockList = append(n.BlockList, block)
}
