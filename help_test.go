package gochecksumtype

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func setupPackages(t *testing.T, code string) (string, []*packages.Package) {
	tmpdir, err := ioutil.TempDir("", "go-test-sumtype-")
	if err != nil {
		t.Fatal(err)
	}
	srcPath := filepath.Join(tmpdir, "src.go")
	if err := ioutil.WriteFile(srcPath, []byte(code), 0666); err != nil {
		t.Fatal(err)
	}
	pkgs, err := tycheckAll([]string{srcPath})
	if err != nil {
		t.Fatal(err)
	}
	return tmpdir, pkgs
}

func teardownPackage(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
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
