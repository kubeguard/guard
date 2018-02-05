package lib

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestParseCsvFile(t *testing.T) {
	res, err := ParseCsvFile("/home/ac/go/src/github.com/appscode/guard/dist/auth/token.csv")
	if err != nil {
		t.Error(err)
	}
	for _, r := range res {
		fmt.Println(r)
		for _, g := range r.groups {
			fmt.Printf("%v\n", g)
		}
	}
}

func TestValidateTokenCvsFile(t *testing.T) {
	test1 := fmt.Sprintln("token1 ,user1 ,uid1,\"group1 ,group2 ,group3\"\ntoken2,user1,uid1,")
	tWriteFile(tokenAuthCsvFile, []byte(test1))
	err := ValidateTokenCvsFile(tokenAuthCsvFile)
	if err != nil {
		t.Error(err)
	}
}

func tWriteFile(file string, data []byte) error {
	return ioutil.WriteFile(file, data, 666)
}
