package lib

import (
	"fmt"
	"testing"

	"github.com/appscode/go/runtime"
)

func TestParseCsvFile(t *testing.T) {
	res, err := LoadTokenFile(runtime.GOPath() + "/src/github.com/appscode/guard/dist/auth/token.csv")
	if err != nil {
		t.Error(err)
	}
	for _, r := range res {
		fmt.Println(r)
		for _, g := range r.Groups {
			fmt.Printf("%v\n", g)
		}
	}
}
