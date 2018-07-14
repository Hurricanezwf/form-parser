package formparser

import (
	"fmt"
	"reflect"
	"testing"
)

var h = Hello{
	A: 1,
	B: "BB",
	C: 2,
	D: 3.14,
	E: []int{2, 0, 32},
	F: Info{CPU: StringPtr("1核")},
	G: true,
	H: []*Info{
		&Info{CPU: StringPtr("2核")},
		&Info{CPU: StringPtr("3核")},
		&Info{CPU: StringPtr("4核")},
	},
	I: map[string]*string{
		"m1": StringPtr("m1"),
		"m2": StringPtr("m2"),
	},
	J: []byte("Golang"),
}

func TestParse(t *testing.T) {
	p := New("a", "-")
	_, err := p.parse(reflect.ValueOf(h))
	if err != nil {
		t.Fatalf(err.Error())
	}
	p.Debug(reflect.ValueOf(h))
}

func BenchmarkParse(b *testing.B) {
	p := New("a", "-")
	for i := 0; i < b.N; i++ {
		_, err := p.parse(reflect.ValueOf(h))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	p.Debug(reflect.ValueOf(h))
}

type Hello struct {
	A int                `a:"a"`
	B string             `a:"b"`
	C int64              `a:"c"`
	D float64            `a:"d"`
	E []int              `a:"e"`
	F Info               `a:"f"`
	G bool               `a:"g"`
	H []*Info            `a:"h"`
	I map[string]*string `a:"i"`
	J []byte             `a:"j"`
}

type Info struct {
	CPU *string `a:"cpu"`
}
