package walk

import (
	"fmt"

	past "github.com/go-python/gpython/ast"
)

var ParentMap map[past.Ast]past.Ast

func init() {
	Cleanup()
}

func Cleanup() {
	ParentMap = make(map[past.Ast]past.Ast, 0)
}

func Walk(ast past.Ast, Visit func(past.Ast) bool) {
	if ast == nil {
		return
	}
	if !Visit(ast) {
		return
	}

	// walk a single ast
	walk := func(parent past.Ast, ast past.Ast) {
		ParentMap[ast] = parent
		Walk(ast, Visit)
	}

	// walkStmts walks all the statements in the slice passed in
	walkStmts := func(parent past.Ast, stmts []past.Stmt) {
		for _, stmt := range stmts {
			walk(parent, stmt)
		}
	}

	// walkExprs walks all the exprs in the slice passed in
	walkExprs := func(parent past.Ast, exprs []past.Expr) {
		for _, expr := range exprs {
			walk(parent, expr)
		}
	}

	// walkComprehensions walks all the comprehensions in the slice passed in
	walkComprehensions := func(parent past.Ast, comprehensions []past.Comprehension) {
		for _, comprehension := range comprehensions {
			// Target Expr
			// Iter   Expr
			// Ifs    []Expr
			walk(parent, comprehension.Target)
			walk(parent, comprehension.Iter)
			walkExprs(parent, comprehension.Ifs)
		}
	}

	switch node := ast.(type) {

	// Module nodes

	case *past.Module:
		// Body []Stmt
		walkStmts(ast, node.Body)

	case *past.Interactive:
		// Body []Stmt
		walkStmts(ast, node.Body)

	case *past.Expression:
		// Body Expr
		walk(ast, node.Body)

	case *past.Suite:
		// Body []Stmt
		walkStmts(ast, node.Body)

	// Statememt nodes

	case *past.FunctionDef:
		// Name          Identifier
		// Args          *Arguments
		// Body          []Stmt
		// DecoratorList []Expr
		// Returns       Expr
		if node.Args != nil {
			walk(ast, node.Args)
		}
		walkStmts(ast, node.Body)
		walkExprs(ast, node.DecoratorList)
		walk(ast, node.Returns)

	case *past.ClassDef:
		// Name          Identifier
		// Bases         []Expr
		// Keywords      []*Keyword
		// Starargs      Expr
		// Kwargs        Expr
		// Body          []Stmt
		// DecoratorList []Expr
		walkExprs(ast, node.Bases)
		for _, k := range node.Keywords {
			walk(ast, k)
		}
		walk(ast, node.Starargs)
		walk(ast, node.Kwargs)
		walkStmts(ast, node.Body)
		walkExprs(ast, node.DecoratorList)

	case *past.Return:
		// Value Expr
		walk(ast, node.Value)

	case *past.Delete:
		// Targets []Expr
		walkExprs(ast, node.Targets)

	case *past.Assign:
		// Targets []Expr
		// Value   Expr
		walkExprs(ast, node.Targets)
		walk(ast, node.Value)

	case *past.AugAssign:
		// Target Expr
		// Op     OperatorNumber
		// Value  Expr
		walk(ast, node.Target)
		walk(ast, node.Value)

	case *past.For:
		// Target Expr
		// Iter   Expr
		// Body   []Stmt
		// Orelse []Stmt
		walk(ast, node.Target)
		walk(ast, node.Iter)
		walkStmts(ast, node.Body)
		walkStmts(ast, node.Orelse)

	case *past.While:
		// Test   Expr
		// Body   []Stmt
		// Orelse []Stmt
		walk(ast, node.Test)
		walkStmts(ast, node.Body)
		walkStmts(ast, node.Orelse)

	case *past.If:
		// Test   Expr
		// Body   []Stmt
		// Orelse []Stmt
		walk(ast, node.Test)
		walkStmts(ast, node.Body)
		walkStmts(ast, node.Orelse)

	case *past.With:
		// Items []*WithItem
		// Body  []Stmt
		for _, wi := range node.Items {
			walk(ast, wi)
		}
		walkStmts(ast, node.Body)

	case *past.Raise:
		// Exc   Expr
		// Cause Expr
		walk(ast, node.Exc)
		walk(ast, node.Cause)

	case *past.Try:
		// Body      []Stmt
		// Handlers  []*ExceptHandler
		// Orelse    []Stmt
		// Finalbody []Stmt
		walkStmts(ast, node.Body)
		for _, h := range node.Handlers {
			walk(ast, h)
		}
		walkStmts(ast, node.Orelse)
		walkStmts(ast, node.Finalbody)

	case *past.Assert:
		// Test Expr
		// Msg  Expr
		walk(ast, node.Test)
		walk(ast, node.Msg)

	case *past.Import:
		// Names []*Alias
		for _, n := range node.Names {
			walk(ast, n)
		}

	case *past.ImportFrom:
		// Module Identifier
		// Names  []*Alias
		// Level  int
		for _, n := range node.Names {
			walk(ast, n)
		}

	case *past.Global:
		// Names []Identifier

	case *past.Nonlocal:
		// Names []Identifier

	case *past.ExprStmt:
		// Value Expr
		walk(ast, node.Value)

	case *past.Pass:

	case *past.Break:

	case *past.Continue:

	// Expr nodes

	case *past.BoolOp:
		// Op     BoolOpNumber
		// Values []Expr
		walkExprs(ast, node.Values)

	case *past.BinOp:
		// Left  Expr
		// Op    OperatorNumber
		// Right Expr
		walk(ast, node.Left)
		walk(ast, node.Right)

	case *past.UnaryOp:
		// Op      UnaryOpNumber
		// Operand Expr
		walk(ast, node.Operand)

	case *past.Lambda:
		// Args *Arguments
		// Body Expr
		if node.Args != nil {
			walk(ast, node.Args)
		}
		walk(ast, node.Body)

	case *past.IfExp:
		// Test   Expr
		// Body   Expr
		// Orelse Expr
		walk(ast, node.Test)
		walk(ast, node.Body)
		walk(ast, node.Orelse)

	case *past.Dict:
		// Keys   []Expr
		// Values []Expr
		walkExprs(ast, node.Keys)
		walkExprs(ast, node.Values)

	case *past.Set:
		// Elts []Expr
		walkExprs(ast, node.Elts)

	case *past.ListComp:
		// Elt        Expr
		// Generators []Comprehension
		walk(ast, node.Elt)
		walkComprehensions(ast, node.Generators)

	case *past.SetComp:
		// Elt        Expr
		// Generators []Comprehension
		walk(ast, node.Elt)
		walkComprehensions(ast, node.Generators)

	case *past.DictComp:
		// Key        Expr
		// Value      Expr
		// Generators []Comprehension
		walk(ast, node.Key)
		walk(ast, node.Value)
		walkComprehensions(ast, node.Generators)

	case *past.GeneratorExp:
		// Elt        Expr
		// Generators []Comprehension
		walk(ast, node.Elt)
		walkComprehensions(ast, node.Generators)

	case *past.Yield:
		// Value Expr
		walk(ast, node.Value)

	case *past.YieldFrom:
		// Value Expr
		walk(ast, node.Value)

	case *past.Compare:
		// Left        Expr
		// Ops         []CmpOp
		// Comparators []Expr
		walk(ast, node.Left)
		walkExprs(ast, node.Comparators)

	case *past.Call:
		// Func     Expr
		// Args     []Expr
		// Keywords []*Keyword
		// Starargs Expr
		// Kwargs   Expr
		walk(ast, node.Func)
		walkExprs(ast, node.Args)
		for _, k := range node.Keywords {
			walk(ast, k)
		}
		walk(ast, node.Starargs)
		walk(ast, node.Kwargs)

	case *past.Num:
		// N Object

	case *past.Str:
		// S py.String

	case *past.Bytes:
		// S py.Bytes

	case *past.NameConstant:
		// Value Singleton

	case *past.Ellipsis:

	case *past.Attribute:
		// Value Expr
		// Attr  Identifier
		// Ctx   ExprContext
		walk(ast, node.Value)

	case *past.Subscript:
		// Value Expr
		// Slice Slicer
		// Ctx   ExprContext
		walk(ast, node.Value)
		walk(ast, node.Slice)

	case *past.Starred:
		// Value Expr
		// Ctx   ExprContext
		walk(ast, node.Value)

	case *past.Name:
		// Id  Identifier
		// Ctx ExprContext

	case *past.List:
		// Elts []Expr
		// Ctx  ExprContext
		walkExprs(ast, node.Elts)

	case *past.Tuple:
		// Elts []Expr
		// Ctx  ExprContext
		walkExprs(ast, node.Elts)

	// Slicer nodes

	case *past.Slice:
		// Lower Expr
		// Upper Expr
		// Step  Expr
		walk(ast, node.Lower)
		walk(ast, node.Upper)
		walk(ast, node.Step)

	case *past.ExtSlice:
		// Dims []Slicer
		for _, s := range node.Dims {
			walk(ast, s)
		}

	case *past.Index:
		// Value Expr
		walk(ast, node.Value)

	// Misc nodes

	case *past.ExceptHandler:
		// ExprType Expr
		// Name     Identifier
		// Body     []Stmt
		walk(ast, node.ExprType)
		walkStmts(ast, node.Body)

	case *past.Arguments:
		// Args       []*Arg
		// Vararg     *Arg
		// Kwonlyargs []*Arg
		// KwDefaults []Expr
		// Kwarg      *Arg
		// Defaults   []Expr
		for _, arg := range node.Args {
			walk(ast, arg)
		}
		if node.Vararg != nil {
			walk(ast, node.Vararg)
		}
		for _, arg := range node.Kwonlyargs {
			walk(ast, arg)
		}
		walkExprs(ast, node.KwDefaults)
		if node.Kwarg != nil {
			walk(ast, node.Kwarg)
		}
		walkExprs(ast, node.Defaults)

	case *past.Arg:
		// Arg        Identifier
		// Annotation Expr
		if node.Annotation != nil {
			walk(ast, node.Annotation)
		}

	case *past.Keyword:
		// Arg   Identifier
		// Value Expr
		walk(ast, node.Value)

	case *past.Alias:
		// Name   Identifier
		// AsName Identifier

	case *past.WithItem:
		// ContextExpr  Expr
		// OptionalVars Expr
		walk(ast, node.ContextExpr)
		walk(ast, node.OptionalVars)

	default:
		panic(fmt.Sprintf("Unknown ast node %T, %#v", node, node))
	}
}

func GetParent(ast past.Ast) (parent past.Ast, ok bool) {
	if p, ok := ParentMap[ast]; ok {
		parent = p
	}

	ok = true
	return
}
