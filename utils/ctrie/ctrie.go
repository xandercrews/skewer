/*
Package ctrie provides an implementation of the Ctrie data structure, which is
a concurrent, lock-free hash trie. This data structure was originally presented
in the paper Concurrent Tries with Efficient Non-Blocking Snapshots:

https://axel22.github.io/resources/docs/ctries-snapshot.pdf

Copyright 2015 Workiva, LLC
Modified by stephane martin

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License
*/
package ctrie

import (
	"errors"
	"hash"
	"hash/fnv"
	"io"
	"sync/atomic"
	"unsafe"

	"github.com/cheekybits/genny/generic"
)

// Data is the generic value type that genny will replace.
type Data generic.Type

const (
	// w controls the number of branches at a node (2^w branches).
	w = 5
	// exp2 is 2^w, which is the hashcode space.
	exp2 = 32
)

// HashFactory returns a new Hash32 used to hash keys.
type HashFactory func() hash.Hash32

func defaultHashFactory() hash.Hash32 {
	return fnv.New32a()
}

// Ctrie is a concurrent, lock-free hash trie. By default, keys are hashed
// using FNV-1a unless a HashFactory is provided to New.
type Ctrie struct {
	root        *iNode
	readOnly    bool
	hashFactory HashFactory
}

// generation demarcates Ctrie snapshots. We use a heap-allocated reference
// instead of an integer to avoid integer overflows. Struct must have a field
// on it since two distinct zero-size variables may have the same address in
// memory.
type generation struct{ _ int }

func newGen() *generation {
	return new(generation)
}

// iNode is an indirection node. I-nodes remain present in the Ctrie even as
// nodes above and below change. Thread-safety is achieved in part by
// performing CAS operations on the I-node instead of the internal node array.
type iNode struct {
	main *mainNode
	gen  *generation

	// rdcss is set during an RDCSS operation. The I-node is actually a wrapper
	// around the descriptor in this case so that a single type is used during
	// CAS operations on the root.
	rdcss *rdcssDescriptor
}

// copyToGen returns a copy of this I-node copied to the given generation.
func (i *iNode) copyToGen(gen *generation, ctrie *Ctrie) *iNode {
	nin := &iNode{gen: gen}
	main := gcasRead(i, ctrie)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&nin.main)), unsafe.Pointer(main))
	return nin
}

// mainNode is either a cNode, tNode, lNode, or failed node which makes up an
// I-node.
type mainNode struct {
	cNode  *cNode
	tNode  *sNode
	lNode  *lNode
	failed *mainNode

	// prev is set as a failed main node when we attempt to CAS and the
	// I-node's generation does not match the root generation. This signals
	// that the GCAS failed and the I-node's main node must be set back to the
	// previous value.
	prev *mainNode
}

// cNode is an internal main node containing a bitmap and the array with
// references to branch nodes. A branch node is either another I-node or a
// singleton S-node.
type cNode struct {
	bmp   uint32
	array []branch
	gen   *generation
}

// newMainNode is a recursive constructor which creates a new mainNode. This
// mainNode will consist of cNodes as long as the hashcode chunks of the two
// keys are equal at the given level. If the level exceeds 2^w, an lNode is
// created.
func newMainNode(x *sNode, xhc uint32, y *sNode, yhc uint32, lev uint, gen *generation) *mainNode {
	if lev < exp2 {
		xidx := (xhc >> lev) & 0x1f
		yidx := (yhc >> lev) & 0x1f
		bmp := uint32((1 << xidx) | (1 << yidx))

		if xidx == yidx {
			// Recurse when indexes are equal.
			main := newMainNode(x, xhc, y, yhc, lev+w, gen)
			iNode := &iNode{main: main, gen: gen}
			return &mainNode{cNode: &cNode{bmp, []branch{branch{i: iNode}}, gen}}
		}
		if xidx < yidx {
			return &mainNode{cNode: &cNode{bmp, []branch{branch{s: x}, branch{s: y}}, gen}}
		}
		return &mainNode{cNode: &cNode{bmp, []branch{branch{s: y}, branch{s: x}}, gen}}
	}
	return &mainNode{lNode: makeLNode(x, y)}
}

