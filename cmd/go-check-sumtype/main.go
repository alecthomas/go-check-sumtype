package main

import (
	"flag"
	"log"
	"os"
	"strings"

	gochecksumtype "github.com/alecthomas/go-check-sumtype"
	"golang.org/x/tools/go/packages"
)

func main() {
	log.SetFlags(0)

	defaultSignifiesExhaustive := flag.Bool(
		"default-signifies-exhaustive",
		true,
		"Presence of \"default\" case in switch statements satisfies exhaustiveness, if all members are not listed.",
	)

	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatalf("Usage: sumtype <packages>\n")
	}
	args := os.Args[flag.NFlag()+1:]

	config := gochecksumtype.Config{
		DefaultSignifiesExhaustive: *defaultSignifiesExhaustive,
	}

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
		log.Fatal(err)
	}
	if errs := gochecksumtype.Run(pkgs, config); len(errs) > 0 {
		var list []string
		for _, err := range errs {
			list = append(list, err.Error())
		}
		log.Fatal(strings.Join(list, "\n"))
	}
}
