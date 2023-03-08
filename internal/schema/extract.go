package schema

type Pieces struct {
	Structs   []*Object
	Enums     []*Enum
	Functions []*Function
	Events    []*Event
}

func Extract(item Item) *Pieces {
	p := &Pieces{}
	p.extract(item)
	return p
}

func (this *Pieces) extract(item Item) {
	switch item := item.(type) {
	case *Namespace:
		for _, thing := range item.Properties {
			this.extract(thing)
		}
		for _, thing := range item.Types {
			this.extract(thing)
		}
		for _, thing := range item.Functions {
			this.extract(thing)
		}
		for _, thing := range item.Events {
			this.extract(thing)
		}
	case *Object:
		for _, prop := range item.Properties {
			this.extract(prop)
		}
		for _, thing := range item.Functions {
			this.extract(thing)
		}
		for _, thing := range item.Events {
			this.extract(thing)
		}
		this.Structs = append(this.Structs, item)
	case *Enum:
		this.Enums = append(this.Enums, item)
	case *Function:
		for _, param := range item.Parameters {
			this.extract(param)
		}
		this.Functions = append(this.Functions, item)
	case *Event:
		for _, param := range item.Parameters {
			this.extract(param)
		}
		for _, param := range item.ExtraParameters {
			this.extract(param)
		}
		for _, param := range item.Filters {
			this.extract(param)
		}
		this.Events = append(this.Events, item)
	}
}
