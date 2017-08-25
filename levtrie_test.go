package levtrie

import (
	"math/rand"
	"sort"
	"strings"
	"testing"
)

func TestExtractRunes(t *testing.T) {
	wants := []string{
		"",
		"a",
		"aь",
		"ь",
		"редактировать",
		"редакти",
		"ред",
	}
	for _, want := range wants {
		got := string(extractRunes(want))
		if got != want {
			t.Errorf("Want %v, got %v", want, got)
		}
	}
}

func expectGet(t *testing.T, r *Trie, key string, val string) {
	if actual, ok := r.Get(key); ok && actual != val {
		t.Errorf("Got val = '%v', ok = %v but want val == '%v', ok = true.",
			actual, ok, val)
	}
}

func expectNotGet(t *testing.T, r *Trie, key string) {
	if actual, ok := r.Get(key); ok {
		t.Errorf("Got val = %v, ok = %v but want !ok", actual, ok)
	}
}

func TestGetEmpty(t *testing.T) {
	r := New()
	if _, ok := r.Get("foo"); ok {
		t.Error("Got ok, want !ok.")
	}
}

func TestSetGet(t *testing.T) {
	r := New()
	r.Set("foo", "bar")
	expectGet(t, r, "foo", "bar")
}

func TestSetDelete(t *testing.T) {
	r := New()
	r.Set("foo", "bar")
	r.Delete("foo")
	expectNotGet(t, r, "foo")
}

func TestSetSetDeleteDelete(t *testing.T) {
	r := New()
	r.Set("foo", "bar")
	r.Set("bar", "foo")
	r.Delete("foo")
	expectNotGet(t, r, "foo")
	expectGet(t, r, "bar", "foo")
	r.Delete("bar")
	expectNotGet(t, r, "foo")
	expectNotGet(t, r, "bar")
}

func TestSetSetSetDeleteDeleteDelete(t *testing.T) {
	r := New()
	r.Set("foo", "bar")
	r.Set("bar", "foo")
	r.Set("baz", "biz")
	r.Delete("foo")
	expectNotGet(t, r, "foo")
	expectGet(t, r, "bar", "foo")
	expectGet(t, r, "baz", "biz")
	r.Delete("bar")
	expectNotGet(t, r, "foo")
	expectNotGet(t, r, "bar")
	expectGet(t, r, "baz", "biz")
	r.Delete("baz")
	expectNotGet(t, r, "foo")
	expectNotGet(t, r, "bar")
	expectNotGet(t, r, "baz")
}

func TestGetUnsuccessful(t *testing.T) {
	r := New()
	r.Set("fooey", "bara")
	r.Set("fooing", "barb")
	r.Set("foozle", "barc")
	expectGet(t, r, "fooey", "bara")
	expectGet(t, r, "fooing", "barb")
	expectGet(t, r, "foozle", "barc")
}

func TestDeleteUnsuccessful(t *testing.T) {
	r := New()
	r.Delete("foo")
	r.Set("fooey", "bara")
	r.Set("fooing", "barb")
	r.Set("foozle", "barc")
	r.Delete("foo")
	r.Delete("fooe")
	r.Delete("fooeyy")
	expectGet(t, r, "fooey", "bara")
	expectGet(t, r, "fooing", "barb")
	expectGet(t, r, "foozle", "barc")
}

func TestDeletePathCleanup(t *testing.T) {
	r := New()
	r.Set("alpha", "1")
	r.Set("alphabet", "2")
	r.Set("alphanumeric", "3")
	r.Set("beta", "4")
	r.Set("delta", "5")
	r.Delete("alpha")
	expectNotGet(t, r, "alpha")
	expectGet(t, r, "alphabet", "2")
	expectGet(t, r, "alphanumeric", "3")
	expectGet(t, r, "beta", "4")
	expectGet(t, r, "delta", "5")
	r.Set("alpha", "1")
	r.Delete("alphanumeric")
	expectGet(t, r, "alpha", "1")
	expectGet(t, r, "alphabet", "2")
	expectNotGet(t, r, "alphanumeric")
	expectGet(t, r, "beta", "4")
	expectGet(t, r, "delta", "5")
	r.Delete("alphabet")
	expectGet(t, r, "alpha", "1")
	expectNotGet(t, r, "alphabet")
	expectNotGet(t, r, "alphanumeric")
	expectGet(t, r, "beta", "4")
	expectGet(t, r, "delta", "5")
	r.Delete("alpha")
	expectNotGet(t, r, "alpha")
	expectNotGet(t, r, "alphabet")
	expectNotGet(t, r, "alphanumeric")
	expectGet(t, r, "beta", "4")
	expectGet(t, r, "delta", "5")
}

