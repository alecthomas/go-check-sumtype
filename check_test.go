package gochecksumtype

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

// TestMissingOne tests that we detect a single missing variant.
func TestMissingOne(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A:
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := Run(pkgs)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, []string{"B"}, missingNames(t, errs[0]))
}

// TestMissingTwo tests that we detect a two missing variants.
func TestMissingTwo(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

type C struct {}
func (c *C) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A:
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := Run(pkgs)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, []string{"B", "C"}, missingNames(t, errs[0]))
}

// TestMissingOneWithPanic tests that we detect a single missing variant even
// if we have a trivial default case that panics.
func TestMissingOneWithPanic(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A:
	default:
		panic("unreachable")
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := Run(pkgs)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, []string{"B"}, missingNames(t, errs[0]))
}

// TestNoMissing tests that we correctly detect exhaustive case analysis.
func TestNoMissing(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

type C struct {}
func (c *C) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A, *B, *C:
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := Run(pkgs)
	assert.Equal(t, 0, len(errs))
}

// TestNoMissingDefault tests that even if we have a missing variant, a default
// case should thwart exhaustiveness checking.
func TestNoMissingDefault(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A:
	default:
		println("legit catch all goes here")
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := Run(pkgs)
	assert.Equal(t, 0, len(errs))
}

// TestNotSealed tests that we report an error if one tries to declare a sum
// type with an unsealed interface.
func TestNotSealed(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T interface {}

func main() {}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := Run(pkgs)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "T", errs[0].(unsealedError).Decl.TypeName)
}

// TestNotInterface tests that we report an error if one tries to declare a sum
// type that doesn't correspond to an interface.
func TestNotInterface(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T struct {}

func main() {}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := Run(pkgs)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "T", errs[0].(notInterfaceError).Decl.TypeName)
}

func missingNames(t *testing.T, err error) []string {
	t.Helper()
	ierr, ok := err.(inexhaustiveError)
	assert.True(t, ok, "error was not inexhaustiveError: %T", err)
	return ierr.Names()
}