// inserted returns a copy of this cNode with the new entry at the given
// position.
func (c *cNode) inserted(pos, flag uint32, br branch, gen *generation) *cNode {
	length := uint32(len(c.array))
	bmp := c.bmp
	array := make([]branch, length+1)
	copy(array, c.array)
	array[pos] = br
	for i, x := pos, uint32(0); x < length-pos; i++ {
		array[i+1] = c.array[i]
		x++
	}
	return &cNode{bmp: bmp | flag, array: array, gen: gen}
}

// updated returns a copy of this cNode with the entry at the given index
// updated.
func (c *cNode) updated(pos uint32, br branch, gen *generation) *cNode {
	array := make([]branch, len(c.array))
	copy(array, c.array)
	array[pos] = br
	return &cNode{bmp: c.bmp, array: array, gen: gen}
}

// removed returns a copy of this cNode with the entry at the given index
// removed.
func (c *cNode) removed(pos, flag uint32, gen *generation) *cNode {
	length := uint32(len(c.array))
	bmp := c.bmp
	array := make([]branch, length-1)
	for i := uint32(0); i < pos; i++ {
		array[i] = c.array[i]
	}
	for i, x := pos, uint32(0); x < length-pos-1; i++ {
		array[i] = c.array[i+1]
		x++
	}
	return &cNode{bmp: bmp ^ flag, array: array, gen: gen}
}

// renewed returns a copy of this cNode with the I-nodes below it copied to the
// given generation.
func (c *cNode) renewed(gen *generation, ctrie *Ctrie) *cNode {
	array := make([]branch, len(c.array))
	for i, br := range c.array {
		if br.i != nil {
			array[i] = branch{i: br.i.copyToGen(gen, ctrie)}
		} else {
			array[i] = br
		}
	}
	return &cNode{bmp: c.bmp, array: array, gen: gen}
}

// untombed returns the S-node contained by the T-node.
func untombed(t *sNode) *sNode {
	return &sNode{key: t.key, hash: t.hash, value: t.value}
}

type lNode struct {
	l *list
}

func makeLNode(x *sNode, y *sNode) *lNode {
	return &lNode{new(list).Add(x).Add(y)}
}

func (l *lNode) length() int {
	return l.l.Length()
}

func (l *lNode) head() *sNode {
	return l.l.Head()
}

// lookup returns the value at the given entry in the L-node or returns false
// if it's not contained.
func (l *lNode) lookup(key string) (Data, bool) {
	// return: value, true if key has been found
	found, ok := l.l.Find(func(sn *sNode) bool {
		return key == sn.key
	})
	if !ok {
		return nil, false
	}
	return found.value, true
}

// inserted creates a new L-node with the added entry.
func (l *lNode) inserted(key string, value Data, hash uint32, replace bool) (*lNode, bool, bool) {
	// return: newlist, replaced, inserted
	idx := l.l.FindIndex(func(sn *sNode) bool {
		return key == sn.key
	})
	if idx < 0 {
		// the key is not present in the list, so we can push the entry on top
		return &lNode{l.l.Add(&sNode{key: key, value: value, hash: hash})}, false, true
	}
	if !replace {
		// the key is already present, but she should not replace the value, so do nothing
		return l, false, false
	}
	// the key is already present in the list, so first we remove it from the list,
	// then we add the entry on top
	newl, _ := l.l.Remove(idx)
	return &lNode{newl.Add(&sNode{key: key, value: value, hash: hash})}, true, false
}

// removed creates a new L-node with the entry removed.
func (l *lNode) removed(key string) (*lNode, bool) {
	// return: value, true if key has been found
	idx := l.l.FindIndex(func(sn *sNode) bool {
		return key == sn.key
	})
	if idx < 0 {
		// the key was not found, so do nothing
		return l, false
	}
	nl, _ := l.l.Remove(idx)
	return &lNode{nl}, true
}

// branch is either an iNode or sNode.
type branch struct {
	i *iNode
	s *sNode
}

// Entry contains a Ctrie key-value pair.
type Entry struct {
	Key   string
	Value Data
}

// sNode is a singleton node which contains a single key and value.
type sNode struct {
	key   string
	value Data
	hash  uint32
}

// New creates an empty Ctrie which uses the provided HashFactory for key
// hashing. If nil is passed in, it will default to FNV-1a hashing.
func New(hashFactory HashFactory) *Ctrie {
	if hashFactory == nil {
		hashFactory = defaultHashFactory
	}
	root := &iNode{main: &mainNode{cNode: &cNode{}}}
	return newCtrie(root, hashFactory, false)
}