func TestSetAndGetCommonPrefix(t *testing.T) {
	r := New()
	r.Set("fooey", "bara")
	r.Set("fooing", "barb")
	r.Set("foozle", "barc")
	expectNotGet(t, r, "foo")
	expectGet(t, r, "fooey", "bara")
	expectGet(t, r, "fooing", "barb")
	expectGet(t, r, "foozle", "barc")
}

func TestSetAndGetSubstrings(t *testing.T) {
	r := New()
	r.Set("fooingly", "bara")
	r.Set("fooing", "barb")
	r.Set("foo", "barc")
	expectGet(t, r, "fooingly", "bara")
	expectGet(t, r, "fooing", "barb")
	expectGet(t, r, "foo", "barc")
}

func TestSetGetDeleteMixedOrder(t *testing.T) {
	rand.Seed(0)
	data := []string{
		"foo",
		"fooa",
		"foob",
		"fooc",
		"fooY",
		"fooZ",
		"fooaa",
		"fooab",
		"fooaaa",
		"fooaaZ",
		"fooaaaa",
		"fooaaac",
		"fooaaaaa",
		"fooaaaaY",
		"fooaaaaaa",
		"fooaaaaaaa",
		"fooaaaaaaaa",
	}
	for i := 0; i < 1000; i++ {
		r := New()
		for j := 0; j < 10; j++ {
			for _, k := range rand.Perm(len(data)) {
				expectNotGet(t, r, data[k])
				r.Set(data[k], data[k])
			}
			for _, key := range data {
				expectGet(t, r, key, key)
			}
			for _, k := range rand.Perm(len(data)) {
				r.Delete(data[k])
			}
		}
	}
}

func TestSetAndGetExhaustive3ByteLowercaseEnglish(t *testing.T) {
	var b [3]byte
	r := New()
	keys := make([]string, 0)
	for i := 97; i < 123; i++ {
		for j := 97; j < 123; j++ {
			for k := 97; k < 123; k++ {
				b[0], b[1], b[2] = byte(i), byte(j), byte(k)
				key := string(b[:])
				keys = append(keys, key)
			}
		}
	}
	for _, key := range keys {
		r.Set(key, key)
	}
	for _, key := range keys {
		expectGet(t, r, key, key)
	}
	for _, key := range keys {
		r.Delete(key)
		expectNotGet(t, r, key)
	}
}

func keystr(x []KV) string {
	z := []string{}
	for _, y := range x {
		z = append(z, y.Key)
	}
	sort.Strings(z)
	return strings.Join(z, " ")
}

func ukeystr(x []KV) string {
	z := []string{}
	for _, y := range x {
		z = append(z, y.Key)
	}
	return strings.Join(z, " ")
}

func TestSuggest(t *testing.T) {
	data := []string{
		"f",
		"x",
		"fo",
		"fx",
		"foo",
		"fooa",
		"foob",
		"fooc",
		"fooY",
		"fooZ",
		"fooaa",
		"fooab",
		"fooaaa",
		"fooaaZ",
		"fooaaaa",
		"fooaaac",
		"fooaaaaa",
		"fooaaaaY",
		"fooaaaaaa",
		"fooaaaaaaa",
		"fooaaaaaaaa",
	}
	r := New()
	var got, want string
	unlimited := len(data) + 1
	for _, key := range data {
		r.Set(key, key)
	}
	got = keystr(r.Suggest("foo", 0, unlimited))
	want = "foo"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.Suggest("foo", 1, unlimited))
	want = "fo foo fooY fooZ fooa foob fooc"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.Suggest("foo", 2, unlimited))
	want = "f fo foo fooY fooZ fooa fooaa fooab foob fooc fx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.Suggest("foo", 3, unlimited))
	want = "f fo foo fooY fooZ fooa fooaa fooaaZ fooaaa fooab foob fooc fx x"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.Suggest("fooaaa", 3, unlimited))
	want = "foo fooY fooZ fooa fooaa fooaaZ fooaaa fooaaaa fooaaaaY fooaaaaa fooaaaaaa fooaaac fooab foob fooc"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.Suggest("foobbb", 3, unlimited))
	want = "foo fooY fooZ fooa fooaa fooaaZ fooaaa fooab foob fooc"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.Suggest("foobbb", 4, unlimited))
	want = "fo foo fooY fooZ fooa fooaa fooaaZ fooaaa fooaaaa fooaaac fooab foob fooc"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
}

