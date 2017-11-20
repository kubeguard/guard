package lib

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestToken(t *testing.T) {
	resp, e := checkAppscode("qacode", "cli-7ynsaphij5qjdlt5a3rtkrucjwor")
	a, er := json.Marshal(resp)
	fmt.Println(string(a), er, e)
}