func newCtrie(root *iNode, hashFactory HashFactory, readOnly bool) *Ctrie {
	return &Ctrie{
		root:        root,
		hashFactory: hashFactory,
		readOnly:    readOnly,
	}
}

// Insert adds the key-value pair to the Ctrie
func (c *Ctrie) Insert(key string, value Data) (inserted bool) {
	c.assertReadWrite()
	return c.insert(key, value, c.hash(key))
}

func (c *Ctrie) Set(key string, value Data) (replaced bool) {
	c.assertReadWrite()
	return c.set(key, value, c.hash(key))
}

// Lookup returns the value for the associated key or returns false if the key
// doesn't exist.
func (c *Ctrie) Lookup(key string) (Data, bool) {
	return c.lookup(key, c.hash(key))
}

// Remove deletes the value for the associated key, returning true if it was
// removed or false if the entry doesn't exist.
func (c *Ctrie) Remove(key string) (Data, bool) {
	c.assertReadWrite()
	return c.remove(key, c.hash(key))
}

// Snapshot returns a stable, point-in-time snapshot of the Ctrie. If the Ctrie
// is read-only, the returned Ctrie will also be read-only.
func (c *Ctrie) Snapshot() *Ctrie {
	return c.snapshot(c.readOnly)
}

// ReadOnlySnapshot returns a stable, point-in-time snapshot of the Ctrie which
// is read-only. Write operations on a read-only snapshot will panic.
func (c *Ctrie) ReadOnlySnapshot() *Ctrie {
	return c.snapshot(true)
}

// snapshot wraps up the CAS logic to make a snapshot or a read-only snapshot.
func (c *Ctrie) snapshot(readOnly bool) *Ctrie {
	if readOnly && c.readOnly {
		return c
	}
	for {
		root := c.readRoot()
		main := gcasRead(root, c)
		if c.rdcssRoot(root, main, root.copyToGen(newGen(), c)) {
			if readOnly {
				// For a read-only snapshot, we can share the old generation
				// root.
				return newCtrie(root, c.hashFactory, readOnly)
			}
			// For a read-write snapshot, we need to take a copy of the root
			// in the new generation.
			return newCtrie(c.readRoot().copyToGen(newGen(), c), c.hashFactory, readOnly)
		}
	}
}

// Clear removes all keys from the Ctrie.
func (c *Ctrie) Clear() {
	for {
		root := c.readRoot()
		gen := newGen()
		newRoot := &iNode{
			main: &mainNode{cNode: &cNode{array: make([]branch, 0), gen: gen}},
			gen:  gen,
		}
		if c.rdcssRoot(root, gcasRead(root, c), newRoot) {
			return
		}
	}
}

// Iterator returns a channel which yields the Entries of the Ctrie.
func (c *Ctrie) Iterate(ch chan Entry) {
	snapshot := c.ReadOnlySnapshot()
	snapshot.traverse(snapshot.readRoot(), ch)
}

func (c *Ctrie) ForEach(f func(Entry)) {
	ch := make(chan Entry)
	go func() {
		c.Iterate(ch)
		close(ch)
	}()
	for e := range ch {
		f(e)
	}
}

func (c *Ctrie) Filter(f func(Entry) bool, ch chan Entry) {
	ich := make(chan Entry)
	go func() {
		c.Iterate(ich)
		close(ich)
	}()
	for e := range ich {
		if f(e) {
			ch <- e
		}
	}
}

// Size returns the number of keys in the Ctrie.
func (c *Ctrie) Size() int {
	// TODO: The size operation can be optimized further by caching the size
	// information in main nodes of a read-only Ctrie – this reduces the
	// amortized complexity of the size operation to O(1) because the size
	// computation is amortized across the update operations that occurred
	// since the last snapshot.
	snapshot := c.ReadOnlySnapshot()
	return snapshot.count(snapshot.readRoot())
}

func (c *Ctrie) count(i *iNode) int {
	main := gcasRead(i, c)
	if main.cNode != nil {
		var total int
		for _, br := range main.cNode.array {
			if br.i != nil {
				total += c.count(br.i)
			} else {
				total++
			}
		}
		return total
	}
	if main.lNode != nil {
		return main.lNode.length()
	}
	if main.tNode != nil {
		return 1
	}
	return 0
}

