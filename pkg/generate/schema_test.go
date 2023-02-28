package generate

import (
	"encoding/json"
	"testing"
)

func TestConvert(t *testing.T) {
	cases := []struct{
		Name string
		Input JSON
		Expected []SchemaItem
	}{
		{
			Name: "Object should become SchemaObject",
			Input: &ObjNode{Items: []*KeyValueNode{
				{"id", "Foo"},
				{"properties", &ObjNode{
					Items: []*KeyValueNode{
						{"someProp", &ObjNode{Items: []*KeyValueNode{
								{"type", "string"},
							}},
						},
					}},
				},
			}},
			Expected: []SchemaItem{
				&SchemaObjectProperty{
					SchemaProperty: SchemaProperty{
						Id: "Foo",
					},
					Properties: []SchemaItem{
						&SchemaStringProperty{
							SchemaProperty: SchemaProperty{
								Name: "someProp",
							}},
					},
				},
			},
		},
		{
			Name: "Nested Object",
			Input: &ObjNode{Items: []*KeyValueNode{
				{"id", "Foo"},
				{"properties", &ObjNode{
					Items: []*KeyValueNode{
						{
							"someProp",
							&ObjNode{Items: []*KeyValueNode{
								{"type", "object"},
								{"properties", &ObjNode{Items: []*KeyValueNode{
									{"nested", &ObjNode{Items: []*KeyValueNode{
										{"type", "string"},
									}}},
								}}},
							}},
						},
					},
				}},
			}},
			Expected: []SchemaItem{
				&SchemaObjectProperty{
					SchemaProperty: SchemaProperty{
						Id: "Foo",
					},
					Properties: []SchemaItem{
						&SchemaObjectProperty{
							SchemaProperty: SchemaProperty{
								Name: "someProp",
							},
							Properties: []SchemaItem{
								&SchemaStringProperty{
									SchemaProperty: SchemaProperty{
										Name: "nested",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T){
			if actual, err := Convert(c.Input); err != nil {
				t.Errorf("Got error; %v", err)
			} else if !ValueEqual(actual, c.Expected) {
				b, _ := json.MarshalIndent(c.Expected, "", " ")
				d, _ := json.MarshalIndent(actual, "", " ")
				t.Errorf("Expected:\n%v\nGot:\n%v", string(b), string(d))
			}
		})
	}
}

func TestMergeNamespaces(t *testing.T) {
	ns1 := &SchemaNamespace{}
	ns1.Name = "foo"
	ns1.Permissions = []string{"perm1", "perm2"}
	ns2 := &SchemaNamespace{}
	ns2.Name = "foo"
	ns2.Permissions = []string{"perm3"}
	typ := &SchemaStringProperty{}
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
