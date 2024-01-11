package testdata

//sumtype:decl
type Sum interface{ sum() }

type A struct{}

func (A) sum() {}

type B struct{}

func (B) sum() {}

type C[T any] struct{}

func (C[T]) sum() {}

func SumSwitch(x Sum) {
	switch x.(type) {
	case A:
	case B:
	}
}