func (c *Ctrie) traverse(i *iNode, ch chan<- Entry) {
	main := gcasRead(i, c)
	if main.cNode != nil {
		for _, br := range main.cNode.array {
			if br.i != nil {
				c.traverse(br.i, ch)
			} else {
				ch <- Entry{Key: br.s.key, Value: br.s.value}
			}
		}
	}
	if main.lNode != nil {
		main.lNode.l.ForEach(
			func(sn *sNode) {
				ch <- Entry{Key: sn.key, Value: sn.value}
			},
		)
	}
	if main.tNode != nil {
		ch <- Entry{Key: main.tNode.key, Value: main.tNode.value}
	}
}

func (c *Ctrie) assertReadWrite() {
	if c.readOnly {
		panic("Cannot modify read-only snapshot")
	}
}

func (c *Ctrie) insert(key string, value Data, hash uint32) (inserted bool) {
	for {
		root := c.readRoot()
		done, _, inserted := c.iinsert(root, key, value, hash, false, 0, nil, root.gen)
		if done {
			return inserted
		}
	}
}

func (c *Ctrie) set(key string, value Data, hash uint32) (replaced bool) {
	for {
		root := c.readRoot()
		done, replaced, _ := c.iinsert(root, key, value, hash, true, 0, nil, root.gen)
		if done {
			return replaced
		}
	}
}

func (c *Ctrie) lookup(key string, hash uint32) (Data, bool) {
	for {
		root := c.readRoot()
		result, exists, ok := c.ilookup(root, key, hash, 0, nil, root.gen)
		if ok {
			return result, exists
		}
	}
}

func (c *Ctrie) remove(key string, hash uint32) (Data, bool) {
	for {
		root := c.readRoot()
		result, exists, ok := c.iremove(root, key, hash, 0, nil, root.gen)
		if ok {
			return result, exists
		}
	}
}

func (c *Ctrie) hash(k string) uint32 {
	hasher := c.hashFactory()
	io.WriteString(hasher, k)
	return hasher.Sum32()
}

// iinsert attempts to insert the entry into the Ctrie. If false is returned,
// the operation should be retried.
func (c *Ctrie) iinsert(i *iNode, key string, value Data, hash uint32, replace bool, lev uint, parent *iNode, startGen *generation) (bool, bool, bool) {
	// return: done, replaced, inserted
	main := gcasRead(i, c)
	if main.cNode != nil {
		cn := main.cNode
		flag, pos := flagPos(hash, lev, cn.bmp)
		if cn.bmp&flag == 0 {
			// If the relevant bit is not in the bitmap, then a copy of the
			// cNode with the new entry is created. The linearization point is
			// a successful CAS.
			rn := cn
			if cn.gen != i.gen {
				rn = cn.renewed(i.gen, c)
			}
			ncn := &mainNode{cNode: rn.inserted(pos, flag, branch{s: &sNode{key: key, value: value, hash: hash}}, i.gen)}
			if gcas(i, main, ncn, c) {
				return true, false, true
			}
			return false, false, false
		}
		// If the relevant bit is present in the bitmap, then its corresponding
		// branch is read from the array.
		bra := cn.array[pos]
		if bra.i != nil {
			// If the branch is an I-node, then iinsert is called recursively.
			if startGen == bra.i.gen {
				return c.iinsert(bra.i, key, value, hash, replace, lev+w, i, startGen)
			}
			if gcas(i, main, &mainNode{cNode: cn.renewed(startGen, c)}, c) {
				return c.iinsert(i, key, value, hash, replace, lev, parent, startGen)
			}
			return false, false, false
		}
		if bra.s.key != key {
			// If the branch is an S-node and its key is not equal to the
			// key being inserted, then the Ctrie has to be extended with
			// an additional level. The C-node is replaced with its updated
			// version, created using the updated function that adds a new
			// I-node at the respective position. The new Inode has its
			// main node pointing to a C-node with both keys. The
			// linearization point is a successful CAS.
			rn := cn
			if cn.gen != i.gen {
				rn = cn.renewed(i.gen, c)
			}
			nsn := &sNode{key: key, value: value, hash: hash}
			nin := &iNode{main: newMainNode(bra.s, bra.s.hash, nsn, nsn.hash, lev+w, i.gen), gen: i.gen}
			ncn := &mainNode{cNode: rn.updated(pos, branch{i: nin}, i.gen)}
			if gcas(i, main, ncn, c) {
				return true, false, true
			}
			return false, false, false
		}
		// If the key in the S-node is equal to the key being inserted,
		// then the C-node is replaced with its updated version with a new
		// S-node. The linearization point is a successful CAS.
		if replace {
			ncn := &mainNode{cNode: cn.updated(pos, branch{s: &sNode{key: key, value: value, hash: hash}}, i.gen)}
			if gcas(i, main, ncn, c) {
				return true, true, false
			}
			return false, false, false
		}
		// the key is already present, but we should not replace, so do nothing
		return true, false, false
	}
	if main.tNode != nil {
		clean(parent, lev-w, c)
		return false, false, false
	}
	if main.lNode != nil {
		newLNode, replaced, inserted := main.lNode.inserted(key, value, hash, replace)
		if !inserted && !replaced {
			// nothing to do
			return true, false, false
		}
		nln := &mainNode{lNode: newLNode}
		if gcas(i, main, nln, c) {
			return true, replaced, inserted
		}
		return false, false, false
	}
	panic("Ctrie is in an invalid state")
}

