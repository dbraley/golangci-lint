package linters

import (
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/ast/astutil"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var ContextAnalyzer = &analysis.Analyzer{
	Name: "context",
	Doc: ("prohibits context.Background in most cases, as well as checking " +
		"that Context is always the first parameter and is named ctx."),
	Run: _runContext,
}

// _lintContextBackground lints for context.Background() calls.
func _lintContextBackground(
	report func(analysis.Diagnostic),
	file *ast.File,
	typesInfo *types.Info,
) {
	// TODO: If we end up with a ton of Inspect-based analyzers,
	// use x/tools/go/analysis/passes/inspect to make them faster.
	ast.Inspect(file, func(node ast.Node) bool {
		if NameOf(node, typesInfo) == "context.Background" {
			report(analysis.Diagnostic{
				Pos:     node.Pos(),
				Message: "do not use context.Background() outside tests",
			})
			// The children of the node representing context.Background aren't
			// independently interesting, so we don't traverse them.  (This
			// also avoids duplicate errors, since the "Background" also ends
			// up referring to context.Background.
			return false
		}
		funcDecl, ok := node.(*ast.FuncDecl)
		// In init and main, you don't usually have a more meaningful
		// context.
		if ok && funcDecl.Recv == nil &&
			(funcDecl.Name.Name == "init" || funcDecl.Name.Name == "main") {
			return false // do not traverse this node's children
		}
		return true
	})
}

// _lintContextParameter lints for incorrect context parameters.
func _lintContextParameter(
	report func(analysis.Diagnostic),
	file *ast.File,
	typesInfo *types.Info,
) {
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		for i, param := range funcDecl.Type.Params.List {
			if NameOf(param.Type, typesInfo) == "context.Context" {
				// the parameter type is context.Context.
				if i != 0 || len(param.Names) > 1 {
					report(analysis.Diagnostic{
						Pos:     param.Pos(),
						Message: "Context should be the first parameter",
					})
				}
				name := param.Names[0].Name
				if name != "ctx" && name != "_" {
					// This duplicates the check in golint, but we may as well
					// do it here too.
					report(analysis.Diagnostic{
						Pos:     param.Pos(),
						Message: "Context parameter should be called 'ctx'",
					})
				}
			}
		}
	}
}

func _runContext(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		_lintContextParameter(pass.Report, file, pass.TypesInfo)

		filename := pass.Fset.File(file.Pos()).Name()
		if strings.HasSuffix(filename, "_test.go") {
			// We allow tests to use context.Background().
			continue
		}
		_lintContextBackground(pass.Report, file, pass.TypesInfo)
	}
	return nil, nil
}

// NameOf takes a node and returns the name of the symbol to which it refers,
// if any, in the form package/path.UnqualifiedName.
//
// This will return a name for functions and types, and not necessarily other
// nodes.  If it can't determine the name, it returns "".
func NameOf(node ast.Node, typesInfo *types.Info) string {
	obj := objectFor(node, typesInfo)
	if obj == nil {
		return ""
	}
	switch obj := obj.(type) {
	case *types.TypeName:
		pkg := obj.Pkg()
		if pkg == nil {
			return obj.Name()
		}
		return pkg.Path() + "." + obj.Name()
	case *types.Func:
		return obj.FullName()
	// TODO(benkraft): In principle we should be able to do this for a
	// const/var too, but I'm not sure how.  We don't need it yet.
	default:
		return ""
	}
}

// objectFor takes an AST node, and returns the corresponding types.Object, if
// there is one.
func objectFor(node ast.Node, typesInfo *types.Info) types.Object {
	// This is mostly cribbed from
	// https://github.com/golang/tools/blob/master/go/types/typeutil/callee.go#L16
	// which does a very similar thing, but only for functions.
	exprNode, ok := node.(ast.Expr)
	if !ok {
		return nil
	}
	exprNode = astutil.Unparen(exprNode)

	switch node := exprNode.(type) {
	case *ast.Ident:
		return typesInfo.Uses[node]
	case *ast.SelectorExpr:
		if sel, ok := typesInfo.Selections[node]; ok {
			return sel.Obj()
		}
		return typesInfo.Uses[node.Sel]
	default:
		return nil
	}
}
