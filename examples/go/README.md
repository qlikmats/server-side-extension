# Writing an SSE plugin using Go

General background information is found here:

* [Writing an SSE Plugin](../../docs/writing_a_plugin.md)
* [Protocol Description](../../docs/SSE_Protocol.md) (API reference)
* Documentation and tutorials for [gRPC](http://www.grpc.io/docs/) and [protobuf](https://developers.google.com/protocol-buffers/docs/overview)

## Configuration

### Configuration file settings

There are three settings in the configuration file. If you change these settings, the example program has to be restarted for new settings to take effect.
#### BasicExample.exe.config settings for grpcHost and grpcPort
These constitute the address of the gRPC server.

#### BasicExample.exe.config setting for certificateFolder
The certificate folder is where the server expects to find the public key to the root certificate and server certificate, along with the private key for the server certificate.
See the guide for [Generating certificates](../../generate_certs_guide/README.md) for more information on how to create and configure certificates.

### Configuring QlikSense to use the sample gRPC server
By default, the C# sample plug-in runs on port 50054 on localhost, so for a QlikSense Desktop installation, the following should be added to settings.ini:

[Settings 7] 

SSEPlugin=SSE_Go, localhost:50055;

Note that the string SSE_Go is the identifier that will prefix all plug-in functions when they are called from within Qlik.
Use a different identifier for your own plug-in, and remember that this exact string has to be used for the supplied qvf file to work with the extension.

The address (localhost:50055) should of course match the address in the server's configuration file.

For single-machine development and testing it is OK to use insecure communication, but for production scenarios you should use certificates. See [Generating certificates](../../generate_certs_guide/README.md).

## Starting the server

The Go gRPC sample server can be built and started using the go command:
go run server.go

## Implementing a server - Protobuf generated files
The interface between the Qlik Engine acting as a client and the Server-side extension acting as server is defined in 
the file [ServerSideExtension.proto](../../proto/ServerSideExtension.proto). 


## RPC methods
The RPC methods implemented in the Basic example are GetCapabilities and ExecuteFunction. There is no script support.

### GetCapabilities method
The GetCapabilities method just returns a Capabilities object.

### ExecuteFunction method
The ExecuteFunction method switches over the numeric function identifier sent in the qlik-functionrequestheader-bin header. 
Each case construct then iterates over the BundledRows elements packed into the request stream and writes the results to the output stream. 

