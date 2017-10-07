/**
 * Implements the qlik.sse.Connector service.
 */

//go:generate protoc -I .\proto .\proto\ServerSideExtension.proto --go_out=plugins=grpc:.\proto

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/qlikmats/server-side-extension/examples/go/basic_example/proto"

	"google.golang.org/grpc"
)

type server struct{}

const (
	port = ":50051"
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
var sumOfRows = pb.FunctionDefinition{
	Name:         "SumOfRow",
	FunctionId:   1,
	FunctionType: pb.FunctionType_TENSOR,
	ReturnType:   pb.DataType_NUMERIC,
	Params: []*pb.Parameter{
		&pb.Parameter{Name: "col1", DataType: pb.DataType_NUMERIC},
		&pb.Parameter{Name: "col2", DataType: pb.DataType_NUMERIC},
	},
}
var functionDefinitions = []*pb.FunctionDefinition{&echoString, &sumOfRows}

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
		fmt.Printf("%+v\n", *c)
	}

	return &capabilities, nil
}

func (s *server) ExecuteFunction(stream pb.Connector_ExecuteFunctionServer) error {
	// var binHdr = ""
	// if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
	// 	binHdr = md["qlik-functionrequestheader-bin"][0]
	// }
	// if id, err := decodeBinHeader(binHdr); err != nil {
	// 	fmt.Printf("Decode error: %v", err)
	// } else {
	// 	fmt.Printf("Function Id: %s\n", string(id))
	// }

	return s.sumOfRow(stream)
}

func (*server) EvaluateScript(pb.Connector_EvaluateScriptServer) error {
	return nil
}

/*
 * Private functions.
 */
func decodeBinHeader(s string) ([]byte, error) {
	if len(s)%4 == 0 {
		// Input was padded, or padding was not necessary.
		return base64.StdEncoding.DecodeString(s)
	}
	return base64.RawStdEncoding.DecodeString(s)
}

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
			sum := row.Duals[0].NumData + row.Duals[1].NumData
			outDual := pb.Dual{
				NumData: sum, StrData: "",
			}
			outRow := pb.Row{
				Duals: []*pb.Dual{&outDual},
			}
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
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %v: %v", port, err)
	}
	s := grpc.NewServer()
	pb.RegisterConnectorServer(s, &server{})

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
