package service

import (
	"reflect"

	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/servicesdk/annotation"
)

type ResourceTypeBuilder interface {
	AddRelationship(name, to, kind, cardinality, reverseName string, keys []string)
	ImmutableAttributes(names ...string)
	ProvidedAttributes(names ...string)
	CustomData(customId, key string, value interface{})
	Tags(tags map[string]string)
	Build(goType interface{}) px.AnnotatedType
}

type rtBuilder struct {
	ctx            px.Context
	relationships  []*types.HashEntry
	immutableAttrs []string
	providedAttrs  []string
	tags           map[string]string
	customData     map[string]map[string]interface{}
}

func (rb *rtBuilder) AddRelationship(name, to, kind, cardinality, reverseName string, keys []string) {
	ln := 4
	if reverseName != `` {
		ln++
	}
	es := make([]*types.HashEntry, ln)
	es[0] = types.WrapHashEntry2(`type`, types.NewTypeReferenceType(to))
	es[1] = types.WrapHashEntry2(`kind`, types.WrapString(kind))
	es[2] = types.WrapHashEntry2(`cardinality`, types.WrapString(cardinality))
	es[3] = types.WrapHashEntry2(`keys`, types.WrapStrings(keys))
	if reverseName != `` {
		es[4] = types.WrapHashEntry2(`reverseName`, types.WrapString(reverseName))
	}
	rb.relationships = append(rb.relationships, types.WrapHashEntry2(name, types.WrapHash(es)))
}

func (rb *rtBuilder) ImmutableAttributes(names ...string) {
	if rb.immutableAttrs == nil {
		rb.immutableAttrs = names
	} else {
		rb.immutableAttrs = append(rb.immutableAttrs, names...)
	}
}

func (rb *rtBuilder) CustomData(customId, key string, value interface{}) {
	if rb.customData == nil {
		rb.customData = make(map[string]map[string]interface{}, 0)
	}
	if cd, ok := rb.customData[customId]; ok {
		cd[key] = value
	} else {
		rb.customData[customId] = map[string]interface{}{key: value}
	}
}

func (rb *rtBuilder) ProvidedAttributes(names ...string) {
	if rb.providedAttrs == nil {
		rb.providedAttrs = names
	} else {
		rb.providedAttrs = append(rb.providedAttrs, names...)
	}
}

func (rb *rtBuilder) Tags(tags map[string]string) {
	if rb.tags == nil {
		rb.tags = tags
	} else {
		for k, v := range tags {
			rb.tags[k] = v
		}
	}
}

func (rb *rtBuilder) Build(goType interface{}) px.AnnotatedType {
	var rt reflect.Type
	switch goType := goType.(type) {
	case reflect.Type:
		rt = goType
	case reflect.Value:
		rt = goType.Type()
	default:
		rt = reflect.TypeOf(goType)
	}

	ae := make([]*types.HashEntry, 0)
	if rb.immutableAttrs != nil || rb.providedAttrs != nil || rb.relationships != nil {
		as := make([]*types.HashEntry, 0, 3)
		if rb.immutableAttrs != nil {
			as = append(as, types.WrapHashEntry2(`immutableAttributes`, types.WrapStrings(rb.immutableAttrs)))
		}
		if rb.providedAttrs != nil {
			as = append(as, types.WrapHashEntry2(`providedAttributes`, types.WrapStrings(rb.providedAttrs)))
		}
		if rb.relationships != nil {
			as = append(as, types.WrapHashEntry2(`relationships`, types.WrapHash(rb.relationships)))
		}
		ae = append(ae, types.WrapHashEntry(annotation.ResourceType, types.WrapHash(as)))
	}
	if rb.customData != nil {
		ae = append(ae, types.WrapHashEntry(annotation.CustomDataType, px.Wrap(rb.ctx, rb.customData)))
	}
	var annotations px.OrderedMap
	if len(ae) > 0 {
		annotations = types.WrapHash(ae)
	} else {
		annotations = px.EmptyMap
	}
	return px.NewAnnotatedType(rt, rb.tags, annotations)
}
