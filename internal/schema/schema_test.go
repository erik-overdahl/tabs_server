package schema

import (
	"testing"

	"github.com/erik-overdahl/tabs_server/internal/util"
)

func TestIdenticalItemsWithDiffParentsAreEqual(t *testing.T) {
	o1 := &Object{Property: Property{parent: &Namespace{}},
		Properties: []Item{
			&String{Property: Property{Name: "foo"}},
		},
	}
	o2 := &Object{Property: Property{parent: &Object{}},
		Properties: []Item{
			&String{Property: Property{Name: "foo"}},
		},
	}
	if !util.ValueEqual(o1, o2) {
		t.Fail()
	}
}
