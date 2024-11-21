package gochecksumtype

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func setupPackages(t *testing.T, code string) []*packages.Package {
	srcPath := filepath.Join(t.TempDir(), "src.go")
	if err := os.WriteFile(srcPath, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	pkgs, err := tycheckAll([]string{srcPath})
	if err != nil {
		t.Fatal(err)
	}
	return pkgs
}

func tycheckAll(args []string) ([]*packages.Package, error) {
	conf := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedTypesSizes |
			packages.NeedImports | packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles,
		// Unfortunately, it appears including the test packages in
		// this lint makes it difficult to do exhaustiveness checking.
		// Namely, it appears that compiling the test version of a
		// package introduces distinct types from the normal version
		// of the package, which will always result in inexhaustive
		// errors whenever a package both defines a sum type and has
		// tests. (Specifically, using `package name`. Using `package
		// name_test` is OK.)
		//
		// It's not clear what the best way to fix this is. :-(
		Tests: false,
	}
	pkgs, err := packages.Load(conf, args...)
	if err != nil {
		return nil, err
	}
	return pkgs, nil
}
