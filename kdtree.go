// Copyright 2012 by Graeme Humphries <graeme@sudo.ca>
//
// kdtree is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// kdtree is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with kdtree.  If not, see http://www.gnu.org/licenses/.

// A K-Dimensional Tree library, based on an algorithmic description from:
// http://en.wikipedia.org/wiki/K-d_tree
// and implementation ideas from:
// http://hackage.haskell.org/package/KdTree
//
// Licensed under the LGPL Version 3: http://www.gnu.org/licenses/
package go4game

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"sync"
)

/***** Basic Tree Operations *****/

// Tree node, can be the parent for a subtree.
type Node struct {
	//Data *interface{}
	Data *GameObject

	// Axis for plane of bisection for this node, determined when added to a tree.
	axis        int
	Coordinates []float64
	tree        *Tree // Tree this node belongs to, avoids reverse scan for root node.
	parent      *Node // Parent == nil is a tree root.
	leftChild   *Node // Nodes < Location on this axis.
	rightChild  *Node // Nodes >= Location on this axis.
}

// Create a new node from a set of coordinates.
func NewNode(coords []float64) *Node {
	n := new(Node)
	n.Coordinates = coords

	return n
}

func String(list []float64) string {
	out := "("
	for i := 0; i < len(list); i++ {
		out += " " + strconv.FormatFloat(list[i], 'G', 5, 64)
	}
	out += " )"
	return out
}

// String representation of a node.
func (n *Node) String() string {
	out := "[ " + String(n.Coordinates) + ": axis = " + strconv.FormatInt(int64(n.axis), 10) + " ]"
	return out
}

// Adds new Node or subtree to existing (sub)tree. Returns an error if the Node can't be added to the tree.
func (n *Node) add(newnode *Node) error {
	// Check dimensions of new node at tree root.
	if n.parent == nil {
		if len(n.Coordinates) != len(newnode.Coordinates) {
			return errors.New("Node with " + string(len(newnode.Coordinates)) + " dimensions can't be added to tree with " + string(len(n.Coordinates)) + " dimensions.")
		}
	}

	// erase any existing parent to node being added
	if newnode.parent != nil {
		newnode.parent = nil
	}

	// re-add any children first
	if newnode.leftChild != nil {
		if err := n.add(newnode.leftChild); err != nil {
			return err
		}
		newnode.leftChild = nil
	}
	if newnode.rightChild != nil {
		if err := n.add(newnode.rightChild); err != nil {
			return err
		}
		newnode.rightChild = nil
	}

	// now place this node
	if newnode.Coordinates[n.axis] < n.Coordinates[n.axis] {
		if n.leftChild == nil {
			newnode.axis = (n.axis + 1) % len(n.Coordinates)
			n.leftChild = newnode
			newnode.parent = n
			return nil
		} else {
			n.leftChild.add(newnode)
		}
	} else {
		if n.rightChild == nil {
			newnode.axis = (n.axis + 1) % len(n.Coordinates)
			n.rightChild = newnode
			newnode.parent = n
			return nil
		} else {
			n.rightChild.add(newnode)
		}
	}

	return nil
}

// Removes node from the tree it belongs to, adjusting other nodes as necessary.
// If this operation creates a new tree root, it is returned, otherwise nil.
func (n *Node) remove() *Node {
	if n.parent != nil {
		if !(n.parent.leftChild == n || n.parent.rightChild == n) {
			panic(n.String() + " to be removed not attached to its parent: " + n.parent.String())
		}
		parent := n.parent
		// remove references to this node from the parent
		if parent.leftChild == n {
			parent.leftChild = nil
		}
		// avoiding "else" auto-corrects the potential error case where parent.leftChild == parent.rightChild == n
		if parent.rightChild == n {
			parent.rightChild = nil
		}
		// remove reference to parent
		n.parent = nil

		// re-add any children to the previous level
		if n.leftChild != nil {
			if err := parent.add(n.leftChild); err != nil {
				panic("Unexpected error while removing node: " + err.Error())
			}
		}
		if n.rightChild != nil {
			if err := parent.add(n.rightChild); err != nil {
				panic("Unexpected error while removing node: " + err.Error())
			}
		}

		// remove references to children
		n.leftChild = nil
		n.rightChild = nil
	} else { // tree root
		switch {
		// arbitrarily rebalance so n.rightChild is the new tree root
		case n.leftChild != nil && n.rightChild != nil:
			n.rightChild.parent = nil
			if err := n.rightChild.add(n.leftChild); err != nil {
				// should never be an error on internal tree ops
				panic("Unexpected error adding subtree to new root: " + err.Error())
			}
			return n.rightChild
		case n.leftChild != nil: // implied n.rightChild == nil
			n.leftChild.parent = nil // new tree root
			return n.leftChild
		case n.rightChild != nil: // implied n.leftChild == nil
			n.rightChild.parent = nil // new tree root
			return n.rightChild
		}
		// case: n.leftChild == nil && n.rightChild == nil means empty tree
		return nil
	}
	n.tree = nil
	return nil
}

