// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.23

// Package cursor augments [inspector.Inspector] with [Cursor]
// functionality allowing more flexibility and control during
// inspection.
//
// This package is a temporary private extension of inspector until
// proposal #70859 is accepted, and which point it will be moved into
// the inspector package, and [Root] will become a method of
// Inspector.
package cursor

import (
	"iter"
	"reflect"
	"slices"

	"github.com/tgo-lang/lang/ast"
	"github.com/tgo-lang/lang/token"

	"github.com/tgo-lang/tools/go/ast/inspector"
)

// A Cursor represents an [ast.Node]. It is immutable.
//
// Two Cursors compare equal if they represent the same node.
//
// Call [Root] to obtain a valid cursor.
type Cursor struct {
	in    *inspector.Inspector
	index int32 // index of push node; -1 for virtual root node
}

// Root returns a cursor for the virtual root node,
// whose children are the files provided to [New].
//
// Its [Cursor.Node] and [Cursor.Stack] methods return nil.
func Root(in *inspector.Inspector) Cursor {
	return Cursor{in, -1}
}

// Node returns the node at the current cursor position,
// or nil for the cursor returned by [Inspector.Root].
func (c Cursor) Node() ast.Node {
	if c.index < 0 {
		return nil
	}
	return c.events()[c.index].node
}

// String returns information about the cursor's node, if any.
func (c Cursor) String() string {
	if c.in == nil {
		return "(invalid)"
	}
	if c.index < 0 {
		return "(root)"
	}
	return reflect.TypeOf(c.Node()).String()
}

// indices return the [start, end) half-open interval of event indices.
func (c Cursor) indices() (int32, int32) {
	if c.index < 0 {
		return 0, int32(len(c.events())) // root: all events
	} else {
		return c.index, c.events()[c.index].index + 1 // just one subtree
	}
}

// Preorder returns an iterator over the nodes of the subtree
// represented by c in depth-first order. Each node in the sequence is
// represented by a Cursor that allows access to the Node, but may
// also be used to start a new traversal, or to obtain the stack of
// nodes enclosing the cursor.
//
// The traversal sequence is determined by [ast.Inspect]. The types
// argument, if non-empty, enables type-based filtering of events. The
// function f if is called only for nodes whose type matches an
// element of the types slice.
//
// If you need control over descent into subtrees,
// or need both pre- and post-order notifications, use [Cursor.Inspect]
func (c Cursor) Preorder(types ...ast.Node) iter.Seq[Cursor] {
	mask := maskOf(types)

	return func(yield func(Cursor) bool) {
		events := c.events()

		for i, limit := c.indices(); i < limit; {
			ev := events[i]
			if ev.index > i { // push?
				if ev.typ&mask != 0 && !yield(Cursor{c.in, i}) {
					break
				}
				pop := ev.index
				if events[pop].typ&mask == 0 {
					// Subtree does not contain types: skip.
					i = pop + 1
					continue
				}
			}
			i++
		}
	}
}

// Inspect visits the nodes of the subtree represented by c in
// depth-first order. It calls f(n, true) for each node n before it
// visits n's children. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of the node, followed
// by a call of f(n, false).
//
// Each node is represented by a Cursor that allows access to the
// Node, but may also be used to start a new traversal, or to obtain
// the stack of nodes enclosing the cursor.
//
// The complete traversal sequence is determined by [ast.Inspect].
// The types argument, if non-empty, enables type-based filtering of
// events. The function f if is called only for nodes whose type
// matches an element of the types slice.
func (c Cursor) Inspect(types []ast.Node, f func(c Cursor, push bool) (descend bool)) {
	mask := maskOf(types)
	events := c.events()
	for i, limit := c.indices(); i < limit; {
		ev := events[i]
		if ev.index > i {
			// push
			pop := ev.index
			if ev.typ&mask != 0 && !f(Cursor{c.in, i}, true) {
				i = pop + 1 // past the pop
				continue
			}
			if events[pop].typ&mask == 0 {
				// Subtree does not contain types: skip to pop.
				i = pop
				continue
			}
		} else {
			// pop
			push := ev.index
			if events[push].typ&mask != 0 {
				f(Cursor{c.in, push}, false)
			}
		}
		i++
	}
}

// Stack returns the stack of enclosing nodes, outermost first:
// from the [ast.File] down to the current cursor's node.
//
// To amortize allocation, it appends to the provided slice, which
// must be empty.
//
// Stack must not be called on the Root node.
//
// TODO(adonovan): perhaps this should be replaced by:
//
//	func (Cursor) Ancestors(filter []ast.Node) iter.Seq[Cursor]
//
// returning a filtering iterator up the parent chain.
// This finesses the question of allocation entirely.
func (c Cursor) Stack(stack []Cursor) []Cursor {
	if len(stack) > 0 {
		panic("stack is non-empty")
	}
	if c.index < 0 {
		panic("Cursor.Stack called on Root node")
	}

	events := c.events()
	for i := c.index; i >= 0; i = events[i].parent {
		stack = append(stack, Cursor{c.in, i})
	}
	slices.Reverse(stack)
	return stack
}

