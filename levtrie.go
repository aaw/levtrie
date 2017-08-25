// Package levtrie provides a Trie implementation that supports fast searches
// for words within a given edit distance of a query string. Edit distance
// bounds are maintained during the search by simulating an NFA that accepts
// all words within distance d of the query string in parallel with the Trie
// traversal.
//
// An example NFA is pictured below for d = 2 and the word "edit":
//
//      ┌──┐   e  ┌──┐   d  ┌──┐   i  ┌──┐   t  ╔══╗
//      |  |─────▷|  |─────▷|  |─────▷|  |─────▷║  ║
//      └──┘     ◹└──┘     ◹└──┘     ◹└──┘     ◹╚══╝
//       △      ╱  △      ╱  △      ╱  △      ╱  △
//       │  ε,*╱   │  ε,*╱   │  ε,*╱   │  ε,*╱   │
//       │    ╱    │    ╱    │    ╱    │    ╱    │
//      *│   ╱    *│   ╱    *│   ╱    *│   ╱    *│
//       │  ╱      │  ╱      │  ╱      │  ╱      │
//      ┌──┐   e  ┌──┐   d  ┌──┐   i  ┌──┐   t  ╔══╗
//      |  |─────▷|  |─────▷|  |─────▷|  |─────▷║  ║
//      └──┘     ◹└──┘     ◹└──┘     ◹└──┘     ◹╚══╝
//       △      ╱  △      ╱  △      ╱  △      ╱  △
//       │  ε,*╱   │  ε,*╱   │  ε,*╱   │  ε,*╱   │
//       │    ╱    │    ╱    │    ╱    │    ╱    │
//      *│   ╱    *│   ╱    *│   ╱    *│   ╱    *│
//       │  ╱      │  ╱      │  ╱      │  ╱      │
//      ┌──┐   e  ┌──┐   d  ┌──┐   i  ┌──┐   t  ╔══╗
//   ──▷|  |─────▷|  |─────▷|  |─────▷|  |─────▷║  ║
//      └──┘      └──┘      └──┘      └──┘      ╚══╝
//
// The state on the bottom left is the initial state and the double-bordered
// states on the far right are accepting states. *-transitions can be taken on
// any character and ε-transitions can be taken without consuming a character.
// Each transition in the NFA above corresponds to an edit operation:
// transitions to the right represent no edit, transitions up represent
// insertions, diagonal ε-transitions represent deletions and diagonal
// *-transitions represent substitutions.
//
// The NFA above is a specific example from a parameterized family of
// Levenshtein NFAs. Creating other Levenshtein NFAs for different words or
// edit distances is straightfoward: add or remove columns as needed to fit
// the length of the word, add rows as needed to increase the edit distance,
// and adjust the labeled horizontal transitions to match the word.
//
// Simulating the NFA involves keeping track of an active set of states. The
// active set of states is defined by the lowest active state on each
// diagonal and only diagonals within a sliding window of 2d + 1 possible
// diagonals ever contain active states during the simulation. These properties
// make it possible to simulate each transition of the NFA in time O(d).
package levtrie

import (
	"unicode/utf8"
)

// Trie supports common map operations as well as lookups within a given edit
// distance bound. Don't create directly, use levtrie.New() instead.
type Trie struct {
	root *node
}

// KV is a key-value pair, the basic storage unit of the Trie.
type KV struct {
	Key   string
	Value string
}

// node is a Trie node.
type node struct {
	child map[rune]*node
	data  *KV
}

// New returns a new Trie.
func New() *Trie {
	return &Trie{root: &node{child: make(map[rune]*node)}}
}

// Get returns the value stored in the Trie at the given key. If there is no
// such key in the Trie, it returns the empty string. The second value returned
// is true exactly when the key exists in the Trie.
func (t *Trie) Get(key string) (string, bool) {
	n := t.root
	var ok bool
	var r rune
	for i, w := 0, 0; i < len(key); i += w {
		r, w = utf8.DecodeRuneInString(key[i:])
		if n, ok = n.child[r]; !ok {
			return "", false
		}
	}
	if n.data != nil {
		return n.data.Value, true
	}
	return "", false
}

// Set associates key with val in the Trie. A subsequent call to Get(key)
// will return (val, true).
func (t *Trie) Set(key string, val string) {
	n := t.root
	var r rune
	for i, w := 0, 0; i < len(key); i += w {
		r, w = utf8.DecodeRuneInString(key[i:])
		if x, ok := n.child[r]; !ok {
			z := &node{child: make(map[rune]*node)}
			n.child[r] = z
			n = z
		} else {
			n = x
		}

	}
	n.data = &KV{Key: key, Value: val}
}

