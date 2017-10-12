# Writing an SSE plugin using Go

General background information is found here:

* [Writing an SSE Plugin](../../docs/writing_a_plugin.md)
* [Protocol Description](../../docs/SSE_Protocol.md) (API reference)
* Documentation and tutorials for [gRPC](http://www.grpc.io/docs/) and [protobuf](https://developers.google.com/protocol-buffers/docs/overview)

## Getting started
First of all, you need an installation of the [Go](https://golang.org/) programming language. 

Next, you need to download and install the [protobuf compiler](https://github.com/google/protobuf/releases) and add the *protoc* executable to your %PATH%. Then, to install the protoc plugin for Go, execute:

*go get -u github.com/golang/protobuf/protoc-gen-go*

Note that the *protoc-gen-go* program also needs to be in your %PATH%.

After setting up your [Go](https://golang.org/) environment, open a command prompt and execute: 

*go get github.com/qlikmats/server-side-extension/examples/go/basic_example*

You will get an "error" saying there are no go files in the \gen folder, but that's  fine - we will generate a file in that folder in the following step.

From your Go SSE plugin folder (i.e. %GOPATH%\src\github.com\qlikmats\server-side-extension\examples\go\basic_example), execute:

*go generate*

This will generate the gRPC/Protobuf file(s) that your server will use  for implementing the qlik.sse.Connector service.



### Configuring QlikSense to use the sample gRPC server
By default, the Go sample plug-in runs on port 50051 on localhost, so for a QlikSense Desktop installation, the following should be added to settings.ini:

*[Settings 7]* 
*SSEPlugin=SSE_Go, localhost:50051[,PATH_TO_CERTIFICATE_FOLDER]*

Note that the string SSE_Go is the identifier that will prefix all plug-in functions when they are called from within Qlik.
Use a different identifier for your own plug-in, and remember that this exact string has to be used for your Sense applications to work with the extension.

The address (localhost:50051) should of course match the address in the server's configuration file.

For single-machine development and testing it is OK to use unsecure communication, but for production scenarios you should use certificates. See [Generating certificates](../../generate_certs_guide/README.md).

## Starting the server

The Go gRPC sample server can be built and started using the go command:

*go run server.go*

This will start the server in unsecure mode. To run in secure mode, execute the following command instead:

*go run server.go -tls -cert_file=..\..\..\generate_certs_guide\sse_qliktest_generated_certs\sse_qliktest_server_certs\sse_server_cert.pem -key_file=..\..\..\generate_certs_guide\sse_qliktest_generated_certs\sse_qliktest_server_certs\sse_server_key.pem*

## Run the example app
Copy the file *SSE Go.qvf* to your Sense apps folder, i.e. *C:\Users\[user]\Qlik\Sense\Apps* and start Sense desktop (make sure the SSE Go plugin is up and running).


## Implementing a server - Protobuf generated files
The interface between the Qlik Engine acting as a client and the Server-side extension acting as server is defined in 
the file [ServerSideExtension.proto](../../proto/ServerSideExtension.proto). 


## RPC methods
The RPC methods implemented in the Basic example are GetCapabilities and ExecuteFunction. There is no script support.

### GetCapabilities method
The GetCapabilities method returns a Capabilities object, describing the operations supported by the plugin.

### ExecuteFunction method
The ExecuteFunction method switches over the numeric function identifier sent in the *qlik-functionrequestheader-bin* header. 
Each case construct then iterates over the BundledRows elements packed into the request stream and writes the results to the output stream.  There are two functions implemented in the Go example plugin:

 - *EchoString* (echoes the supplied string back to Sense)
 - *SumOfRow* (summarizes two columns, in a row-wise manner)

