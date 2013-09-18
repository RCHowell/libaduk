package libaduk

import (
    "log"
    "fmt"
)

const (
    SEQUENCE_START = '('
    SEQUENCE_END = ')'
    NODE_START = ';'
    PROPERTY_START = '['
    PROPERTY_END = ']'
)

type Node struct {
    Previous *Node // Parent Node
    Next *Node // Child Node
    Up *Node // Upper sibling Node
    Down *Node // Lower sibling Node
    sgfData string
    numChildren int
    level int
}

func NewNode(prev *Node) *Node {
    return &Node { prev, nil, nil, nil, "", 0, 0 }
}

// The cursor parses Sgf data and provides methods to traverse the game
type Cursor struct {
    tree *Node
}

// Create a new cursor struct for given sgf data
func NewCursor(sgf []byte) (*Cursor, error) {
    tree, err := parse(string(sgf))

    if err != nil {
        return nil, err
    }

    return &Cursor { tree }, nil
}

// Begin parse an sgf string
func parse(sgf string) (*Node, error) {
    log.Printf("Parsing: %s\n", sgf)

    tree := NewNode(nil)
    lastNode := tree
    sequenceNodes := make([]*Node, 0)
    nodeStartIndex := -1
    lastParsedType := SEQUENCE_END
    isInProperty := false

    // range on string handles unicode automatically
    for i, value := range sgf {

        // If value is not a control character, ignore it
        if !(value == SEQUENCE_START || value == SEQUENCE_END || value == PROPERTY_START ||
                value == PROPERTY_END || value == NODE_START) {
            continue
        }

        // When in property, continue until end of property
        if isInProperty {
            if value == PROPERTY_END {
                // Here we check if the end of proptery is really the end or just an escaped character
                numberOfEscapes := 0
                for j := i - 1; sgf[j] == '\\'; j-- {
                    numberOfEscapes++
                }

                // If number of escapes is even, the property really ends here
                if numberOfEscapes % 2  == 0 {
                    isInProperty = false
                }
            }
            continue
        }

        if value == PROPERTY_START {
            isInProperty = true
        }

        // Sequence starts
        if value == SEQUENCE_START {

            // Safe sgf string to current node before creating a new one
            if lastParsedType != SEQUENCE_END && nodeStartIndex != -1 {
                lastNode.sgfData = sgf[nodeStartIndex:i]
            }

            // Create new Node for Sequence
            node := NewNode(lastNode)

            // Has current node already a child, than node is a sibling of lastNode
            if lastNode.Next != nil {
                last := lastNode.Next

                for last.Down != nil {
                    last = last.Down
                }

                node.Up = last
                last.Down = node
                node.level = last.level + 1
            } else {
                lastNode.Next = node
            }

            // Update current to new sequence
            lastNode.numChildren++

            // Add sequence to stack
            sequenceNodes = append(sequenceNodes, lastNode)

            lastNode = node
            nodeStartIndex = -1
            lastParsedType = SEQUENCE_START
        }

        // Sequence ends
        if value == SEQUENCE_END {
            // Safe sgf string to current node before creating a new one
            if lastParsedType != SEQUENCE_END && nodeStartIndex != -1 {
                lastNode.sgfData = sgf[nodeStartIndex:i]
            }

            // If we had sequences in the stack, set current node to last in stack
            if len(sequenceNodes) > 0 {
                lastNode = sequenceNodes[len(sequenceNodes) - 1]
                sequenceNodes = sequenceNodes[:len(sequenceNodes) - 1]
            } else {
                // If there was no sequence start for this sequence end, the sgf is malformed
                return nil, fmt.Errorf("Malformed SGF (No Sequence start found for sequence end at position %d)!", i)
            }

            lastParsedType = SEQUENCE_END
        }

        // Node starts
        if value == NODE_START {
            if nodeStartIndex != -1 {
                // Safe sgf string to current node before creating a new one
                lastNode.sgfData = sgf[nodeStartIndex:i]

                // Create new node and update current
                node := NewNode(lastNode)
                lastNode.numChildren = 1
                lastNode.Next = node
                lastNode = node

            }

            nodeStartIndex = i
        }
    }

    // If we are in a property or sequence after parsing, the sgf is malformed
    if isInProperty || len(sequenceNodes) > 0 {
        return nil, fmt.Errorf("Malformed SGF (Still in Property or Sequence after parsing)!")
    }

    return tree, nil
}