package tool

import (
	"fmt"
	"testing"
)

func TestHashed(t *testing.T) {
	a := "ok"
	b := "ok"
	hashedA := GenerateHashOfLength64(a)
	hashedB := GenerateHashOfLength64(b)
	if GenerateHashOfLength64(a) != GenerateHashOfLength64(b) {
		t.Fatalf("error hashed")
	}
	fmt.Printf("hashedA:%v\nhashedB:%v\n", hashedA, hashedB)
}
