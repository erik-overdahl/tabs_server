package schema

import "github.com/erik-overdahl/tabs_server/internal/util"

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
		this.addStruct(item)
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

/*
 * If another struct with the same name and same properties exists, we
 * can skip. If same name but different properties, we set the Id to be
 * parent name + struct name for BOTH structs.
 */
func (this *Pieces) addStruct(s *Object) {
	for _, other := range this.Structs {
		sameProps := util.ValueEqual(s.Properties, other.Properties)
		// already have this struct?
		if s.Name == other.Name && sameProps {
			return
		// names equal but structs differ?
		} else if s.Name == other.Name {
			s.Id = s.parent.Base().Name + util.Exportable(s.Name)
			other.Id = other.parent.Base().Name + util.Exportable(other.Name)
		}
	}
	this.Structs = append(this.Structs, s)
}
