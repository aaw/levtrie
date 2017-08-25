levtrie
=======

This go package implements a Trie that acts like a map from strings to strings,
allowing the usual `Get`, `Set`, and `Delete` operations on key-value
associations. In addition, the Trie supports a few variations on searches for
key-value pairs where the key is within a particular edit distance of a query
string.

Here's an example of how you might use `levtrie` to find the 10 closest words
to the word "similar" that are within edit distance 2 from a list of words
held in the slice `wordlist`:

```
import "github.com/aaw/levtrie"

...

t := levtrie.New()
for word := range wordlist {
    t.Set(word, "")
}
results := t.Suggest("similar", 2, 10)
```

See [the godoc for this package](https://godoc.org/github.com/aaw/levtrie)
for complete documentation.

See examples/typeahead in this repo for an extended example of implementing
typeahead-style query suggestions with `levtrie`. You can launch a small webapp
with `go run examples/typeahead/typeahead.go` from the top level clone of this
repo that lets you type in a query box to find suggestions for spelling
corrections.

All of the searches restricted by edit distance in `levtrie` are accomplished
by generating a non-deterministic Levenshtein Automata on the fly and simulating
it in parallel with the Trie search. I've described this technique in more
detail in [this post](http://blog.aaw.io/2017/08/25/levenshtein-automata.html).
