package schema

import (
	"encoding/json"
	"testing"

	"github.com/erik-overdahl/tabs_server/internal/util"
)

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
			Expected: &Pieces{Structs: []*Object{
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
		t.Run(c.Name, func(t *testing.T) {
			actual := Extract(c.Input)
			if !util.ValueEqual(actual, c.Expected) {
				b, _ := json.MarshalIndent(c.Expected, "", " ")
				d, _ := json.MarshalIndent(actual, "", " ")
				t.Errorf("Expected:\n%v\nGot:\n%v", string(b), string(d))
			}
		})
	}
}