// Delete removes the key from the Trie. A subsequent call to Get(key) will
// return ("", false).
func (t *Trie) Delete(key string) {
	n := t.root
	var ok bool
	// If the path through the Trie that we're trying to delete ends in a
	// leaf node, there will be a path of nodes starting from the last node
	// with more than one child between the root and the leaf and ending at
	// the leaf that should be cleaned up. We keep track of the root of that
	// path here with cnode/crune and prune it after the deletion.
	var cnode *node
	var r, crune rune
	for i, w := 0, 0; i < len(key); i += w {
		r, w = utf8.DecodeRuneInString(key[i:])
		if len(n.child) > 1 || cnode == nil {
			cnode, crune = n, r
		}
		if n, ok = n.child[r]; !ok {
			return
		}
	}
	n.data = nil
	if len(n.child) == 0 {
		delete(cnode.child, crune)
	}
}

// state is a state in the simulation of a Levenshtein NFA. This state
// corresponds to a set of states in the original NFA. Don't create one
// directly, use newState to create one instead.
type state struct {
	offset int
	arr    []int8
}

func newState(d int8, offset int) state {
	arr := make([]int8, 2*d+1)
	for i := range arr {
		arr[i] = int8(d + 1)
	}
	return state{offset: offset, arr: arr}
}

// nfa is a Levenshtein NFA.
type nfa struct {
	rs   []rune // The word this NFA matches, split into runes.
	d    int8   // The edit distance of the NFA.
	jump []int8 // Scratch space used by the transition method.
}

func newNfa(rs []rune, d int8) *nfa {
	return &nfa{rs: rs, d: d, jump: make([]int8, 3*int(d)+2)}
}

// start returns the start state of the nfa.
func (n nfa) start() state {
	initial := newState(n.d, int(-2*n.d))
	initial.arr[2*n.d] = 0
	return initial
}

// accepts returns true exactly when the NFA state passed is accepting.
func (n nfa) accepts(s state) bool {
	for i, x := range s.arr {
		dist := int8(len(n.rs) - s.offset - i)
		if dist <= n.d && dist >= x {
			return true
		}
	}
	return false
}

// transition computes the effect of a rune transition on a set of NFA states.
// Given a set of NFA states and a rune, it returns a new NFA state and the
// minimum edit distance among those states. The minimum edit distance is used
// to guide the Trie traversal in the direction of the matches with smallest
// edit distance.
func (n nfa) transition(s state, r rune) (state, int8) {
	ns := newState(n.d, s.offset+1)
	min := n.d + 1
	// Populate jump array, which lets us compute the horizontal transition
	// contribution in constant time below. jump stores information about
	// the position of r values within the string that's used by the next
	// for loop to figure out where active horizontal r-transitions on a
	// diagonal might occur.
	for i, next := len(n.jump)-1, n.d+1; i >= 0; i, next = i-1, next+1 {
		x := s.offset + i
		if x < len(n.rs) && x >= 0 && n.rs[x] == r {
			next = 0
		}
		n.jump[i] = next
	}
	for j := range ns.arr {
		val := n.d + 1
		// Compute horizontal transition contribution.
		cr := s.arr[j] + n.jump[j+int(s.arr[j])]
		if cr < val {
			val = cr
		}
		// Compute diagonal transition contribution.
		if j < len(s.arr)-1 && s.arr[j+1]+1 < val {
			val = s.arr[j+1] + 1
		}
		// Compute vertical transition contribution.
		if j < len(s.arr)-2 && s.arr[j+2]+1 < val {
			val = s.arr[j+2] + 1
		}
		if val < n.d+1 {
			ns.arr[j] = val
		}
		if val < min {
			min = val
		}
	}
	return ns, min
}

// frame is the complete state needed during a traversal of the Trie that's
// informed by a Levenshtein NFA: a node from the Trie plus a set of states in
// the NFA.
type frame struct {
	n node
	s state
}

// extractRunes converts a string to an array of runes.
func extractRunes(s string) []rune {
	rs := []rune{}
	i := 0
	var r rune
	for w := 0; i < len(s); i += w {
		r, w = utf8.DecodeRuneInString(s[i:])
		rs = append(rs, r)
	}
	return rs
}

// doNotExpandSuffixes is a strategy for searching a Trie that does not expand
// a node to explore suffixes of matches.
func doNotExpandSuffixes(n node, limit int) (results []KV, halt bool) {
	halt = false // Continue exploring this node from the traversal
	if n.data != nil {
		results = append(results, *n.data)
	}
	return
}

