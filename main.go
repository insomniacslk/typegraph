package main

// thanks to https://medium.com/justforfunc/understanding-go-programs-with-go-parser-c4e88a6edb87

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"unicode"
)

var (
	flagPublicOnly = flag.Bool("public", false, "Only consider public identifiers")
)

// Something is just here to prove the point. So you can call this program on
// this source file.
type Something struct {
	A int
	B string
}

var _ Something

type edge struct {
	Left, Right string
	Label       string
}

func exprName(e ast.Expr) string {
	switch e.(type) {
	case *ast.Ident:
		return e.(*ast.Ident).Name
	case *ast.StarExpr:
		return exprName(e.(*ast.StarExpr).X)
	case *ast.ArrayType:
		return exprName(e.(*ast.ArrayType).Elt)
	case *ast.MapType:
		mt := e.(*ast.MapType)
		return fmt.Sprintf("map[%s]%s", exprName(mt.Key), exprName(mt.Value))
	case *ast.SelectorExpr:
		s := e.(*ast.SelectorExpr)
		return fmt.Sprintf("%s.%s", exprName(s.X), s.Sel.Name)
	case *ast.ChanType:
		ch := e.(*ast.ChanType)
		var chtype string
		if ch.Dir == ast.SEND {
			chtype = "chan<-"
		} else if ch.Dir == ast.RECV {
			chtype = "<-chan"
		} else {
			chtype = "chan"
		}
		return fmt.Sprintf("%s %s", chtype, exprName(ch.Value))
	case *ast.StructType:
		return "struct{} (unknown name)"
	case *ast.InterfaceType:
		return "interface (unknown name)"
	default:
		return fmt.Sprintf("unhandled expr (%T)", e)
	}
}

func exprNameAndType(n ast.Expr) (string, string) {
	switch t := n.(type) {
	case *ast.Ident:
		return n.(*ast.Ident).Name, "value"
	case *ast.ArrayType:
		elt := n.(*ast.ArrayType).Elt
		return exprName(elt), "array"
	case *ast.MapType:
		return exprName(n.(*ast.MapType)), "map"
	case *ast.StarExpr:
		return exprName(n.(*ast.StarExpr).X), "ptr"
	case *ast.SelectorExpr:
		return exprName(n.(*ast.SelectorExpr)), "selector"
	case *ast.ChanType:
		return exprName(n.(*ast.ChanType)), "chan"
	case *ast.StructType:
		return exprName(n.(*ast.StructType)), "struct"
	case *ast.InterfaceType:
		return exprName(n.(*ast.InterfaceType)), "interface"
	default:
		return fmt.Sprintf("unhandled (%T)", t), "unknown"
	}
}

type visitor []edge

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	switch t := n.(type) {
	case *ast.TypeSpec:
		if _, ok := t.Type.(*ast.StructType); !ok {
			return v
		}
		fields := t.Type.(*ast.StructType).Fields
		for _, f := range fields.List {
			typeName, label := exprNameAndType(f.Type)
			*v = append(*v, edge{
				Left:  t.Name.Name,
				Right: typeName,
				Label: label,
			})
		}
	}
	return v
}

type dotConfig struct {
	// only consider public identifiers
	PublicOnly bool
}

// public returns true if both Left and Right are public identifiers, i.e. their
// first character is upper-case.
func public(e edge) bool {
	if len(e.Left) == 0 || len(e.Right) == 0 {
		return false
	}
	return unicode.IsUpper(rune(e.Left[0])) && unicode.IsUpper(rune(e.Right[0]))
}

// ToDot converts a list of edges to a GraphViz dot file.
func ToDot(edges []edge, conf dotConfig) string {
	ret := "DiGraph {\n"
	for _, edge := range edges {
		if conf.PublicOnly && !public(edge) {
			continue
		}
		var label string
		if edge.Label != "" {
			label = fmt.Sprintf(" [label=\"%s\"]", edge.Label)
		}
		ret += fmt.Sprintf("    \"%s\" -> \"%s\"%s\n", edge.Left, edge.Right, label)
	}
	return ret + "}\n"
}

func visitFiles(fileNames []string) []edge {
	var edges []edge
	for _, fname := range fileNames {
		fset := token.NewFileSet()

		code, err := parser.ParseFile(fset, fname, nil, parser.AllErrors)
		if err != nil {
			log.Fatal(err)
		}
		var v visitor
		ast.Walk(&v, code)
		edges = append(edges, v...)
	}
	return edges
}

func main() {
	flag.Parse()
	fnames := flag.Args()
	if len(fnames) == 0 {
		fmt.Println("Need at least one file name")
		os.Exit(1)
	}

	dc := dotConfig{
		PublicOnly: *flagPublicOnly,
	}
	fmt.Println(ToDot(visitFiles(fnames), dc))
}
