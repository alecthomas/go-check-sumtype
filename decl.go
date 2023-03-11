package gochecksumtype

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

// sumTypeDecl is a declaration of a sum type in a Go source file.
type sumTypeDecl struct {
	// The package path that contains this decl.
	Package *packages.Package
	// The type named by this decl.
	TypeName string
	// The file path where this declaration was found.
	Path string
	// The line number where this declaration was found.
	Line int
}

// Location returns a short string describing where this declaration was found.
func (d sumTypeDecl) Location() string {
	return fmt.Sprintf("%s:%d", d.Path, d.Line)
}

// findSumTypeDecls searches every package given for sum type declarations of
// the form `sumtype:decl`.
func findSumTypeDecls(pkgs []*packages.Package) ([]sumTypeDecl, error) {
	var decls []sumTypeDecl
	var retErr error
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(node ast.Node) bool {
				if node == nil {
					return true
				}
				decl, ok := node.(*ast.GenDecl)
				if !ok || decl.Doc == nil {
					return true
				}
				var tspec *ast.TypeSpec
				for _, spec := range decl.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					tspec = ts
				}
				for _, line := range decl.Doc.List {
					if !strings.HasPrefix(line.Text, "//sumtype:decl") {
						continue
					}
					pos := pkg.Fset.Position(decl.Pos())
					if tspec == nil {
						retErr = notFoundError{Decl: sumTypeDecl{Package: pkg, Path: pos.Filename, Line: pos.Line}}
						return false
					}
					decl := sumTypeDecl{
						Package:  pkg,
						TypeName: tspec.Name.Name,
						Path:     pos.Filename,
						Line:     pos.Line,
					}
					decls = append(decls, decl)
					break
				}
				return true
			})
		}
	}
	return decls, retErr
}