// ilookup attempts to fetch the entry from the Ctrie. The first two return
// values are the entry value and whether or not the entry was contained in the
// Ctrie. The last bool indicates if the operation succeeded. False means it
// should be retried.
func (c *Ctrie) ilookup(i *iNode, key string, hash uint32, lev uint, parent *iNode, startGen *generation) (Data, bool, bool) {
	// return Data, present, done
	main := gcasRead(i, c)
	switch {
	case main.cNode != nil:
		cn := main.cNode
		flag, pos := flagPos(hash, lev, cn.bmp)
		if cn.bmp&flag == 0 {
			// If the bitmap does not contain the relevant bit, a key with the
			// required hashcode prefix is not present in the trie.
			return nil, false, true
		}
		// Otherwise, the relevant branch at index pos is read from the array.
		bra := cn.array[pos]

		if bra.i != nil {
			// If the branch is an I-node, the ilookup procedure is called
			// recursively at the next level.
			if c.readOnly || startGen == bra.i.gen {
				return c.ilookup(bra.i, key, hash, lev+w, i, startGen)
			}
			if gcas(i, main, &mainNode{cNode: cn.renewed(startGen, c)}, c) {
				return c.ilookup(i, key, hash, lev, parent, startGen)
			}
			return nil, false, false
		}

		// If the branch is an S-node, then the key within the S-node is
		// compared with the key being searched – these two keys have the
		// same hashcode prefixes, but they need not be equal. If they are
		// equal, the corresponding value from the S-node is
		// returned and a NOTFOUND value otherwise.
		if bra.s.key == key {
			return bra.s.value, true, true
		}
		return nil, false, true
	case main.tNode != nil:
		return cleanReadOnly(main.tNode, lev, parent, c, key, hash)
	case main.lNode != nil:
		// Hash collisions are handled using L-nodes, which are essentially
		// persistent linked lists.
		val, ok := main.lNode.lookup(key)
		return val, ok, true
	default:
		panic("Ctrie is in an invalid state")
	}
}

