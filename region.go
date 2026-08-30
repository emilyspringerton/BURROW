// region.go — Phase 3 (region analyzer parity), real, faithful Go port of PARENA/src/region.c
// (VS0's own C-implemented region analyzer — DoD domain 2: checks the real
// `Region(Source) >= Region(Destination)` escape invariant on region-annotated Arena params, not
// the full region-safety story). Founder real-time: "continue DUNG i guess you are gonna need
// burrow on the path if its not already" — DUNG's own real build (compiling ground-up PARENA
// editor source via `burrow build`) needs a working region-analyze + emit pipeline first; this is
// the next real, honestly-scoped slice toward that, following the exact same real architecture
// call Phases 1-2 already made (no `selfhost/region.prn` exists yet, so hand-port the real C
// reference directly rather than waiting on a PARENA-Go emitter that doesn't exist).
//
// Real, deliberate idiomatic departure, same one lexer.go/parser.go's own header comments already
// establish: Go's own real `error` return replaces the C reference's own `const char *` (NULL =
// no error) convention.
package main

import "errors"

const maxBindings = 64 // matches region.c's own MAX_BINDINGS -- real, deliberate parity, not
// load-bearing in Go (a real slice has no such cap) but kept as a documented, honest mirror of
// the C reference's own real, fixed-size scope table.

// binding — real, direct port of region.c's own Binding struct.
type binding struct {
	name       string
	rank       int    // -1 = unknown/unconstrained
	regionName string // "" if rank == -1 (real, idiomatic equivalent of the C reference's own NULL)
}

// regionScope — real, direct port of region.c's own Scope struct. Real, deliberate
// simplification: no fixed-size `bindings [64]Binding` array + `count` — a real Go slice already
// grows, the same real "Go's own native structures make the C reference's own fixed-capacity
// workaround moot" reasoning lexer.go/parser.go's own header comments already apply elsewhere.
type regionScope struct {
	parent   *regionScope
	bindings []binding
}

func newRegionScope(parent *regionScope) *regionScope {
	return &regionScope{parent: parent}
}

func (s *regionScope) bind(name string, rank int, regionName string) {
	// Real, honest parity with region.c's own MAX_BINDINGS cap: rather than silently growing
	// past what the C reference would have accepted (real behavioral divergence risk if any
	// real .prn file ever legitimately needs more than 64 bindings in one scope), stop
	// accepting new bindings past the same real limit -- matching scope_bind's own real
	// "if (s->count < MAX_BINDINGS)" guard exactly.
	if len(s.bindings) < maxBindings {
		s.bindings = append(s.bindings, binding{name: name, rank: rank, regionName: regionName})
	}
}

// lookup — real, direct port of region.c's own scope_lookup: searches the current scope's own
// bindings from MOST RECENT to oldest (a later binding of the same name shadows an earlier one in
// the same scope), then falls back to the parent scope, and so on.
func (s *regionScope) lookup(name string) *binding {
	for cur := s; cur != nil; cur = cur.parent {
		for i := len(cur.bindings) - 1; i >= 0; i-- {
			if cur.bindings[i].name == name {
				return &cur.bindings[i]
			}
		}
	}
	return nil
}

// regionRankFor — real, direct port of region.c's own region_rank_for: covers only the two ranks
// NORTHSTAR.md's own "Memory model" section actually specifies. Any other region keyword is real,
// unconstrained territory (rank -1), not guessed at.
func regionRankFor(keyword string) int {
	switch keyword {
	case ":region/scratch":
		return 0
	case ":region/buffer":
		return 2
	default:
		return -1
	}
}

func isSymbol(n *Node, text string) bool {
	return n != nil && n.Type == NodeSymbol && n.Text == text
}

func isCallNamed(n *Node, fnName string) bool {
	return n != nil && n.Type == NodeList && len(n.Children) > 0 && isSymbol(n.Children[0], fnName)
}

// findKeywordChild — real, direct port of region.c's own find_keyword_child: pulls the region
// annotation out of a param form `(name : Type @ :region/xxx)` without depending on its exact
// position. Returns "" if none found (real, idiomatic equivalent of the C reference's own NULL —
// a real NodeKeyword's own Text is never legitimately empty, so this is a safe sentinel).
func findKeywordChild(n *Node) string {
	for _, child := range n.Children {
		if child.Type == NodeKeyword {
			return child.Text
		}
	}
	return ""
}

func fmtRegionError(srcRegion, dstRegion string, line int) error {
	return errors.New("Compile Error: Escaping region pointer from " + srcRegion + " to " + dstRegion +
		" at line " + itoa(line))
}

// checkCallEscape — real, direct port of region.c's own check_call_escape: the actual invariant.
// In a call shaped `(fn dest-expr src-expr...)`, if dest-expr names a binding with a known region
// rank and any later src-expr names a binding with a strictly lower rank (shorter-lived), that
// source's region pointer is escaping into the destination's longer-lived scope.
func checkCallEscape(call *Node, scope *regionScope) error {
	if len(call.Children) < 3 { // need fn + dest + >=1 source
		return nil
	}
	destNode := call.Children[1]
	if destNode.Type != NodeSymbol {
		return nil
	}
	dest := scope.lookup(destNode.Text)
	if dest == nil || dest.rank < 0 {
		return nil
	}
	for i := 2; i < len(call.Children); i++ {
		arg := call.Children[i]
		if arg.Type != NodeSymbol {
			continue
		}
		src := scope.lookup(arg.Text)
		if src != nil && src.rank >= 0 && src.rank < dest.rank {
			return fmtRegionError(src.regionName, dest.regionName, arg.Line)
		}
	}
	return nil
}