func TestSuggestReturnsResultsInIncreasingEditDistance(t *testing.T) {
	data := []string{
		"y",
		"yx",
		"xx",
		"xxx",
		"xxzx",
		"xxxxz",
		"xxxxxx",
		"aaaaaaa",
		"cccccccc",
		"bbbbbbbbb",
	}
	r := New()
	var got, want string
	unlimited := len(data) + 1
	for _, key := range data {
		r.Set(key, key)
	}
	got = ukeystr(r.Suggest("y", 10, unlimited))
	want = "y yx xx xxx xxzx xxxxz xxxxxx aaaaaaa cccccccc bbbbbbbbb"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = ukeystr(r.Suggest("y", 10, 5))
	want = "y yx xx xxx xxzx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = ukeystr(r.Suggest("y", 3, unlimited))
	want = "y yx xx xxx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = ukeystr(r.Suggest("xxxxxx", 3, unlimited))
	// Because of how we push candidates on the stack, we'll return prefixes
	// first, then closest candidates by edit distance.
	want = "xxx xxxxxx xxxxz xxzx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
}

func TestSuggestWithLimit(t *testing.T) {
	data := []string{
		"aaaaaaaa",
		"aaaaaaab",
		"aaaaaaba",
		"aaaaabaa",
		"aaaabaaa",
		"aaabaaaa",
		"aabaaaaa",
		"abaaaaaa",
		"baaaaaaa",
		"bbaaaaaa", // Not within edit distance 1 of "aaaaaaaa"
		"aaaaaabb", // Not within edit distance 1 of "aaaaaaaa"
		"aaaaabbb", // Not within edit distance 1 of "aaaaaaaa"
	}
	r := New()
	var got, want string
	for _, key := range data {
		r.Set(key, key)
	}
	got = ukeystr(r.Suggest("aaaaaaaa", 1, 1))
	want = "aaaaaaaa"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = ukeystr(r.Suggest("aaaaaaaa", 1, 2))
	want = "aaaaaaaa aaaaaaab"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = ukeystr(r.Suggest("aaaaaaaa", 1, 3))
	want = "aaaaaaaa aaaaaaab aaaaaaba"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = ukeystr(r.Suggest("aaaaaaaa", 1, 4))
	want = "aaaaaaaa aaaaaaab aaaaaaba aaaaabaa"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = ukeystr(r.Suggest("aaaaaaaa", 1, 5))
	want = "aaaaaaaa aaaaaaab aaaaaaba aaaaabaa aaaabaaa"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
}

func TestSuggestAfterExactPrefix(t *testing.T) {
	data := []string{
		"a",
		"aa",
		"aaafoo",
		"aaf",
		"aafo",
		"aafoo",
		"aafoox",
		"aafooxx",
		"aafooxxx",
		"aafox",
		"aafx",
		"aafxx",
		"abfoo",
		"abfooxx",
		"b",
		"bbfoo",
		"foo",
	}
	r := New()
	var got, want string
	unlimited := len(data) + 1
	for _, key := range data {
		r.Set(key, key)
	}
	got = keystr(r.SuggestAfterExactPrefix("aafoo", 2, 0, unlimited))
	want = "aafoo"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestAfterExactPrefix("aafoo", 2, 1, unlimited))
	want = "aaafoo aafo aafoo aafoox aafox"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestAfterExactPrefix("aafoo", 2, 2, unlimited))
	want = "aaafoo aaf aafo aafoo aafoox aafooxx aafox aafx aafxx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestAfterExactPrefix("aafoo", 2, 3, unlimited))
	want = "aa aaafoo aaf aafo aafoo aafoox aafooxx aafooxxx aafox aafx aafxx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
}

func TestSuggestSuffixes(t *testing.T) {
	data := []string{
		"", "afoo", "f", "fo", "foo", "fooey", "fooeyz", "fooeyzz", "foox",
		"fooxx", "fooxxx", "fooxxxaaaaa", "fooz", "fox", "fx", "fxx", "gog",
		"gogx", "gogy", "gogyy", "gogyyy",
	}
	r := New()
	var got, want string
	unlimited := len(data) + 1
	for _, key := range data {
		r.Set(key, key)
	}
	got = keystr(r.SuggestSuffixes("foo", 0, unlimited))
	want = "foo fooey fooeyz fooeyzz foox fooxx fooxxx fooxxxaaaaa fooz"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestSuffixes("foo", 1, unlimited))
	want = "afoo fo foo fooey fooeyz fooeyzz foox fooxx fooxxx fooxxxaaaaa fooz fox"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestSuffixes("foo", 2, unlimited))
	want = "afoo f fo foo fooey fooeyz fooeyzz foox fooxx fooxxx fooxxxaaaaa fooz fox fx fxx gog gogx gogy gogyy gogyyy"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestSuffixes("foo", 3, unlimited))
	want = " afoo f fo foo fooey fooeyz fooeyzz foox fooxx fooxxx fooxxxaaaaa fooz fox fx fxx gog gogx gogy gogyy gogyyy"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
}