// Performs a left depth first tree traversal, running function f on every Node found.
func (n *Node) traverse(f func(*Node)) {
	if n != nil {
		if n.leftChild != nil {
			n.leftChild.traverse(f)
		}
		if n.rightChild != nil {
			n.rightChild.traverse(f)
		}
		f(n)
	}
}

/***** Node list management functions *****/

// Returns a slice of all distinct nodes in the tree. This is done by a tree traversal,
// and will be equally slow.
func (t *Tree) NodeList() []*Node {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.nodeList()
}

// Returns a slice of all distinct nodes in the tree. This is done by a tree traversal,
// and will be equally slow.
func (n *Node) nodeList() []*Node {
	nodelist := make([]*Node, 0, 100)
	f := func(n *Node) {
		nodelist = append(nodelist, n)
	}
	n.traverse(f)

	return nodelist
}

// Wrapper for a slice of nodes implementing sort.Interface for different dimensional axes.
type sortableNodeList struct {
	// dimension axis to sort on
	Axis  int
	Nodes []*Node
}

func (snl *sortableNodeList) Len() int {
	return len(snl.Nodes)
}

func (snl *sortableNodeList) Less(i, j int) bool {
	return snl.Nodes[i].Coordinates[snl.Axis] < snl.Nodes[j].Coordinates[snl.Axis]
}

func (snl *sortableNodeList) Swap(i, j int) {
	snl.Nodes[i], snl.Nodes[j] = snl.Nodes[j], snl.Nodes[i]
}

// Perform the same search as Node.FindRange() on a list of nodes, used in
// unit testing. Axis is ignored in this function.
func (snl *sortableNodeList) findrange(ranges map[int]Range) ([]*Node, error) {
	result := make([]*Node, 0, len(snl.Nodes))
	for _, n := range snl.Nodes {
		add := true
		for a, r := range ranges {
			if a >= len(n.Coordinates) {
				return nil, errors.New("Range on axis " + string(a) + " exceeds tree dimensions.")
			}
			if a < 0 {
				return nil, errors.New("Negative axes are invalid.")
			}

			if n.Coordinates[a] < r.Min || n.Coordinates[a] > r.Max {
				add = false
			}
		}
		if add {
			result = append(result, n)
		}
	}

	return result, nil
}

// Find a Node in a slice of Nodes. Returns (slice index, true) if found, or (_, false) if missing.
func find_nl(nl []*Node, n1 *Node) (int, bool) {
	for i, n2 := range nl {
		if n1 == n2 {
			return i, true
		}
	}
	return 0, false
}

/***** Tree Search Functions *****/

// Searches Tree for node at exact coords. Returns (nil, nil) if no node matching coords found,
// or (nil, error) if len(coords) != tree dimensions.
func (t *Tree) Find(coords []float64) (*Node, error) {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.find(coords)
}

// Searches (sub)tree for node at exact coords. Returns (nil, nil) if no node matching coords found,
// or (nil, error) if len(coords) != tree dimensions.
func (n *Node) find(coords []float64) (*Node, error) {
	if len(coords) != len(n.Coordinates) {
		return nil, errors.New("Search coordinates have " + string(len(coords)) + " dimensions, tree has " + string(len(n.Coordinates)) + " dimensions.")
	}

	axis := n.axis
	if coords[axis] < n.Coordinates[axis] {
		if n.leftChild == nil {
			return nil, nil
		} else {
			return n.leftChild.find(coords)
		}
	} else if coords[axis] == n.Coordinates[axis] {
		if equal_fl(coords, n.Coordinates) {
			return n, nil
		}
	}
	// implicit else
	if n.rightChild == nil {
		return nil, nil
	}
	// implicit else
	return n.rightChild.find(coords)
}

// Finds the root of the tree from an arbitrary node.
func (n *Node) root() *Node {
	if n.parent == nil {
		return n
	}
	return n.parent.root()
}

// Range parameter, used to search the k-d tree.
type Range struct {
	Min float64
	Max float64
}

