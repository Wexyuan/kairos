package id

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

// Snowflake generates unique IDs.
type Snowflake struct {
	node *snowflake.Node // underlying snowflake node
}

func New(node int64) (*Snowflake, error) {
	if node < 0 || node > 1023 {
		return nil, fmt.Errorf("[id] node %d out of range: must be between 0 and 1023", node)
	}
	n, err := snowflake.NewNode(node)
	if err != nil {
		return nil, fmt.Errorf("[id] create snowflake node %d error: %w", node, err)
	}
	return &Snowflake{node: n}, nil
}

// NextID generates a new ID.
func (s *Snowflake) NextID() int64 {
	return s.node.Generate().Int64()
}
