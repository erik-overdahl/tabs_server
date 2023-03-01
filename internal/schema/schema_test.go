package schema

import (
	"encoding/json"
	"testing"

	"github.com/erik-overdahl/tabs_server/internal/util"
	ojson "github.com/erik-overdahl/tabs_server/internal/json"
)

func TestConvert(t *testing.T) {
	cases := []struct{
		Name string
		Input ojson.JSON
		Expected []Item
	}{
		{
			Name: "Object should become SchemaObject",
			Input: &ojson.Object{Items: []*ojson.KeyValue{
				{"id", "Foo"},
				{"properties", &ojson.Object{
					Items: []*ojson.KeyValue{
						{"someProp", &ojson.Object{Items: []*ojson.KeyValue{
								{"type", "string"},
							}},
						},
					}},
				},
			}},
			Expected: []Item{
				&Object{
					Property: Property{
						Id: "Foo",
					},
					Properties: []Item{
						&String{
							Property: Property{
								Name: "someProp",
							}},
					},
				},
			},
		},
		{
			Name: "Nested Object",
			Input: &ojson.Object{Items: []*ojson.KeyValue{
				{"id", "Foo"},
				{"properties", &ojson.Object{
					Items: []*ojson.KeyValue{
						{
							"someProp",
							&ojson.Object{Items: []*ojson.KeyValue{
								{"type", "object"},
								{"properties", &ojson.Object{Items: []*ojson.KeyValue{
									{"nested", &ojson.Object{Items: []*ojson.KeyValue{
										{"type", "string"},
									}}},
								}}},
							}},
						},
					},
				}},
			}},
			Expected: []Item{
				&Object{
					Property: Property{
						Id: "Foo",
					},
					Properties: []Item{
						&Object{
							Property: Property{
								Name: "someProp",
							},
							Properties: []Item{
								&String{
									Property: Property{
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
			} else if !util.ValueEqual(actual, c.Expected) {
				b, _ := json.MarshalIndent(c.Expected, "", " ")
				d, _ := json.MarshalIndent(actual, "", " ")
				t.Errorf("Expected:\n%v\nGot:\n%v", string(b), string(d))
			}
		})
	}
}

func TestMergeNamespaces(t *testing.T) {
	ns1 := &Namespace{}
	ns1.Name = "foo"
	ns1.Permissions = []string{"perm1", "perm2"}
	ns2 := &Namespace{}
	ns2.Name = "foo"
	ns2.Permissions = []string{"perm3"}
	typ := &String{}
	typ.Id = "String"
	ns2.Types = []Item{typ}
	spaces := []*Namespace{ns1, ns2}

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