// Find a list of Nodes in Tree matching the supplied map of dimensional
// Ranges. The map index is used as the axis to restrict.
// Use math.Inf() to create remove the restriction on Min or Max.
//
// If no results are found, (nil, nil) is returned.
// If an axis outside of the tree's dimensions is specified, nil is returned with an error.
func (t *Tree) FindRange(ranges map[int]Range) ([]*Node, error) {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.findRange(ranges)
}

// Find a list of nodes matching the supplied map of dimensional
// Ranges. The map index is used as the axis to restrict.
// Use math.Inf() to create remove the restriction on Min or Max.
//
// If no results are found, (nil, nil) is returned.
// If an axis outside of the tree's dimensions is specified, nil is returned with an error.
func (n *Node) findRange(ranges map[int]Range) ([]*Node, error) {
	if n == nil {
		return nil, nil
	}

	result := make([]*Node, 0, 10)
	// check to see if the current node should be returned
	add := true
	for a, r := range ranges {
		if a >= len(n.Coordinates) {
			return nil, errors.New("Range on axis " + string(a) + " exceeds tree dimensions.")
		}
		if a < 0 {
			return nil, errors.New("Negative axes are invalid.")
		}

		if n.Coordinates[a] < r.Min || n.Coordinates[a] > r.Max {
			add = false
			break
		}
	}
	if add {
		result = append(result, n)
	}

	// search subtrees
	r, ok := ranges[n.axis]
	// search subtree if we're not restricting this axis, or if restrictions match.
	if !ok || r.Min < n.Coordinates[n.axis] {
		if left, err := n.leftChild.findRange(ranges); err == nil {
			result = append(result, left...)
		} else {
			return result, err
		}
	}
	if !ok || r.Max >= n.Coordinates[n.axis] {
		if right, err := n.rightChild.findRange(ranges); err == nil {
			result = append(result, right...)
		} else {
			return result, err
		}
	}

	return result, nil
}

// Tests equality of float slices, returns false if lengths or any values contained within differ.
func equal_fl(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

/***** Tree Object *****/
// Tree is needed for locking, to prevent syncronization issues.
type Tree struct {
	Mutex sync.RWMutex

	Root *Node
}

/***** Tree Functions *****/
// These functions wrap the private Node functions in lock operations so that
// they're thread-safe.

// Adds new Node to existing Tree. Returns an error if the Node can't be added to the tree.
func (t *Tree) Add(newnode *Node) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if t.Root == nil {
		t.Root = newnode
		return nil
	}
	return t.Root.add(newnode)
}

// Removes node from the Tree, rebalancoing other nodes as necessary.
// Returns an error if Node isn't a member of this Tree.
func (t *Tree) Remove(n *Node) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if n.tree != t {
		return errors.New("Node not a member of tree.")
	}
	if newroot := n.remove(); newroot != nil {
		t.Root = newroot
	}

	return nil
}

// Performs a left depth first tree traversal, running function f on every Node found.
func (t *Tree) Traverse(f func(*Node)) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	f(t.Root)
}

/***** Tree Management Functions *****/

// Builds a new tree from a list of nodes. This is destructive, and
// will remove any existing tree membership from nodes passed to it.
func BuildTree(nodes []*Node) *Tree {
	tree := new(Tree)
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	tree.Root = buildRootNode(nodes, 0, nil)
	f := func(n *Node) {
		n.tree = tree
	}
	tree.Root.traverse(f)

	return tree
}

// Builds a tree from a list of nodes. Returns the root Node of the new tree.
// This is destructive, and will break any existing tree these nodes may be a member of.
// This is intended to be used to build an new tree, or as part of a tree Balance.
// This is a recursive function, you should always call it with depth = 0, parent = nil.
func buildRootNode(nodes []*Node, depth int, parent *Node) *Node {
	var root *Node
	// special case handling first
	switch len(nodes) {
	case 0:
		root = nil
	case 1:
		dimensions := len(nodes[0].Coordinates)
		root = nodes[0]

		root.axis = depth % dimensions
		root.parent = parent
		root.leftChild = nil
		root.rightChild = nil
	default:
		median := (len(nodes) / 2) - 1 // -1 so that it's a slice index
		dimensions := len(nodes[0].Coordinates)

		snl := new(sortableNodeList)
		snl.Axis = depth % dimensions
		snl.Nodes = make([]*Node, len(nodes))
		copy(snl.Nodes, nodes)
		sort.Sort(snl)

		root = snl.Nodes[median]

		root.axis = snl.Axis
		root.parent = parent
		root.leftChild = buildRootNode(snl.Nodes[0:median], depth+1, root)
		root.rightChild = buildRootNode(snl.Nodes[median+1:], depth+1, root)
	}

	return root
}