// Parent returns the parent of the current node.
//
// Parent must not be called on the Root node (whose [Cursor.Node] returns nil).
func (c Cursor) Parent() Cursor {
	if c.index < 0 {
		panic("Cursor.Parent called on Root node")
	}

	return Cursor{c.in, c.events()[c.index].parent}
}

// NextSibling returns the cursor for the next sibling node in the
// same list (for example, of files, decls, specs, statements, fields,
// or expressions) as the current node. It returns zero if the node is
// the last node in the list, or is not part of a list.
//
// NextSibling must not be called on the Root node.
func (c Cursor) NextSibling() (Cursor, bool) {
	if c.index < 0 {
		panic("Cursor.NextSibling called on Root node")
	}

	events := c.events()
	i := events[c.index].index + 1 // after corresponding pop
	if i < int32(len(events)) {
		if events[i].index > i { // push?
			return Cursor{c.in, i}, true
		}
	}
	return Cursor{}, false
}

// PrevSibling returns the cursor for the previous sibling node in the
// same list (for example, of files, decls, specs, statements, fields,
// or expressions) as the current node. It returns zero if the node is
// the first node in the list, or is not part of a list.
//
// It must not be called on the Root node.
func (c Cursor) PrevSibling() (Cursor, bool) {
	if c.index < 0 {
		panic("Cursor.PrevSibling called on Root node")
	}

	events := c.events()
	i := c.index - 1
	if i >= 0 {
		if j := events[i].index; j < i { // pop?
			return Cursor{c.in, j}, true
		}
	}
	return Cursor{}, false
}

// FirstChild returns the first direct child of the current node,
// or zero if it has no children.
func (c Cursor) FirstChild() (Cursor, bool) {
	events := c.events()
	i := c.index + 1                                   // i=0 if c is root
	if i < int32(len(events)) && events[i].index > i { // push?
		return Cursor{c.in, i}, true
	}
	return Cursor{}, false
}

// LastChild returns the last direct child of the current node,
// or zero if it has no children.
func (c Cursor) LastChild() (Cursor, bool) {
	events := c.events()
	if c.index < 0 { // root?
		if len(events) > 0 {
			// return push of final event (a pop)
			return Cursor{c.in, events[len(events)-1].index}, true
		}
	} else {
		j := events[c.index].index - 1 // before corresponding pop
		// Inv: j == c.index if c has no children
		//  or  j is last child's pop.
		if j > c.index { // c has children
			return Cursor{c.in, events[j].index}, true
		}
	}
	return Cursor{}, false
}

// Children returns an iterator over the direct children of the
// current node, if any.
func (c Cursor) Children() iter.Seq[Cursor] {
	return func(yield func(Cursor) bool) {
		c, ok := c.FirstChild()
		for ok && yield(c) {
			c, ok = c.NextSibling()
		}
	}
}

// FindNode returns the cursor for node n if it belongs to the subtree
// rooted at c. It returns zero if n is not found.
func (c Cursor) FindNode(n ast.Node) (Cursor, bool) {

	// FindNode is equivalent to this code,
	// but more convenient and 15-20% faster:
	if false {
		for candidate := range c.Preorder(n) {
			if candidate.Node() == n {
				return candidate, true
			}
		}
		return Cursor{}, false
	}

	// TODO(adonovan): opt: should we assume Node.Pos is accurate
	// and combine type-based filtering with position filtering
	// like FindPos?

	mask := maskOf([]ast.Node{n})
	events := c.events()

	for i, limit := c.indices(); i < limit; i++ {
		ev := events[i]
		if ev.index > i { // push?
			if ev.typ&mask != 0 && ev.node == n {
				return Cursor{c.in, i}, true
			}
			pop := ev.index
			if events[pop].typ&mask == 0 {
				// Subtree does not contain type of n: skip.
				i = pop
			}
		}
	}
	return Cursor{}, false
}

// FindPos returns the cursor for the innermost node n in the tree
// rooted at c such that n.Pos() <= start && end <= n.End().
// It returns zero if none is found.
// Precondition: start <= end.
//
// See also [astutil.PathEnclosingInterval], which
// tolerates adjoining whitespace.
func (c Cursor) FindPos(start, end token.Pos) (Cursor, bool) {
	if end < start {
		panic("end < start")
	}
	events := c.events()

	// This algorithm could be implemented using c.Inspect,
	// but it is about 2.5x slower.

	best := int32(-1) // push index of latest (=innermost) node containing range
	for i, limit := c.indices(); i < limit; i++ {
		ev := events[i]
		if ev.index > i { // push?
			if ev.node.Pos() > start {
				break // disjoint, after; stop
			}
			nodeEnd := ev.node.End()
			if end <= nodeEnd {
				// node fully contains target range
				best = i
			} else if nodeEnd < start {
				i = ev.index // disjoint, before; skip forward
			}
		}
	}
	if best >= 0 {
		return Cursor{c.in, best}, true
	}
	return Cursor{}, false
}
