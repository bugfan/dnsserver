package nameserver

import (
	"fmt"
	"testing"
)

func TestWildcardNames(t *testing.T) {
	res := wildcardNames("sub1.sub2.example.com.", "example.com.")
	fmt.Printf("%#v\n", res)

}