// Rebalances a whole Tree.
func (t *Tree) Balance() {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	nodelist := t.Root.nodeList()
	t.Root = buildRootNode(nodelist, 0, nil)
}

// Checks that Tree is a valid kdtree. Returns an error of there are problems.
func (t *Tree) Validate() error {
	t.Mutex.RLock()
	t.Mutex.RUnlock()
	return t.Root.validate()
}

// Checks that the (sub)tree below this node is valid:
// - All children to the left of it are < it on the axis.
// - All children to the right of it are >= it on the axis.
// - All child axes are their parent's axis + 1 (mod # dimensions).
// - All children have the correct parent.
//
// Returns nil if valid, or an error describing something broken in the tree.
func (n *Node) validate() error {
	var err error = nil
	if n.leftChild != nil {
		f := func(check *Node) {
			if check.Coordinates[n.axis] >= n.Coordinates[n.axis] {
				err = errors.New(check.String() + " is right of " + n.String() + " on axis " + strconv.FormatInt(int64(n.axis), 10))
			}
		}
		n.leftChild.traverse(f)
		// check all subtrees / dimensions
		if err != nil {
			return err
		}
		// make sure axes are sensible
		if expected := (n.axis + 1) % len(n.Coordinates); n.leftChild.axis != expected {
			return errors.New("Child axis " + strconv.FormatInt(int64(n.leftChild.axis), 10) + " isn't parent axis + 1 (" + strconv.FormatInt(int64(expected), 10))
		}
		// make sure parental relationships are correct
		if n.leftChild.parent == nil {
			return errors.New("Child " + n.leftChild.String() + " is missing parent " + n.String())
		} else if n.leftChild.parent != n {
			return errors.New("Child " + n.leftChild.String() + " is has incorrect parent " + n.leftChild.parent.String() + ", should be " + n.String())
		}

		// finally check all subtrees
		if err = n.leftChild.validate(); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	if n.rightChild != nil {
		f := func(check *Node) {
			if check.Coordinates[n.axis] < n.Coordinates[n.axis] {
				err = errors.New(check.String() + " is left of " + n.String() + " on axis " + strconv.FormatInt(int64(n.axis), 10))
			}
		}
		n.rightChild.traverse(f)
		// check all subtrees / dimensions
		if err != nil {
			return err
		}
		// make sure axes are sensible
		if expected := (n.axis + 1) % len(n.Coordinates); n.rightChild.axis != expected {
			return errors.New("Child axis " + strconv.FormatInt(int64(n.rightChild.axis), 10) + " isn't parent axis + 1 (" + strconv.FormatInt(int64(expected), 10))
		}
		// make sure parental relationships are correct
		if n.rightChild.parent == nil {
			return errors.New("Child " + n.rightChild.String() + " is missing parent " + n.String())
		} else if n.rightChild.parent != n {
			return errors.New("Child " + n.rightChild.String() + " is has incorrect parent " + n.rightChild.parent.String() + ", should be " + n.String())
		}

		// finally check all subtrees
		if err = n.rightChild.validate(); err != nil {
			return err
		}
	}
	return err
}

// Returns Depth of the deepest branch of this Tree.
func (t *Tree) Depth() int {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.depth()
}

// Returns depth of the deepest branch of this (sub)tree.
func (n *Node) depth() int {
	if n == nil {
		return 0
	}
	left_depth := n.leftChild.depth() + 1
	right_depth := n.rightChild.depth() + 1
	if left_depth > right_depth {
		return left_depth
	}
	return right_depth
}

// Returns number of nodes in the Tree.
func (t *Tree) Size() int {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.size()
}

// Returns number of nodes in this (sub)tree.
func (n *Node) size() int {
	return len(n.nodeList())
}

func (t *Tree) IsCollision(target *GameObject) bool {
	if t == nil {
		return false
	}
	ranges := make(map[int]Range)
	for i, v := range target.posVector {
		ranges[i] = Range{
			Min: v - 20,
			Max: v + 20,
		}
	}
	results1, err := t.FindRange(ranges)
	if err != nil {
		log.Printf("kdtree error %v", err)
		return false
	}
	if results1 != nil {
		for _, v := range results1 {
			if v.Data.IsCollision(target) {
				return true
			}
		}
	}
	return false
}

func (w *World) makeKDTree() (t *Tree) {
	nodes := make([]*Node, 0)
	for _, t := range w.Teams {
		for _, m := range t.GameObjs {
			n := NewNode(m.posVector[:])
			n.Data = m
			nodes = append(nodes, n)
		}
	}
	return BuildTree(nodes)
}
