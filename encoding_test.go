package iradix

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

var spewConfig = spew.ConfigState{
	DisablePointerAddresses: true,
	DisableCapacities:       true,
}

type nodeNoChannel struct {
	leaf   *leafNodeNoChannel
	prefix []byte
	edges  []edgeNoChannel
}

func stripNodeChannel(n *Node) *nodeNoChannel {
	if n == nil {
		return nil
	}
	return &nodeNoChannel{
		leaf:   stripLeafChannel(n.leaf),
		prefix: n.prefix,
		edges:  stripEdgesChannel(n.edges),
	}
}

type edgeNoChannel struct {
	label byte
	node  *nodeNoChannel
}

func stripEdgesChannel(edges edges) []edgeNoChannel {
	var newEdges []edgeNoChannel = nil
	if edges != nil {
		for _, edge := range edges {
			newEdges = append(newEdges, edgeNoChannel{
				label: edge.label,
				node:  stripNodeChannel(edge.node),
			})
		}
	}
	return newEdges
}

type leafNodeNoChannel struct {
	key []byte
	val interface{}
}

func stripLeafChannel(leaf *leafNode) *leafNodeNoChannel {
	if leaf == nil {
		return nil
	}
	return &leafNodeNoChannel{
		key: leaf.key,
		val: leaf.val,
	}
}

type treeNoChannel struct {
	root *nodeNoChannel
	size int
}

func stripTreeChannel(t *Tree) *treeNoChannel {
	return &treeNoChannel{
		root: stripNodeChannel(t.root),
		size: t.size,
	}
}

func TestEncodeTree(t *testing.T) {
	r := New()
	r, _, _ = r.Insert([]byte("foo"), nil)
	r, _, _ = r.Insert([]byte("bar"), nil)
	r, _, _ = r.Insert([]byte("foobar"), nil)

	bys := EncodeTree(nil, r)
	r1 := DecodeTree(nil, bys)

	assert.Equal(t, spewConfig.Sdump(stripTreeChannel(r)), spewConfig.Sdump(stripTreeChannel(r1)))
}
