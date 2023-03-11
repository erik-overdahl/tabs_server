package schema

import (
	"encoding/json"
	"testing"

	"github.com/erik-overdahl/tabs_server/internal/util"
)

func makeExtractTest(input Item, expected *Pieces) func(t *testing.T) {
	return func(t *testing.T) {
		actual := Extract(input)
		if !util.ValueEqual(actual, expected) {
			b, _ := json.MarshalIndent(expected, "", " ")
			d, _ := json.MarshalIndent(actual, "", " ")
			t.Errorf("Expected:\n%v\nGot:\n%v", string(b), string(d))
		}
	}
}

func TestExtract(t *testing.T) {
	cases := []struct {
		Name     string
		Input    Item
		Expected *Pieces
	}{
		{
			Name: "Anon object property should become struct",
			Input: &Object{
				Property: Property{
					Id: "Foo",
				},
				Properties: []Item{
					&Object{Property: Property{
						Name: "objProp",
					},
						Properties: []Item{
							&String{Property: Property{
								Name: "stringProp",
							}},
						}},
				},
			},
			Expected: &Pieces{
				Structs: []*Object{
					{Property: Property{
						Name: "objProp",
					},
						Properties: []Item{
							&String{Property: Property{
								Name: "stringProp",
							}},
						},
					},
					{
						Property: Property{
							Id: "Foo",
						},
						Properties: []Item{
							&Object{Property: Property{
								Name: "objProp",
							},
								Properties: []Item{
									&String{Property: Property{
										Name: "stringProp",
									}},
								}},
						},
					},
				},
			},
		},
		{
			Name: "Anon func param should become struct",
			Input: &Function{Property: Property{Name: "foo"},
				Parameters: []Item{
					&Object{Property: Property{Name: "details"},
						Properties: []Item{
							&String{Property: Property{Name: "stringProp"}},
						},
					},
				},
			},
			Expected: &Pieces{
				Functions: []*Function{
					{Property: Property{Name: "foo"},
						Parameters: []Item{
							&Object{Property: Property{Name: "details"},
								Properties: []Item{
									&String{Property: Property{Name: "stringProp"}},
								},
							},
						},
					}},
				Structs: []*Object{
					{Property: Property{Name: "details"},
						Properties: []Item{
							&String{Property: Property{Name: "stringProp"}},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, makeExtractTest(c.Input, c.Expected))
	}
}

func TestSameStructExtract(t *testing.T) {
	input := &Namespace{
		Functions: []*Function{
			{
				Property: Property{Name: "fooFunc"},
				Parameters: []Item{
					&Object{
						Property: Property{Name: "details"},
						Properties: []Item{
							&Int{Property: Property{Name: "tabId"}},
						},
					},
				},
			},
			{
				Property: Property{Name: "barFunc"},
				Parameters: []Item{
					&Object{
						Property: Property{Name: "details"},
						Properties: []Item{
							&Int{Property: Property{Name: "tabId"}},
						},
					},
				},
			},
		},
	}
	expected := &Pieces{
		Namespace: input,
		Structs: []*Object{
			{
				Property: Property{Name: "details"},
				Properties: []Item{
					&Int{Property: Property{Name: "tabId"}},
				},
			},
		},
		Functions: []*Function{
			{
				Property: Property{Name: "fooFunc"},
				Parameters: []Item{
					&Object{
						Property: Property{Name: "details"},
						Properties: []Item{
							&Int{Property: Property{Name: "tabId"}},
						},
					},
				},
			},
			{
				Property: Property{Name: "barFunc"},
				Parameters: []Item{
					&Object{
						Property: Property{Name: "details"},
						Properties: []Item{
							&Int{Property: Property{Name: "tabId"}},
						},
					},
				},
			},
		},
	}
	makeExtractTest(input, expected)(t)
}

func TestDifferentStructSameName(t *testing.T) {
	f1 := &Function{Property: Property{Name: "fooFunc"}}
	s1 := &Object{
		Property: Property{
			Name:   "details",
			parent: f1,
		},
		Properties: []Item{
			&Int{Property: Property{Name: "tabId"}},
		},
	}
	f1.Parameters = []Item{s1}

	f2 := &Function{Property: Property{Name: "barFunc"}}
	s2 := &Object{
		Property: Property{
			Name:   "details",
			parent: f2,
		},
		Properties: []Item{
			&String{Property: Property{Name: "barName"}},
		},
	}
	f2.Parameters = []Item{s2}

	input := &Namespace{
		Functions: []*Function{f1, f2},
	}
	expected := &Pieces{
		Namespace: input,
		Structs: []*Object{
			{
				Property: Property{
					parent: f1,
					Name:   "details",
					Id:     "fooFuncDetails",
				},
				Properties: []Item{
					&Int{Property: Property{Name: "tabId"}},
				},
			},
			{
				Property: Property{
					parent: f2,
					Name:   "details",
					Id:     "barFuncDetails",
				},
				Properties: []Item{
					&String{Property: Property{Name: "barName"}},
				},
			},
		},
		Functions: []*Function{
			{
				Property: Property{Name: "fooFunc"},
				Parameters: []Item{
					&Object{
						Property: Property{
							Name: "details",
							Id:   "fooFuncDetails",
						},
						Properties: []Item{
							&Int{Property: Property{Name: "tabId"}},
						},
					},
				},
			},
			{
				Property: Property{Name: "barFunc"},
				Parameters: []Item{
					&Object{
						Property: Property{
							Name: "details",
							Id:   "barFuncDetails",
						},
						Properties: []Item{
							&String{Property: Property{Name: "barName"}},
						},
					},
				},
			},
		},
	}
	makeExtractTest(input, expected)(t)
}
