package years

import (
	"fmt"
	"testing"
)

func TestVoyagerPrepare(t *testing.T) {
	v := NewVoyager("2006/Jan/2006-01-02.txt")
	err := v.Prepare("internal/testdata/calendar")
	if err != nil {
		t.Fatalf("voyager failed: %s", err)
	}

	// todo: actual asserts here
	fmt.Println(v.layout)
}
