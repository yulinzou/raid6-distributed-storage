package raid6

type BlockType int

const (
    Normal BlockType = iota
    pParity
	qParity
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
	status bool // true for active, false for inactive(failure)
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
		status: true,
    }
}

func (n *Node) AddBlock(block *Block) {
    n.BlockList = append(n.BlockList, block)
}

func (n *Node) GE(){ // 寄
	n.status = false
}