// iremove attempts to remove the entry from the Ctrie. The first two return
// values are the entry value and whether or not the entry was contained in the
// Ctrie. The last bool indicates if the operation succeeded. False means it
// should be retried.
func (c *Ctrie) iremove(i *iNode, key string, hash uint32, lev uint, parent *iNode, startGen *generation) (Data, bool, bool) {
	// return value, removed, done
	main := gcasRead(i, c)
	if main.cNode != nil {
		cn := main.cNode
		flag, pos := flagPos(hash, lev, cn.bmp)
		if cn.bmp&flag == 0 {
			// If the bitmap does not contain the relevant bit, a key with the
			// required hashcode prefix is not present in the trie.
			return nil, false, true
		}
		// Otherwise, the relevant branch at index pos is read from the array.
		bra := cn.array[pos]

		if bra.i != nil {
			// If the branch is an I-node, the iremove procedure is called
			// recursively at the next level.
			if startGen == bra.i.gen {
				return c.iremove(bra.i, key, hash, lev+w, i, startGen)
			}
			if gcas(i, main, &mainNode{cNode: cn.renewed(startGen, c)}, c) {
				return c.iremove(i, key, hash, lev, parent, startGen)
			}
			return nil, false, false
		}

		// If the branch is an S-node, its key is compared against the key
		// being removed.
		if bra.s.key != key {
			// If the keys are not equal, the NOTFOUND value is returned.
			return nil, false, true
		}
		//  If the keys are equal, a copy of the current node without the
		//  S-node is created. The contraction of the copy is then created
		//  using the toContracted procedure. A successful CAS will
		//  substitute the old C-node with the copied C-node, thus removing
		//  the S-node with the given key from the trie – this is the
		//  linearization point
		ncn := cn.removed(pos, flag, i.gen)
		cntr := toContracted(ncn, lev)
		if gcas(i, main, cntr, c) {
			if parent != nil {
				main = gcasRead(i, c)
				if main.tNode != nil {
					cleanParent(parent, i, hash, lev-w, c, startGen)
				}
			}
			return bra.s.value, true, true
		}
		return nil, false, false
	}
	if main.tNode != nil {
		clean(parent, lev-w, c)
		return nil, false, false
	}
	if main.lNode != nil {
		newLNode, _ := main.lNode.removed(key)
		nln := &mainNode{lNode: newLNode}
		if nln.lNode.length() == 1 {
			nln = entomb(nln.lNode.head())
		}
		if gcas(i, main, nln, c) {
			val, ok := main.lNode.lookup(key)
			return val, ok, true
		}
		return nil, false, true
	}
	panic("Ctrie is in an invalid state")
}

// toContracted ensures that every I-node except the root points to a C-node
// with at least one branch. If a given C-Node has only a single S-node below
// it and is not at the root level, a T-node which wraps the S-node is
// returned.
func toContracted(cn *cNode, lev uint) *mainNode {
	if lev > 0 && len(cn.array) == 1 {
		bra := cn.array[0]
		if bra.s != nil {
			return entomb(bra.s)
		}
		return &mainNode{cNode: cn}
	}
	return &mainNode{cNode: cn}
}

// toCompressed compacts the C-node as a performance optimization.
func toCompressed(cn *cNode, lev uint) *mainNode {
	tmpArray := make([]branch, len(cn.array))
	for i, sub := range cn.array {
		if sub.i != nil {
			mainPtr := (*unsafe.Pointer)(unsafe.Pointer(&sub.i.main))
			main := (*mainNode)(atomic.LoadPointer(mainPtr))
			tmpArray[i] = resurrect(sub.i, main)
		} else {
			tmpArray[i] = sub
		}
	}

	return toContracted(&cNode{bmp: cn.bmp, array: tmpArray}, lev)
}

func entomb(m *sNode) *mainNode {
	return &mainNode{tNode: &sNode{key: m.key, value: m.value, hash: m.hash}}
}

func resurrect(iNode *iNode, main *mainNode) branch {
	if main.tNode != nil {
		return branch{s: untombed(main.tNode)}
	}
	return branch{i: iNode}
}

func clean(i *iNode, lev uint, ctrie *Ctrie) bool {
	main := gcasRead(i, ctrie)
	if main.cNode != nil {
		return gcas(i, main, toCompressed(main.cNode, lev), ctrie)
	}
	return true
}

func cleanReadOnly(tn *sNode, lev uint, p *iNode, ctrie *Ctrie, key string, hash uint32) (val Data, exists bool, ok bool) {
	if !ctrie.readOnly {
		clean(p, lev-5, ctrie)
		return nil, false, false
	}
	if tn.hash == hash && tn.key == key {
		return tn.value, true, true
	}
	return nil, false, true
}

func cleanParent(p, i *iNode, hc uint32, lev uint, ctrie *Ctrie, startGen *generation) {
	for {
		var (
			mainPtr  = (*unsafe.Pointer)(unsafe.Pointer(&i.main))
			main     = (*mainNode)(atomic.LoadPointer(mainPtr))
			pMainPtr = (*unsafe.Pointer)(unsafe.Pointer(&p.main))
			pMain    = (*mainNode)(atomic.LoadPointer(pMainPtr))
		)
		if pMain.cNode == nil {
			return
		}
		flag, pos := flagPos(hc, lev, pMain.cNode.bmp)
		if pMain.cNode.bmp&flag == 0 {
			return
		}
		sub := pMain.cNode.array[pos]
		if sub.i != i {
			return
		}
		if main.tNode == nil {
			return
		}
		ncn := pMain.cNode.updated(pos, resurrect(i, main), i.gen)
		if gcas(p, pMain, toContracted(ncn, lev), ctrie) {
			return
		}
		if ctrie.readRoot().gen != startGen {
			return
		}
	}
}

