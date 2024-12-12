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
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{DefaultSignifiesExhaustive: true})
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, []string{"B"}, missingNames(t, errs[0]))
}

// TestMissingTwo tests that we detect two missing variants.
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
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{DefaultSignifiesExhaustive: true})
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
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{DefaultSignifiesExhaustive: true})
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
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{DefaultSignifiesExhaustive: true})
	assert.Equal(t, 0, len(errs))
}

// TestNoMissingDefaultWithDefaultSignifiesExhaustive tests that even if we have a missing variant, a default
// case should thwart exhaustiveness checking when Config.DefaultSignifiesExhaustive is true.
func TestNoMissingDefaultWithDefaultSignifiesExhaustive(t *testing.T) {
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
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{DefaultSignifiesExhaustive: true})
	assert.Equal(t, 0, len(errs))
}

// TestNoMissingDefaultAndDefaultDoesNotSignifiesExhaustive tests that even if we have a missing variant, a default
// case should thwart exhaustiveness checking when Config.DefaultSignifiesExhaustive is false.
func TestNoMissingDefaultAndDefaultDoesNotSignifiesExhaustive(t *testing.T) {
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
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{DefaultSignifiesExhaustive: false})
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, []string{"B"}, missingNames(t, errs[0]))
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
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{DefaultSignifiesExhaustive: true})
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
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{DefaultSignifiesExhaustive: true})
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "T", errs[0].(notInterfaceError).Decl.TypeName)
}

// TestSubTypeInSwitch tests that if a shared interface is declared in the switch
// statement, we don't report an error if structs that implement the interface are not explicitly
// declared in the switch statement.
func TestSubTypeInSwitch(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T1 interface { sealed1() }
type T2 interface { 
  T1
  sealed2()
}


type A struct {}
func (a *A) sealed1() {}

type B struct {}
func (b *B) sealed1() {}
func (b *B) sealed2() {}

type C struct {}
func (c *C) sealed1() {}
func (c *C) sealed2() {}

func main() {
	switch T1(nil).(type) {
	case *A:
	case T2:
	}
}
`
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{IncludeSharedInterfaces: true})
	assert.Equal(t, 0, len(errs))
}

// TestAllLeavesInSwitch tests that we do not report an error if a switch statement
// covers all leaves of the sum type, even if any SubTypes are not explicitly covered
func TestAllLeavesInSwitch(t *testing.T) {
	code := `
package gochecksumtype

//sumtype:decl
type T1 interface { sealed1() }
type T2 interface { 
  T1
  sealed2()
}


type A struct {}
func (a *A) sealed1() {}

type B struct {}
func (b *B) sealed1() {}
func (b *B) sealed2() {}

type C struct {}
func (c *C) sealed1() {}
func (c *C) sealed2() {}

func main() {
	switch T1(nil).(type) {
	case *A:
	case *B:
	case *C:
	}
}
`
	pkgs := setupPackages(t, code)

	errs := Run(pkgs, Config{})
	assert.Equal(t, 0, len(errs))
}

func missingNames(t *testing.T, err error) []string {
	t.Helper()
	ierr, ok := err.(inexhaustiveError)
	assert.True(t, ok, "error was not inexhaustiveError: %T", err)
	return ierr.Names()
}
