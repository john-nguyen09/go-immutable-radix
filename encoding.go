package iradix

import "encoding/binary"

func putUInt64(dst []byte, v uint64) []byte {
	// binary.BigEndian.PutUint64 assumes dst has enough space and it does bounds check
	// before putting the value in. However, putUInt64 always append to the end of the slice
	// therefore it is a waste to use binary.BigEndian.PutUint64
	return append(dst, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func readUInt64(src []byte) (uint64, []byte) {
	// binary.BigEndian.Uint64 is faster because it provides bounds check hints to compiler
	// before reading the slice
	return binary.BigEndian.Uint64(src), src[8:]
}

func putBool(dst []byte, v bool) []byte {
	by := byte(0)
	if v {
		by = byte(1)
	}
	return append(dst, by)
}

func readBool(src []byte) (bool, []byte) {
	by := src[0]
	return by == byte(1), src[1:]
}

func putBytes(dst []byte, b []byte) []byte {
	dst = putUInt64(dst, uint64(len(b)))
	return append(dst, b...)
}

func readBytes(src []byte) ([]byte, []byte) {
	l, src := readUInt64(src)
	bys := []byte(nil)
	if l > 0 {
		bys, src = src[:l], src[l:]
	}
	return bys, src
}

// ValueEncoder is the encoder to encode and decode value
type ValueEncoder interface {
	// Encode encodes the value into byte slice
	Encode([]byte, interface{}) []byte
	// Decode decodes byte slice into value
	Decode([]byte) (interface{}, []byte)
}

// EncodeTree converts the tree to byte slice
func EncodeTree(e ValueEncoder, tree *Tree) []byte {
	var dst []byte = nil
	if tree.root != nil {
		dst = putBool(dst, true)
		dst = encodeNode(e, dst, tree.root)
	} else {
		dst = putBool(dst, false)
	}
	dst = putUInt64(dst, uint64(tree.size))
	return dst
}

// DecodeTree converts byte slice to tree
func DecodeTree(e ValueEncoder, data []byte) *Tree {
	t := &Tree{}
	hasRoot, data := readBool(data)
	if hasRoot {
		t.root, data = decodeNode(e, data)
	}
	size, _ := readUInt64(data)
	t.size = int(size)
	return t
}

func encodeNode(e ValueEncoder, dst []byte, n *Node) []byte {
	if n.leaf != nil {
		dst = putBool(dst, true)
		dst = encodeLeafNode(e, dst, n.leaf)
	} else {
		dst = putBool(dst, false)
	}
	dst = putBytes(dst, n.prefix)
	dst = encodeEdges(e, dst, n.edges)
	return dst
}

func decodeNode(e ValueEncoder, data []byte) (*Node, []byte) {
	n := &Node{
		mutateCh: make(chan struct{}),
	}
	hasLeaf, data := readBool(data)
	if hasLeaf {
		n.leaf, data = decodeLeafNode(e, data)
	}
	n.prefix, data = readBytes(data)
	n.edges, data = decodeEdges(e, data)
	return n, data
}

func encodeLeafNode(e ValueEncoder, dst []byte, leaf *leafNode) []byte {
	dst = putBytes(dst, leaf.key)
	if e != nil {
		dst = putBool(dst, true)
		dst = e.Encode(dst, leaf.val)
	} else { // No encoder for value mark this as no value
		dst = putBool(dst, false)
	}
	return dst
}

func decodeLeafNode(e ValueEncoder, data []byte) (*leafNode, []byte) {
	leaf := &leafNode{
		mutateCh: make(chan struct{}),
	}
	leaf.key, data = readBytes(data)
	hasValue, data := readBool(data)
	if hasValue && e != nil {
		leaf.val, data = e.Decode(data)
	}
	return leaf, data
}

func encodeEdges(e ValueEncoder, dst []byte, edges edges) []byte {
	dst = putUInt64(dst, uint64(len(edges)))
	for _, edge := range edges {
		dst = encodeEdge(e, dst, edge)
	}
	return dst
}

func decodeEdges(e ValueEncoder, data []byte) (edges, []byte) {
	var edges edges = nil
	l, data := readUInt64(data)
	if l > 0 {
		var edge edge
		for i := 0; i < int(l); i++ {
			edge, data = decodeEdge(e, data)
			edges = append(edges, edge)
		}
	}
	return edges, data
}

func encodeEdge(e ValueEncoder, dst []byte, edge edge) []byte {
	dst = append(dst, edge.label)
	if edge.node != nil {
		dst = putBool(dst, true)
		dst = encodeNode(e, dst, edge.node)
	} else {
		dst = putBool(dst, false)
	}
	return dst
}

func decodeEdge(e ValueEncoder, data []byte) (edge, []byte) {
	edge, data := edge{
		label: data[0],
	}, data[1:]
	hasNode, data := readBool(data)
	if hasNode {
		edge.node, data = decodeNode(e, data)
	}
	return edge, data
}