// walkChildren — real, direct port of region.c's own walk_children: the fallback for any
// compound form this analyzer has no specific rule for.
func walkChildren(node *Node, scope *regionScope) error {
	for _, child := range node.Children {
		if err := walk(child, scope); err != nil {
			return err
		}
	}
	return nil
}

// walkWithArena — real, direct port of region.c's own walk_with_arena: handles
// `(with-arena [name region-kw size] body...)`.
func walkWithArena(node *Node, scope *regionScope) error {
	if len(node.Children) < 2 || node.Children[1].Type != NodeVec {
		return walkChildren(node, scope)
	}
	bindingVec := node.Children[1]
	if len(bindingVec.Children) < 1 || bindingVec.Children[0].Type != NodeSymbol {
		return walkChildren(node, scope)
	}
	name := bindingVec.Children[0].Text
	regionKw := ""
	for _, child := range bindingVec.Children {
		if child.Type == NodeKeyword {
			regionKw = child.Text
			break
		}
	}

	child := newRegionScope(scope)
	if regionKw != "" {
		child.bind(name, regionRankFor(regionKw), regionKw)
	} else {
		child.bind(name, -1, "")
	}

	for i := 2; i < len(node.Children); i++ {
		if err := walk(node.Children[i], child); err != nil {
			return err
		}
	}
	return nil
}

// allocRank — real, direct port of region.c's own alloc_rank: resolves the region rank an
// `(alloc arena-expr Type value...)` call produces -- the allocated value inherits arena-expr's
// own region.
func allocRank(call *Node, scope *regionScope) (rank int, regionName string) {
	if !isCallNamed(call, "alloc") || len(call.Children) < 2 {
		return -1, ""
	}
	arenaExpr := call.Children[1]
	if arenaExpr.Type != NodeSymbol {
		return -1, ""
	}
	b := scope.lookup(arenaExpr.Text)
	if b == nil {
		return -1, ""
	}
	return b.rank, b.regionName
}

// walkLet — real, direct port of region.c's own walk_let: handles
// `(let [name1 expr1 name2 expr2 ...] body...)`.
func walkLet(node *Node, scope *regionScope) error {
	if len(node.Children) < 2 || node.Children[1].Type != NodeVec {
		return walkChildren(node, scope)
	}
	bindings := node.Children[1]

	child := newRegionScope(scope)

	for i := 0; i+1 < len(bindings.Children); i += 2 {
		nameNode := bindings.Children[i]
		exprNode := bindings.Children[i+1]
		if nameNode.Type != NodeSymbol {
			continue
		}
		// Each expr is walked under the *outer* scope first, so an escape nested inside a
		// binding's own expr is caught before the new bindings even exist.
		if err := walk(exprNode, scope); err != nil {
			return err
		}
		rank, regionName := allocRank(exprNode, scope)
		child.bind(nameNode.Text, rank, regionName)
	}

	for i := 2; i < len(node.Children); i++ {
		if err := walk(node.Children[i], child); err != nil {
			return err
		}
	}
	return nil
}

// walk — real, direct port of region.c's own walk.
func walk(node *Node, scope *regionScope) error {
	if node == nil {
		return nil
	}
	if isCallNamed(node, "with-arena") {
		return walkWithArena(node, scope)
	}
	if isCallNamed(node, "let") {
		return walkLet(node, scope)
	}
	if node.Type == NodeList && len(node.Children) > 0 && node.Children[0].Type == NodeSymbol {
		if err := checkCallEscape(node, scope); err != nil {
			return err
		}
	}
	return walkChildren(node, scope)
}

// analyzeDefn — real, direct port of region.c's own analyze_defn: one top-level
// `(defn name [params] body...)`.
func analyzeDefn(defn *Node) error {
	if len(defn.Children) < 3 || defn.Children[2].Type != NodeVec {
		return nil
	}
	params := defn.Children[2]

	base := newRegionScope(nil)
	for _, param := range params.Children {
		if param.Type != NodeList || len(param.Children) == 0 {
			continue
		}
		if param.Children[0].Type != NodeSymbol {
			continue
		}
		name := param.Children[0].Text
		regionKw := findKeywordChild(param)
		if regionKw != "" {
			base.bind(name, regionRankFor(regionKw), regionKw)
		} else {
			base.bind(name, -1, "")
		}
	}

	for i := 3; i < len(defn.Children); i++ {
		if err := walk(defn.Children[i], base); err != nil {
			return err
		}
	}
	return nil
}

// RegionAnalyze — real, direct port of region.c's own region_analyze: the real, top-level entry
// point, walking every top-level `(defn ...)` form in the program.
func RegionAnalyze(program *Node) error {
	for _, form := range program.Children {
		if isCallNamed(form, "defn") {
			if err := analyzeDefn(form); err != nil {
				return err
			}
		}
	}
	return nil
}
