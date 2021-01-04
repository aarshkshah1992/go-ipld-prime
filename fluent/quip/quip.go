// quip is a package of quick ipld patterns.
//
// Most quip functions take a pointer to an error as their first argument.
// This has two purposes: if there's an error there, the quip function will do nothing;
// and if the quip function does something and creates an error, it puts it there.
// The effect of this is that most logic can be written very linearly.
//
// quip functions can be used to increase brevity without worrying about performance costs.
// None of the quip functions cause additional allocations in the course of their work.
// Benchmarks indicate no measurable speed penalties versus longhand manual error checking.
//
// This package is currently considered experimental, and may change.
// Feel free to use it, but be advised that code using it may require frequent
// updating until things settle; naming conventions in this package may be
// revised with relatively little warning.
package quip

import (
	"github.com/ipld/go-ipld-prime"
)

// TODO:REVIEW: a few things about this outline and its symbol naming:
//  - consistency/nomenclature check:
//    - fluent package uses "BuildMap" at package scope.  And then "CreateMap" on its NB facade (which takes a callback param).  It's calling "NodeBuilder.BeginMap" internally, of course.  (is fluent's current choice odd and deserving of revising?)
//      - the "Build" vs "Create" prefixes here indicate methods that start with a builder or prototype and return a full node, vs methods that work on assemblers.  (the fluent package doesn't have any methods that *don't* take callbacks, so there's no name component to talk about that.)
//    - here currently we've used "BeginMap" for something that returns a MapAssembler (consistent with ipld.NB), and "BuildMap" takes a callback.
//  - ditto the above bullet tree for List instead of Map.
//  - many of these methods could have varying parameter types: ipld.Node forms, interface{} and dtrt forms, and callback-taking-assembler forms.  They all need unique names.
//    - this is the case for values, and similarly for keys (though for keys the most important ones are probably string | ipld.Node | pathSegment).  The crossproduct there is sizable.
//    - do we want to fill out this matrix completely?
//    - what naming convention can we use to make this consistent?
//  - do we actually want the non-callback/returns-MapAssembler BeginMap style to be something this package bothers to export?
//    - hard to imagine actually wanting to use these.
//    - since it turns out that the callbacks *do* get optimized quite reasonably by the compiler in common cases, there's no reason to avoid them.

// TODO: a few notable gaps in what's provided, which should improve terseness even more:
//  - we don't have top-level functions for doing a full Build returning a Node (and saving you the np.NewBuilder preamble and nb.Build postamble).  see naming issues discussion.
//  - we don't have great shorthand for scalar assignment at the leaves.  current best is a composition like `quip.AbsorbError(&err, na.AssignString(x))`.
//    - do we want to make an `quip.Assign{Kind}(*error, *ipld.NodeAssembler, {kindPrimitive})` for each kind?  seems like a lot of symbols.
//    - there's also generally a whole `quip.MapEntry(... {...})` or `ListEntry` clause around *that*, with the single Absorb+Assign line in the middle.  that's boilerplate that needs trimming as well.
//      - so... twice as many Assign helpers, because it's kinda necessary to specialize them to map and list assemblers too?  uff.

func BeginMap(na ipld.NodeAssembler, sizeHint int64) ipld.MapAssembler {
	ma, err := na.BeginMap(sizeHint)
	if err != nil {
		panic(err)
	}
	return ma
}

func BuildMap(na ipld.NodeAssembler, sizeHint int64, fn func(ma ipld.MapAssembler)) (err error) {
	// TODO(mvdan): if the call stack contains another quip helper, panic
	// immediately, to prevent incorrectly using BuildMap instead of Map.
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	Map(na, sizeHint, fn)
	return nil
}

func Map(na ipld.NodeAssembler, sizeHint int64, fn func(ma ipld.MapAssembler)) {
	ma, err := na.BeginMap(sizeHint)
	if err != nil {
		panic(err)
	}
	fn(ma)
	err = ma.Finish()
	if err != nil {
		panic(err)
	}
}

func MapEntry(ma ipld.MapAssembler, k string, fn func(va ipld.NodeAssembler)) {
	va, err := ma.AssembleEntry(k)
	if err != nil {
		panic(err)
	}
	fn(va)
}

func BeginList(na ipld.NodeAssembler, sizeHint int64) ipld.ListAssembler {
	la, err := na.BeginList(sizeHint)
	if err != nil {
		panic(err)
	}
	return la
}

func BuildList(na ipld.NodeAssembler, sizeHint int64, fn func(la ipld.ListAssembler)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	List(na, sizeHint, fn)
	return nil
}

func List(na ipld.NodeAssembler, sizeHint int64, fn func(la ipld.ListAssembler)) {
	la, err := na.BeginList(sizeHint)
	if err != nil {
		panic(err)
	}
	fn(la)
	err = la.Finish()
	if err != nil {
		panic(err)
	}
}

func ListEntry(la ipld.ListAssembler, fn func(va ipld.NodeAssembler)) {
	fn(la.AssembleValue())
}

func CopyRange(la ipld.ListAssembler, src ipld.Node, start, end int64) {
	if start >= src.Length() {
		return
	}
	if end < 0 {
		end = src.Length()
	}
	if end < start {
		return
	}
	for i := start; i < end; i++ {
		n, err := src.LookupByIndex(i)
		if err != nil {
			panic(err)
		}
		if err := la.AssembleValue().AssignNode(n); err != nil {
			panic(err)
		}
	}
}

type String string

func (s String) Assign(va ipld.NodeAssembler) {
	if err := va.AssignString(string(s)); err != nil {
		panic(err)
	}
}

func AssignString(s string) func(ipld.NodeAssembler) {
	return String(s).Assign
}

type Int int64

func (s Int) Assign(va ipld.NodeAssembler) {
	if err := va.AssignInt(int64(s)); err != nil {
		panic(err)
	}
}

func AssignInt(i int64) func(ipld.NodeAssembler) {
	return Int(i).Assign
}
