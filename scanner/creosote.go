package scanner

import (
	"fmt"
	"os"
	"strings"

	"github.com/yargevad/filepathx"

	"github.com/brianpan/go-creosote/walk"
	"github.com/go-python/gpython/ast"
	_ "github.com/go-python/gpython/modules"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
)

type PythonScanner struct {
	File string
}

type CreosoteResult struct {
	Filename string
	State    ScanState
	Lineno   int
}

type ScanState uint

const (
	OK = ScanState(iota + 1)
	PotentialThreat
	Threat
)

func ScanAll(dirname string) (r []CreosoteResult, err error) {
	r = make([]CreosoteResult, 0)
	pattern := fmt.Sprintf("%s/*.py", dirname)
	files, err := filepathx.Glob(pattern)
	if err != nil {
		return
	}

	for _, file := range files {
		_, err := os.Open(file)
		if err != nil {
			fmt.Println("err, ", err)
			continue
		}
		ps := &PythonScanner{
			File: file,
		}
		if lineno, state, err := ps.Scan(); err != nil {
			fmt.Println("err, ", err)
			continue
		} else {
			if state != OK {
				r = append(r, CreosoteResult{
					Filename: file,
					State:    state,
					Lineno:   lineno,
				})
			}
		}
	}

	return
}

func checkOpen(a ast.Ast) (state ScanState) {
	state = PotentialThreat
	arr := make([]bool, 0)
	detector := func(a ast.Ast) bool {
		if attr, ok := a.(*ast.Attribute); ok {
			if attr.Attr == "open" {
				arr = append(arr, true)
				return false
			}
		}
		return true
	}
	ast.Walk(a, detector)
	if len(arr) > 0 {
		state = Threat
	}

	return
}

// if ".getmembers()" in ast.unparse(for_node.iter)
func checkgetmembers(forNode *ast.For) (ok bool) {
	iter := forNode.Iter
	arr := make([]bool, 0)
	detector := func(a ast.Ast) bool {
		if attr, ok := a.(*ast.Attribute); ok {
			if attr.Attr == "getmembers" {
				arr = append(arr, true)
				return false
			}
		}

		return true
	}
	ast.Walk(iter, detector)

	if len(arr) > 0 {
		ok = true
	}
	return
}

// reference: https://www.trellix.com/en-us/about/newsroom/stories/research/tarfile-exploiting-the-world.html
func (ps *PythonScanner) Scan() (lineno int, state ScanState, err error) {

	// collect tokens
	in, err := os.Open(ps.File)
	if err != nil {
		return
	}

	a, err := parser.Parse(in, ps.File, py.ExecMode)

	if err != nil {
		return
	}

	dummy := func(a ast.Ast) bool {
		return true
	}

	walk.Walk(a, dummy)

	// for node, parent := range walk.ParentMap {
	// 	fmt.Println(fmt.Sprintf("node %T, parent %T", node, parent))
	// }
	state = ScanState(OK)

	// main function to visit AST to detect the threats
	detector := func(a ast.Ast) bool {
		if attr, ok := a.(*ast.Attribute); ok {
			// extractall method
			if attr.Attr == "extractall" || attr.Attr == "extract" {
				// get node.parent.parent.parent
				if p, ok := walk.GetParent(a); ok {
					if pp, ok := walk.GetParent(p); ok {
						if ppp, ok := walk.GetParent(pp); ok {
							// get with node for extractall case
							if attr.Attr == "extractall" {
								if withNode, ok := ppp.(*ast.With); ok {
									for _, wi := range withNode.Items {
										// item.context_expr and type(item.context_expr) == ast.Call
										if call, ok := wi.ContextExpr.(*ast.Call); ok {
											// type(item.context_expr.func) == ast.Attribute and item.context_expr.func.attr == "open"
											if attr, ok := call.Func.(*ast.Attribute); ok && attr.Attr == "open" {
												args := call.Args
												keywords := call.Keywords
												// len(args) > 1 and type(args[1]) == ast.Constant and "r" in args[1].value
												if len(args) > 1 {
													if con, ok := args[1].(*ast.Str); ok && strings.Contains(string(con.S), "r") {
														// vuln found
														state = ScanState(Threat)
														lineno = p.GetLineno()
														// no need to keep traveral
														return false
													}
													// len(keywords) > 1:
												} else if len(keywords) > 1 {
													for _, k := range keywords {
														//  if keyword.arg and keyword.arg == "mode"
														if k.Arg == "mode" {
															// type(keyword.value) == ast.Constant and "r" in keyword.value.value
															if con, ok := k.Value.(*ast.Str); ok && strings.Contains(string(con.S), "r") {
																// vuln found
																state = ScanState(Threat)
																lineno = p.GetLineno()
																// no need to keep traveral
																return false
															}
														}
													}
												} else if len(args) == 1 {
													// vuln found
													state = ScanState(Threat)
													lineno = p.GetLineno()
													// no need to keep traveral
													return false
												} else {
													state = ScanState(PotentialThreat)
													lineno = p.GetLineno()
													return false
												}
											}

										}
									}
								}
								return true
							}

							// extract case
							// if type(node.parent) == ast.Call and node.parent.args and (len(node.parent.args) != 0 or len(node.parent.keywords) != 0)
							if call, ok := p.(*ast.Call); ok && (len(call.Args) != 0 || len(call.Keywords) != 0) {
								// if type(node.parent.parent.parent) == ast.For
								if forNode, ok := ppp.(*ast.For); ok {
									if checkgetmembers(forNode) {
										if forParent, ok := walk.GetParent(forNode); ok {
											// # check if we have an open with read
											state = checkOpen(forParent)
											lineno = pp.GetLineno()
											// no need to keep traveral
											return false
										}
									}
								}
							}

							// elif node.attr == "extract":
							state = ScanState(PotentialThreat)
							lineno = p.GetLineno()
							return false
						}
					}
				}
			}
		}
		return true
	}

	ast.Walk(a, detector)

	return
}