func flagPos(hashcode uint32, lev uint, bmp uint32) (uint32, uint32) {
	idx := (hashcode >> lev) & 0x1f
	flag := uint32(1) << uint32(idx)
	mask := uint32(flag - 1)
	pos := bitCount(bmp & mask)
	return flag, pos
}

func bitCount(x uint32) uint32 {
	x -= (x >> 1) & 0x55555555
	x = ((x >> 2) & 0x33333333) + (x & 0x33333333)
	x = ((x >> 4) + x) & 0x0f0f0f0f
	x *= 0x01010101
	return x >> 24
}

// gcas is a generation-compare-and-swap which has semantics similar to RDCSS,
// but it does not create the intermediate object except in the case of
// failures that occur due to the snapshot being taken. This ensures that the
// write occurs only if the Ctrie root generation has remained the same in
// addition to the I-node having the expected value.
func gcas(in *iNode, old, n *mainNode, ct *Ctrie) bool {
	prevPtr := (*unsafe.Pointer)(unsafe.Pointer(&n.prev))
	atomic.StorePointer(prevPtr, unsafe.Pointer(old))
	swapped := atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&in.main)),
		unsafe.Pointer(old),
		unsafe.Pointer(n),
	)
	if swapped {
		gcasComplete(in, n, ct)
		return atomic.LoadPointer(prevPtr) == nil
	}
	return false
}

// gcasRead performs a GCAS-linearizable read of the I-node's main node.
func gcasRead(in *iNode, ctrie *Ctrie) *mainNode {
	m := (*mainNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&in.main))))
	prev := (*mainNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&m.prev))))
	if prev == nil {
		return m
	}
	return gcasComplete(in, m, ctrie)
}

// gcasComplete commits the GCAS operation.
func gcasComplete(i *iNode, m *mainNode, ctrie *Ctrie) *mainNode {
	for {
		if m == nil {
			return nil
		}
		prev := (*mainNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&m.prev))))
		root := ctrie.rdcssReadRoot(true)
		if prev == nil {
			return m
		}

		if prev.failed != nil {
			// Signals GCAS failure. Swap old value back into I-node.
			fn := prev.failed
			swapped := atomic.CompareAndSwapPointer(
				(*unsafe.Pointer)(unsafe.Pointer(&i.main)),
				unsafe.Pointer(m),
				unsafe.Pointer(fn),
			)
			if swapped {
				return fn
			}
			m = (*mainNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&i.main))))
			continue
		}

		if root.gen == i.gen && !ctrie.readOnly {
			// Commit GCAS.
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&m.prev)), unsafe.Pointer(prev), nil) {
				return m
			}
			continue
		}

		// Generations did not match. Store failed node on prev to signal
		// I-node's main node must be set back to the previous value.
		atomic.CompareAndSwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&m.prev)),
			unsafe.Pointer(prev),
			unsafe.Pointer(&mainNode{failed: prev}),
		)
		m = (*mainNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&i.main))))
		return gcasComplete(i, m, ctrie)
	}
}

// rdcssDescriptor is an intermediate struct which communicates the intent to
// replace the value in an I-node and check that the root's generation has not
// changed before committing to the new value.
type rdcssDescriptor struct {
	old       *iNode
	expected  *mainNode
	nv        *iNode
	committed int32
}

// readRoot performs a linearizable read of the Ctrie root. This operation is
// prioritized so that if another thread performs a GCAS on the root, a
// deadlock does not occur.
func (c *Ctrie) readRoot() *iNode {
	return c.rdcssReadRoot(false)
}

// rdcssReadRoot performs a RDCSS-linearizable read of the Ctrie root with the
// given priority.
func (c *Ctrie) rdcssReadRoot(abort bool) *iNode {
	r := (*iNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.root))))
	if r.rdcss != nil {
		return c.rdcssComplete(abort)
	}
	return r
}

