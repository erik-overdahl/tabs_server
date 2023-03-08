package schema

import (
	"encoding/json"
	"testing"

	"github.com/erik-overdahl/tabs_server/internal/util"
	ojson "github.com/erik-overdahl/tabs_server/internal/json"
)

func makeTest(input ojson.JSON, expected []Item) func(*testing.T) {
	return func(t *testing.T) {
		if actual, err := Convert(input); err != nil {
			t.Errorf("Test failed with ERR: %v", err)
		} else if !util.ValueEqual(actual, expected) {
			b, _ := json.MarshalIndent(expected, "", " ")
			d, _ := json.MarshalIndent(actual, "", " ")
			t.Errorf("Expected:\n%v\nGot:\n%v", string(b), string(d))
		}
	}
}

func TestEnum(t *testing.T) {
	cases := []struct{
		Name string
		Input ojson.JSON
		Expected []Item
	}{
		{
			Name: "Enum",
			Input: &ojson.Object{[]*ojson.KeyValue{
				{Key: "id", Value: "IsEnum"},
				{Key: "type", Value: "string"},
				{Key: "enum", Value: &ojson.List{Items: []any{
					"foo", "bar",
				}}},
			}},
			Expected: []Item{
				&Enum{Property: Property{
					Id: "IsEnum",
				},
					Enum: []EnumValue{
						{Name: "foo"},
						{Name: "bar"},
					},
				},
			},
		},
		{
			Name: "Not Enum",
			Input: &ojson.Object{Items: []*ojson.KeyValue{
				{Key: "id", Value: "IsNotEnum"},
				{Key: "type", Value: "string"},
			}},
			Expected: []Item{
				&String{Property: Property{
					Id: "IsNotEnum",
				},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, makeTest(c.Input, c.Expected))
	}
}

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
				{Key: "id", Value: "Foo"},
				{Key: "properties", Value: &ojson.Object{
					Items: []*ojson.KeyValue{
						{
							Key: "someProp",
							Value: &ojson.Object{Items: []*ojson.KeyValue{
								{Key: "type", Value: "object"},
								{Key: "properties", Value: &ojson.Object{Items: []*ojson.KeyValue{
									{Key: "nested", Value: &ojson.Object{Items: []*ojson.KeyValue{
										{Key: "type", Value: "string"},
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
		{
			Name: "Funcs and Events",
			Input: &ojson.Object{Items: []*ojson.KeyValue{
				{Key: "namespace", Value: "foospace"},
				{Key: "functions", Value: &ojson.List{Items: []any{
					&ojson.Object{Items: []*ojson.KeyValue{
						{Key: "type", Value: "function"},
						{Key: "name", Value: "fooFunc"},
					}},
				}}},
				{Key: "events", Value: &ojson.List{Items: []any{
					&ojson.Object{Items: []*ojson.KeyValue{
						{Key: "type", Value: "function"},
						{Key: "name", Value: "onFoo"},
					}},
				}}}},
			},
			Expected: []Item{&Namespace{Property: Property{Name: "foospace"},
				Functions: []*Function{
					&Function{Property: Property{Name: "fooFunc"}},
				},
				Events: []*Event{
					&Event{Property: Property{Name: "onFoo"}},
				}},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, makeTest(c.Input, c.Expected))
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