// expandSuffixes is a strategy for searching a Trie that adds all descendents
// of a node to the result set.
func expandSuffixes(n node, limit int) (results []KV, halt bool) {
	halt = true // Stop exploring this node from the traversal
	stack := []node{n}
	for len(stack) > 0 {
		var x node
		x, stack = stack[len(stack)-1], stack[:len(stack)-1]
		if x.data != nil {
			results = append(results, *x.data)
			if len(results) >= limit {
				break
			}
		}
		for _, child := range x.child {
			stack = append(stack, *child)
		}
	}
	return
}

// Suggest returns up to n KVs with keys that are within edit distance d of the
// input key. Example: Suggest("banana", 2, 10) would return up to 10 results
// which might include keys like "bahama", "bananas", or "panama".
func (t Trie) Suggest(key string, d int8, n int) []KV {
	return suggest(doNotExpandSuffixes, *t.root, extractRunes(key), d, n)
}

// SuggestSuffixes returns up to n KVs, all of whose keys have a prefix that
// is within edit distance d of the input key. Example:
// SuggestSuffixes("eat", 1, 10) would return up to 10 results which might
// include keys like "eaten", "eating", "beaten", and "meatball"
func (t Trie) SuggestSuffixes(key string, d int8, n int) []KV {
	return suggest(expandSuffixes, *t.root, extractRunes(key), d, n)
}

// SuggestAfterExactPrefix returns up to n KVs that share an exact prefix of
// length p with the input key and are within edit distance d of the input key.
// Example: SuggestAfterExactPrefix("britney", 3, 2, 10) would return up to 10
// results which might include "brine" and "briney" but not "jitney".
func (t Trie) SuggestAfterExactPrefix(key string, p int, d int8, n int) []KV {
	runes := extractRunes(key)
	var ok bool
	curr := t.root
	for _, r := range runes[:p] {
		if curr, ok = curr.child[r]; !ok {
			return nil
		}
	}
	return suggest(doNotExpandSuffixes, *curr, runes[p:], d, n)
}

// SuggestSuffixesAfterExactPrefix returns up to n KVs, all of whose keys have
// a prefix that is within edit distance d of the input key and share an exact
// prefix of at least length p with the input key. Example:
// SuggestSuffixesAfterExactPrefix("toads", 1, 2, 10) would return up to 10
// results which might include "toadstool" and "toast" but not "roads".
func (t Trie) SuggestSuffixesAfterExactPrefix(key string, p int, d int8, n int) []KV {
	runes := extractRunes(key)
	var ok bool
	curr := t.root
	for _, r := range runes[:p] {
		if curr, ok = curr.child[r]; !ok {
			return nil
		}
	}
	return suggest(expandSuffixes, *curr, runes[p:], d, n)
}

type processAcceptingNode func(n node, limit int) ([]KV, bool)

// suggest runs the traversal of the Trie, using frames consisting of a Trie
// state and a set of NFA nodes to store state. These frames are pushed on a
// stack and explored using the strategy defined by the process parameter to
// decide whether to halt or keep exploring suffixes after a match is found.
//
// Each state in the NFA corresponds to an edit distance. The edit distance of a
// state can't decrease when a transition occurs in the NFA and similarly,
// during the simulation of the NFA, the minimum edit distance of a set of NFA
// states can't decrease after a transition. While traversing the Trie, it makes
// sense to always explore the frame whose set of NFA states has the smallest
// minimum edit distance. We accomplish this below using a 2-dimensional stack
// where stack[i] stores all of the frames we haven't explored that have edit
// distance i. Once all frames have been popped and explored from stack[i], new
// frames will only be pushed to stack[i+1] or greater so we never need to
// backtrack through stack indexes.
func suggest(process processAcceptingNode, root node, runes []rune, d int8, limit int) []KV {
	n := newNfa(runes, d)
	start := n.start()
	stacks := make([][]frame, d+1)
	stacks[0] = []frame{frame{n: root, s: start}}
	var results []KV
	for i := range stacks {
		for len(stacks[i]) > 0 {
			var f frame
			// Pop the top frame from stacks[i]
			f, stacks[i] = stacks[i][len(stacks[i])-1], stacks[i][:len(stacks[i])-1]
			if n.accepts(f.s) {
				rs, halt := process(f.n, limit-len(results))
				results = append(results, rs...)
				if len(results) >= limit {
					return results[:limit]
				}
				if halt {
					continue
				}
			}
			// Register each of the current Trie node's children
			// for a traversal.
			for r, node := range f.n.child {
				if ns, min := n.transition(f.s, r); min < d+1 {
					stacks[min] = append(stacks[min], frame{n: *node, s: ns})
				}
			}
		}
	}
	return results
}