// rdcssRoot performs a RDCSS on the Ctrie root. This is used to create a
// snapshot of the Ctrie by copying the root I-node and setting it to a new
// generation.
func (c *Ctrie) rdcssRoot(old *iNode, expected *mainNode, nv *iNode) bool {
	desc := &iNode{
		rdcss: &rdcssDescriptor{
			old:      old,
			expected: expected,
			nv:       nv,
		},
	}
	if c.casRoot(old, desc) {
		c.rdcssComplete(false)
		return atomic.LoadInt32(&desc.rdcss.committed) == 1
	}
	return false
}

// rdcssComplete commits the RDCSS operation.
func (c *Ctrie) rdcssComplete(abort bool) *iNode {
	for {
		r := (*iNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.root))))
		if r.rdcss == nil {
			return r
		}

		var (
			desc = r.rdcss
			ov   = desc.old
			exp  = desc.expected
			nv   = desc.nv
		)

		if abort {
			if c.casRoot(r, ov) {
				return ov
			}
			continue
		}

		oldeMain := gcasRead(ov, c)
		if oldeMain == exp {
			// Commit the RDCSS.
			if c.casRoot(r, nv) {
				atomic.StoreInt32(&desc.committed, 1)
				return nv
			}
			continue
		}
		if c.casRoot(r, ov) {
			return ov
		}
		continue
	}
}

// casRoot performs a CAS on the Ctrie root.
func (c *Ctrie) casRoot(ov, nv *iNode) bool {
	c.assertReadWrite()
	return atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&c.root)),
		unsafe.Pointer(ov),
		unsafe.Pointer(nv),
	)
}

// ErrEmptyList is returned when an invalid operation is performed on an
// empty list.
var ErrEmptyList = errors.New("Empty list")

type list struct {
	head *sNode
	tail *list
}

// Head returns the head of the list. The bool will be false if the list is empty.
func (l *list) Head() *sNode {
	if l == nil {
		return nil
	}
	return l.head
}

// Tail returns the tail of the list. The bool will be false if the list is empty.
func (l *list) Tail() (*list, bool) {
	if l == nil {
		return nil, false
	}
	return l.tail, true
}

// IsEmpty indicates if the list is empty.
func (l *list) IsEmpty() bool {
	return l == nil
}

// Length returns the number of items in the list.
func (l *list) Length() (length int) {
	for {
		if l == nil {
			return length
		}
		length++
		l = l.tail
	}
}

// Add will add the item to the list, returning the new list.
func (l *list) Add(head *sNode) *list {
	if head == nil {
		return l
	}
	return &list{head: head, tail: l}
}

// Insert will insert the item at the given position, returning the new list or
// an error if the position is invalid.
func (l *list) Insert(val *sNode, pos int) (*list, error) {
	if val == nil {
		return l, nil
	}

	result := &list{}
	cur := result
	for {
		if pos == 0 {
			cur.head = val
			cur.tail = l
			return result, nil
		}
		if l == nil {
			return nil, ErrEmptyList
		}
		cur.head = l.head
		cur.tail = new(list)
		cur = cur.tail
		l = l.tail
		pos--
	}
}

// Get returns the item at the given position
func (l *list) Get(pos int) (*sNode, bool) {
	for {
		if l == nil {
			return nil, false
		}
		if pos == 0 {
			return l.head, true
		}
		l = l.tail
		pos--
	}
}

// Remove will remove the item at the given position, returning the new list or
// an error if the position is invalid.
func (l *list) Remove(pos int) (*list, error) {
	result := &list{}
	cur := result
	for {
		if l == nil {
			return nil, ErrEmptyList
		}
		if pos == 0 {
			cur.head = l.tail.head
			cur.tail = l.tail.tail
			return result, nil
		}
		cur.head = l.head
		cur.tail = new(list)
		cur = cur.tail
		l = l.tail
		pos--
	}
}

// Find applies the predicate function to the list and returns the first item
// which matches.
func (l *list) Find(pred func(*sNode) bool) (*sNode, bool) {
	for {
		if l == nil {
			return nil, false
		}
		if pred(l.head) {
			return l.head, true
		}
		l = l.tail
	}
}

// FindIndex applies the predicate function to the list and returns the index
// of the first item which matches or -1 if there is no match.
func (l *list) FindIndex(pred func(*sNode) bool) (idx int) {
	for {
		if l == nil {
			return -1
		}
		if pred(l.head) {
			return idx
		}
		l = l.tail
		idx++
	}
}

func (l *list) ForEach(f func(*sNode)) {
	for {
		if l == nil {
			return
		}
		f(l.head)
		l = l.tail
	}
}
