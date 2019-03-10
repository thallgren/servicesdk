package service

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/types"
	"github.com/lyraproj/semver/semver"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/wfapi"
	"reflect"
	"strings"
	"sync"
)

var ServerVersion = semver.MustParseVersion(`0.1.0`)

type Server struct {
	context   px.Context
	id        px.TypedName
	lock      sync.RWMutex
	typeSet   px.TypeSet
	metadata  px.List
	stateConv wfapi.StateConverter
	states    map[string]wfapi.State
	callables map[string]px.Value
}

func (s *Server) AddApi(name string, callable interface{}) serviceapi.Definition {
	rv := reflect.ValueOf(callable)
	rt := rv.Type()
	pt, ok := s.context.ImplementationRegistry().ReflectedToType(rt)
	if !ok {
		panic(px.Error(WF_API_TYPE_NOT_REGISTERED, issue.H{`type`: rt.Name()}))
	}

	s.lock.RLock()
	_, found := s.callables[name]
	s.lock.RUnlock()

	if found {
		panic(px.Error(WF_ALREADY_REGISTERED, issue.H{`namespace`: px.NsInterface, `identifier`: name}))
	}

	props := make([]*types.HashEntry, 0, 2)
	props = append(props, types.WrapHashEntry2(`interface`, pt))
	props = append(props, types.WrapHashEntry2(`style`, types.WrapString(`callable`)))
	def := serviceapi.NewDefinition(px.NewTypedName(px.NsDefinition, name), s.id, types.WrapHash(props))

	nmd := s.metadata.Add(def)
	defw := px.WrapReflected(s.context, rv)

	s.lock.Lock()
	s.callables[name] = defw
	s.metadata = nmd
	s.lock.Unlock()

	return def
}

func (s *Server) State(c px.Context, name string, input px.OrderedMap) px.PuppetObject {
	if s.stateConv != nil {
		s.lock.RLock()
		st, ok := s.states[name]
		s.lock.RUnlock()
		if ok {
			return s.stateConv(c, st, input)
		}
		panic(px.Error(WF_NO_SUCH_STATE, issue.H{`name`: name}))
	}
	panic(px.Error(WF_NO_STATE_CONVERTER, issue.H{`name`: name}))
}

func (s *Server) Identifier(px.Context) px.TypedName {
	return s.id
}

func (s *Server) Invoke(c px.Context, api, name string, arguments ...px.Value) (result px.Value) {
	s.lock.RLock()
	api = strings.Title(api)
	iv, ok := s.callables[api]
	s.lock.RUnlock()
	if ok {
		if m, ok := iv.PType().(px.TypeWithCallableMembers).Member(name); ok {
			defer func() {
				if x := recover(); x != nil {
					if err, ok := x.(issue.Reported); ok && string(err.Code()) == px.GoFunctionError {
						result = serviceapi.ErrorFromReported(c, err)
						return
					}
					panic(x)
				}
			}()
			result = m.Call(c, iv, nil, arguments)
			return
		}
		panic(px.Error(WF_NO_SUCH_METHOD, issue.H{`api`: api, `method`: name}))
	}
	panic(px.Error(WF_NO_SUCH_API, issue.H{`api`: api}))
}

func (s *Server) Metadata(px.Context) (typeSet px.TypeSet, definitions []serviceapi.Definition) {
	ds := make([]serviceapi.Definition, s.metadata.Len())
	s.lock.RLock()
	s.metadata.EachWithIndex(func(v px.Value, i int) { ds[i] = v.(serviceapi.Definition) })
	s.lock.RUnlock()
	return s.typeSet, ds
}
