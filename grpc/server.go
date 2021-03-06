package grpc

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/lyraproj/data-protobuf/datapb"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-evaluator/eval"
	"github.com/lyraproj/puppet-evaluator/proto"
	"github.com/lyraproj/puppet-evaluator/serialization"
	"github.com/lyraproj/puppet-evaluator/threadlocal"
	"github.com/lyraproj/puppet-evaluator/types"
	"github.com/lyraproj/servicesdk/serviceapi"
	"github.com/lyraproj/servicesdk/servicepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net/rpc"

	// Ensure that pcore is initialized
	_ "github.com/lyraproj/puppet-evaluator/pcore"
)

type GRPCServer struct {
	ctx  eval.Context
	impl serviceapi.Service
}

func (a *GRPCServer) Server(*plugin.MuxBroker) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no server implementation for rpc`, a)
}

func (a *GRPCServer) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no RPC client implementation for rpc`, a)
}

func (a *GRPCServer) GRPCServer(broker *plugin.GRPCBroker, impl *grpc.Server) error {
	servicepb.RegisterDefinitionServiceServer(impl, a)
	return nil
}

func (a *GRPCServer) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (interface{}, error) {
	return nil, fmt.Errorf(`%T has no client implementation for rpc`, a)
}

func (a *GRPCServer) Do(doer func(c eval.Context)) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(issue.Reported); ok {
				err = e
			} else {
				panic(x)
			}
		}
	}()
	c := a.ctx.Fork()
	threadlocal.Init()
	threadlocal.Set(eval.PuppetContextKey, c)
	doer(c)
	return nil
}

func (d *GRPCServer) Identity(context.Context, *servicepb.EmptyRequest) (result *datapb.Data, err error) {
	err = d.Do(func(c eval.Context) {
		result = ToDataPB(d.impl.Identifier(c))
	})
	return
}

func (d *GRPCServer) Invoke(_ context.Context, r *servicepb.InvokeRequest) (result *datapb.Data, err error) {
	err = d.Do(func(c eval.Context) {
		wrappedArgs := FromDataPB(c, r.Arguments)
		arguments := wrappedArgs.(*types.ArrayValue).AppendTo([]eval.Value{})
		rrr := d.impl.Invoke(
			c,
			r.Identifier,
			r.Method,
			arguments...)
		result = ToDataPB(rrr)
	})
	return
}

func (d *GRPCServer) Metadata(_ context.Context, r *servicepb.EmptyRequest) (result *servicepb.MetadataResponse, err error) {
	err = d.Do(func(c eval.Context) {
		ts, ds := d.impl.Metadata(c)
		vs := make([]eval.Value, len(ds))
		for i, d := range ds {
			vs[i] = d
		}
		result = &servicepb.MetadataResponse{Typeset: ToDataPB(ts), Definitions: ToDataPB(types.WrapValues(vs))}
	})
	return
}

func (d *GRPCServer) State(_ context.Context, r *servicepb.StateRequest) (result *datapb.Data, err error) {
	err = d.Do(func(c eval.Context) {
		result = ToDataPB(d.impl.State(c, r.Identifier, FromDataPB(c, r.Input).(eval.OrderedMap)))
	})
	return
}

func ToDataPB(v eval.Value) *datapb.Data {
	pc := proto.NewProtoConsumer()
	serialization.NewSerializer(eval.Puppet.RootContext(), eval.EMPTY_MAP).Convert(v, pc)
	return pc.Value()
}

func FromDataPB(c eval.Context, d *datapb.Data) eval.Value {
	ds := serialization.NewDeserializer(c, eval.EMPTY_MAP)
	proto.ConsumePBData(d, ds)
	return ds.Value()
}

// Serve the supplied Server as a go-plugin
func Serve(c eval.Context, s serviceapi.Service) {
	cfg := &plugin.ServeConfig{
		HandshakeConfig: handshake,
		Plugins: map[string]plugin.Plugin{
			"server": &GRPCServer{ctx: c, impl: s},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     hclog.Default(),
	}
	id := s.Identifier(c)
	log.Printf("Starting to serve %s\n", id)
	plugin.Serve(cfg)
	log.Printf("Done serve %s\n", id)
}
