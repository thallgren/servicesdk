module github.com/lyraproj/servicesdk

require (
	github.com/golang/protobuf v1.2.0
	github.com/hashicorp/go-hclog v0.0.0-20181001195459-61d530d6c27f
	github.com/hashicorp/go-plugin v0.0.0-20181212150838-f444068e8f5a
	github.com/lyraproj/data-protobuf v0.0.0-20181217135414-3d508204b820
	github.com/lyraproj/issue v0.0.0-20181208172701-8d203563a8dc
	github.com/lyraproj/puppet-evaluator v0.0.0
	github.com/lyraproj/semver v0.0.0-20181213164306-02ecea2cd6a2
	golang.org/x/net v0.0.0-20181220203305-927f97764cc3
	google.golang.org/grpc v1.17.0
)

replace github.com/lyraproj/puppet-evaluator => ../puppet-evaluator
