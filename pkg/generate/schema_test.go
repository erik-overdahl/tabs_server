package generate

import (
	"testing"
)

func TestMergeNamespaces(t *testing.T) {
	ns1 := &SchemaNamespace{SchemaProperty: &SchemaProperty{}}
	ns1.Name = "foo"
	ns1.Permissions = []string{"perm1", "perm2"}
	ns2 := &SchemaNamespace{SchemaProperty: &SchemaProperty{}}
	ns2.Name = "foo"
	ns2.Permissions = []string{"perm3"}
	typ := &SchemaStringProperty{SchemaProperty: &SchemaProperty{}}
	typ.Id = "String"
	ns2.Types = []SchemaItem{typ}
	spaces := []*SchemaNamespace{ns1, ns2}

	spaces = MergeNamespaces(spaces)

	if len(spaces) != 1 {
		t.Fatalf("Expected 1 ns, got %d", len(spaces))
	}
	if len(ns1.Permissions) != 3 {
		t.Fatalf("Expected 3 permissions, got %d", len(ns1.Permissions))
	}
	if len(ns1.Types) != 1 {
		t.Fatalf("Expected 1 permissions, got %d", len(ns1.Types))
	}
}
