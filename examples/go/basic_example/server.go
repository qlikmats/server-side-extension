/**
 * Implements the qlik.sse.Connector service.
 */

//go:generate protoc -I ../../../proto ../../../proto/ServerSideExtension.proto --go_out=plugins=grpc:./gen

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/golang/protobuf/proto"

	pb "github.com/qlikmats/server-side-extension/examples/go/basic_example/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type server struct{}

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 50051, "The server port")
)

// Function definitions.
var echoString = pb.FunctionDefinition{
	Name:         "EchoString",
	FunctionId:   0,
	FunctionType: pb.FunctionType_TENSOR,
	ReturnType:   pb.DataType_STRING,
	Params: []*pb.Parameter{
		&pb.Parameter{Name: "str1", DataType: pb.DataType_STRING},
	},
}
var sumOfRow = pb.FunctionDefinition{
	Name:         "SumOfRow",
	FunctionId:   1,
	FunctionType: pb.FunctionType_TENSOR,
	ReturnType:   pb.DataType_NUMERIC,
	Params: []*pb.Parameter{
		&pb.Parameter{Name: "col1", DataType: pb.DataType_NUMERIC},
		&pb.Parameter{Name: "col2", DataType: pb.DataType_NUMERIC},
	},
}
var functionDefinitions = []*pb.FunctionDefinition{&echoString, &sumOfRow}

// Plugin capabilities.
var capabilities = pb.Capabilities{
	AllowScript:      false,
	PluginIdentifier: "SSE Go Plugin",
	PluginVersion:    "1.0.0",
	Functions:        functionDefinitions}

/*
 * Service impl.
 */
func (*server) GetCapabilities(context.Context, *pb.Empty) (*pb.Capabilities, error) {
	for _, c := range capabilities.Functions {
		log.Printf("%+v", *c)
	}

	return &capabilities, nil
}

func (s *server) ExecuteFunction(stream pb.Connector_ExecuteFunctionServer) error {
	var functionRequestHeader = &pb.FunctionRequestHeader{}
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		binHdr := md["qlik-functionrequestheader-bin"][0]

		if err := proto.Unmarshal([]byte(binHdr), functionRequestHeader); err != nil {
			return errors.New("could not unmarshal header")
		}
	} else {
		return errors.New("failed to retrieve metadata")
	}

	log.Printf("ExecuteFunction (id: %d)", functionRequestHeader.FunctionId)

	switch functionRequestHeader.FunctionId {
	case echoString.FunctionId:
		return s.echoString(stream)
	case sumOfRow.FunctionId:
		return s.sumOfRow(stream)
	default:
		return errors.New("unimplemented function")
	}
}

func (*server) EvaluateScript(pb.Connector_EvaluateScriptServer) error {
	return errors.New("not supported/implemented")
}

/*
 * Private functions.
 */

func (*server) echoString(stream pb.Connector_ExecuteFunctionServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return nil
		}
		return stream.Send(in)
	}
}

func (*server) sumOfRow(stream pb.Connector_ExecuteFunctionServer) error {
	outBundle := new(pb.BundledRows)
	outBundle.Rows = make([]*pb.Row, 0)

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return nil
		}
		for _, row := range in.Rows {
			sum := 0.0
			for _, dual := range row.Duals {
				sum += dual.NumData
			}
			outDual := pb.Dual{NumData: sum}
			outRow := pb.Row{Duals: []*pb.Dual{&outDual}}
			outBundle.Rows = append(outBundle.Rows, &outRow)
		}
		if err := stream.Send(outBundle); err != nil {
			return err
		}
	}
}

/*
 * Main function.
 */
func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			log.Fatalf("cert_file needs to be specified.")
		}
		if *keyFile == "" {
			log.Fatalf("key_file needs to be specified.")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	s := grpc.NewServer(opts...)
	pb.RegisterConnectorServer(s, &server{})

	log.Printf("Running SSE plugin on port: %d (secure mode: %t)", *port, *tls)
	if err = s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