func TestSuggestSuffixesAfterExactPrefix(t *testing.T) {
	data := []string{
		"foo", "xxxfoo", "xxxgoo", "xyyfoo", "xyzfoo", "xyzfoox", "xyzfooxx",
		"xyzfooxxxxxx", "xyzgo", "xyzgog", "xyzgogxxxxx", "xyzgoo", "xyzgooxxxx",
		"xyzxxx", "xyzxxxxxxxxxx", "xyxfoo",
	}
	r := New()
	var got, want string
	unlimited := len(data) + 1
	for _, key := range data {
		r.Set(key, key)
	}
	got = keystr(r.SuggestSuffixesAfterExactPrefix("xyzfoo", 3, 0, unlimited))
	want = "xyzfoo xyzfoox xyzfooxx xyzfooxxxxxx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestSuffixesAfterExactPrefix("xyzfoo", 3, 1, unlimited))
	want = "xyzfoo xyzfoox xyzfooxx xyzfooxxxxxx xyzgoo xyzgooxxxx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestSuffixesAfterExactPrefix("xyzfoo", 3, 2, unlimited))
	want = "xyzfoo xyzfoox xyzfooxx xyzfooxxxxxx xyzgo xyzgog xyzgogxxxxx xyzgoo xyzgooxxxx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
	got = keystr(r.SuggestSuffixesAfterExactPrefix("xyzfoo", 3, 3, unlimited))
	want = "xyzfoo xyzfoox xyzfooxx xyzfooxxxxxx xyzgo xyzgog xyzgogxxxxx xyzgoo xyzgooxxxx xyzxxx xyzxxxxxxxxxx"
	if got != want {
		t.Errorf("Got '%v', want '%v'\n", got, want)
	}
}

// Returns the edit distance between s and t.
func editDistance(s string, t string) int8 {
	rs := extractRunes(s)
	rt := extractRunes(t)
	return editDistanceHelper(rs, rt)
}

func editDistanceHelper(s []rune, t []rune) int8 {
	if len(s) == 0 {
		return int8(len(t))
	} else if len(t) == 0 {
		return int8(len(s))
	} else if s[len(s)-1] == t[len(t)-1] {
		return editDistanceHelper(s[:len(s)-1], t[:len(t)-1])
	}
	x := editDistanceHelper(s, t[:len(t)-1])
	y := editDistanceHelper(s[:len(s)-1], t)
	z := editDistanceHelper(s[:len(s)-1], t[:len(t)-1])
	d := x
	if y < d {
		d = y
	}
	if z < d {
		d = z
	}
	return 1 + d
}

// Start with a seed string of length k, repeatedly select a sample string,
// choose an edit to apply (delete, substitute, insert) and return the edited
// string to the list of samples. Stop when there are n distinct samples.
func generateEdits(k int, n int) []string {
	alphabet := []rune{'A', 'ἑ', 'й', 'ლ', 'ô', 'Z', '1'}
	seed := []rune{}
	for len(seed) < k {
		seed = append(seed, alphabet[rand.Intn(len(alphabet))])
	}
	seedStr := string(seed)
	resultSet := map[string]bool{}
	resultSet[seedStr] = true
	results := []string{seedStr}
	for len(results) < n {
		sample := results[rand.Intn(len(results))]
		runes := extractRunes(sample)
		if len(runes) == 0 {
			continue
		}
		switch rand.Intn(3) {
		case 0: // Delete
			i := rand.Intn(len(runes))
			runes = append(runes[:i], runes[i+1:]...)
		case 1: // Insert
			i, j := rand.Intn(len(runes)), rand.Intn(len(alphabet))
			runes = append(append(runes[:i], alphabet[j]), runes[i:]...)
		case 2: // Substitute
			i, j := rand.Intn(len(runes)), rand.Intn(len(alphabet))
			runes = append(append(runes[:i], alphabet[j]), runes[i+1:]...)
		}
		edited := string(runes)
		if !resultSet[edited] {
			resultSet[edited] = true
			results = append(results, edited)
		}
	}
	return results
}

// Returns all strings in xs that are at most edit distance d from s.
func filterByEditDistance(xs []string, s string, d int8) []KV {
	results := []KV{}
	for _, x := range xs {
		if editDistance(x, s) <= d {
			results = append(results, KV{Key: x, Value: x})
		}
	}
	return results
}

func TestSuggestFuzz(t *testing.T) {
	rand.Seed(0)
	r := New()
	haystack := generateEdits(5, 5000)
	for _, s := range haystack {
		r.Set(s, s)
	}
	for dist := int8(0); dist < 6; dist++ {
		needle := haystack[rand.Intn(len(haystack))]
		results := keystr(r.Suggest(needle, dist, len(haystack)))
		expected := keystr(filterByEditDistance(haystack, needle, dist))
		if results != expected {
			t.Errorf("When asking for strings edit distance %v away from %v,"+
				"got:\n%v\nbut want:\n%v", dist, needle, results, expected)
		}
	}
}
