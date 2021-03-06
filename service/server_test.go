package service_test

import (
	"fmt"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/wfapi"
	"os"

	// Initialize pcore
	_ "github.com/lyraproj/puppet-evaluator/pcore"
	_ "github.com/lyraproj/servicesdk/wf"
)

type testAPI struct{}

func (*testAPI) First() string {
	return `first`
}

func (*testAPI) Second(suffix string) string {
	return `second ` + suffix
}

func ExampleServer_Invoke() {
	eval.Puppet.Do(func(c eval.Context) {
		api := `My::TheApi`
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterAPI(api, &testAPI{})

		s := sb.Server()
		fmt.Println(s.Invoke(c, api, `first`))
		fmt.Println(s.Invoke(c, api, `second`, eval.Wrap(c, `place`)))
	})

	// Output:
	// first
	// second place
}

type MyRes struct {
	Name  string
	Phone string
}

func ExampleServer_Metadata_typeSet() {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterAPI(`My::TheApi`, &testAPI{})
		sb.RegisterTypes("My", &MyRes{})

		s := sb.Server()
		ts, _ := s.Metadata(c)
		ts.ToString(os.Stdout, eval.PRETTY_EXPANDED, nil)
		fmt.Println()
	})

	// Output:
	// TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'My',
	//   version => '0.1.0',
	//   types => {
	//     MyRes => {
	//       attributes => {
	//         'name' => String,
	//         'phone' => String
	//       }
	//     },
	//     TheApi => {
	//       functions => {
	//         'first' => Callable[
	//           [0, 0],
	//           String],
	//         'second' => Callable[
	//           [String],
	//           String]
	//       }
	//     }
	//   }
	// }]
}

func ExampleServer_Metadata_definitions() {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterTypes("My", &MyRes{})
		sb.RegisterActivity(wfapi.NewWorkflow(c, func(b wfapi.WorkflowBuilder) {
			b.Name(`My::Test`)
			b.Resource(func(w wfapi.ResourceBuilder) {
				w.Name(`X`)
				w.Input(w.Parameter(`a`, `String`))
				w.Input(w.Parameter(`b`, `String`))
				w.StateStruct(&MyRes{Name: `Bob`, Phone: `12345`})
			})
		}))

		s := sb.Server()
		_, defs := s.Metadata(c)
		for _, def := range defs {
			fmt.Println(eval.ToPrettyString(def))
		}
	})

	// Output:
	// Service::Definition(
	//   'identifier' => TypedName(
	//     'namespace' => 'definition',
	//     'name' => 'My::Test'
	//   ),
	//   'serviceId' => TypedName(
	//     'namespace' => 'service',
	//     'name' => 'My::Service'
	//   ),
	//   'properties' => {
	//     'activities' => [
	//       Service::Definition(
	//         'identifier' => TypedName(
	//           'namespace' => 'definition',
	//           'name' => 'My::Test::X'
	//         ),
	//         'serviceId' => TypedName(
	//           'namespace' => 'service',
	//           'name' => 'My::Service'
	//         ),
	//         'properties' => {
	//           'input' => [
	//             Parameter(
	//               'name' => 'a',
	//               'type' => String
	//             ),
	//             Parameter(
	//               'name' => 'b',
	//               'type' => String
	//             )],
	//           'resource_type' => My::MyRes,
	//           'style' => 'resource'
	//         }
	//       )],
	//     'style' => 'workflow'
	//   }
	// )
	//
}

func ExampleServer_Metadata_state() {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterTypes("My", &MyRes{})
		sb.RegisterStateConverter(service.GoStateConverter)
		sb.RegisterActivity(wfapi.NewWorkflow(c, func(b wfapi.WorkflowBuilder) {
			b.Name(`My::Test`)
			b.Resource(func(w wfapi.ResourceBuilder) {
				w.Name(`X`)
				w.Input(w.Parameter(`a`, `String`))
				w.Input(w.Parameter(`b`, `String`))
				w.StateStruct(&MyRes{Name: `Bob`, Phone: `12345`})
			})
		}))

		s := sb.Server()
		fmt.Println(eval.ToPrettyString(s.State(c, `My::Test::X`, eval.EMPTY_MAP)))
	})

	// Output:
	// My::MyRes(
	//   'name' => 'Bob',
	//   'phone' => '12345'
	// )
}

type MyIdentityService struct {
	extToId map[string]eval.URI
	idToExt map[eval.URI]string
}

func (is *MyIdentityService) GetExternal(id eval.URI) (string, error) {
	if ext, ok := is.idToExt[id]; ok {
		return ext, nil
	}
	return ``, wfapi.NotFound
}

func (is *MyIdentityService) GetInternal(ext string) (eval.URI, error) {
	if id, ok := is.extToId[ext]; ok {
		return id, nil
	}
	return eval.URI(``), wfapi.NotFound
}

func ExampleServer_Metadata_api() {
	eval.Puppet.Do(func(c eval.Context) {
		sb := service.NewServerBuilder(c, `My::Service`)

		sb.RegisterAPI(`My::Identity`, &MyIdentityService{map[string]eval.URI{}, map[eval.URI]string{}})

		s := sb.Server()
		ts, defs := s.Metadata(c)
		ts.ToString(os.Stdout, eval.PRETTY_EXPANDED, nil)
		fmt.Println()
		for _, def := range defs {
			fmt.Println(eval.ToPrettyString(def))
		}
	})

	// Output:
	// TypeSet[{
	//   pcore_uri => 'http://puppet.com/2016.1/pcore',
	//   pcore_version => '1.0.0',
	//   name_authority => 'http://puppet.com/2016.1/runtime',
	//   name => 'My',
	//   version => '0.1.0',
	//   types => {
	//     Identity => {
	//       functions => {
	//         'get_external' => Callable[
	//           [String],
	//           String],
	//         'get_internal' => Callable[
	//           [String],
	//           String]
	//       }
	//     }
	//   }
	// }]
	// Service::Definition(
	//   'identifier' => TypedName(
	//     'namespace' => 'definition',
	//     'name' => 'My::Identity'
	//   ),
	//   'serviceId' => TypedName(
	//     'namespace' => 'service',
	//     'name' => 'My::Service'
	//   ),
	//   'properties' => {
	//     'interface' => My::Identity,
	//     'style' => 'callable'
	//   }
	// )
	//
}
