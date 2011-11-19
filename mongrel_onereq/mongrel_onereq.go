package main

import (
	"fmt"
	"github.com/alecthomas/gozmq"
	"mongrel2"
	"os"
	"time"
)

//Simple demo program that processes one request and returns one response to
//the server and client that sent the request.
//
//This expects that you have configured your mongrel two server with a handler
//like this:
//handler_test = Handler(	send_spec='tcp://127.0.0.1:10070',
//                      	send_ident='34f9ceee-cd52-4b7f-b197-88bf2f0ec378',
//                      	recv_spec='tcp://127.0.0.1:10071',
//							recv_ident='') 
// Also, somewhere in your config you need to bind that handler to a path.


func main() {

	// do a version check
	x, y, z := gozmq.Version()
	if x != 2 && y != 1 {
		fmt.Printf("version of zmq is %d.%d.%d and this code was tested primarily on 2.1.10\n", x, y, z)
	}
	// we need to pick a name and register it
	addr, err := mongrel2.GetHandlerAddress("some_name") //we dont really care

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get an address for our handler:%s\n", err)
		return
	}
	
	//initialize zmq... only once per address space
	ctx, err := gozmq.NewContext()
	if err!=nil {
		fmt.Fprintf(os.Stderr, "unable to initialize zmq context:%s\n",err)
	}
	//remember to close it
	defer func() {
		ctx.Close()
	}()

	//allocate channels so we can talk to the mongrel2 system with go
	// abstractions
	in := make(chan *mongrel2.Request)
	out := make(chan *mongrel2.Response)

	// this allocates the "raw" abstraction for talking to a mongrel server
	// mongrel doc refers to this as a "handler"
	handler, err := mongrel2.NewHandler(addr, in, out, ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing mongrel connection:%s\n", err)
		return
	}
		
	// don't forget to clean up various resources when done
	defer handler.Shutdown()

	fmt.Printf("waiting on a message from the mongrel2 server...\n")
	//block until we get a message from the server
	req := <- in 
			
	// there are many interesting fields in req, but we just print out a couple
	fmt.Printf("server %s sent %s from client %d\n", req.ServerId, req.Path, req.ClientId)

	//create a response to go back to the client
	response := new(mongrel2.Response)
	response.ServerId = req.ServerId
	response.ClientId = []int{req.ClientId}
	response.Body = fmt.Sprintf("<pre>hello there, %s with client %d!</pre>", req.ServerId, req.ClientId)

	//send it via the other channel
	fmt.Printf("Responding to server with %d bytes of content\n",len(response.Body))
	out <- response
	
	//this is what we have to do to make sure the sent message gets delivered
	//before we shut down.  we tried to use ZMQ_LINGER(-1) at init time but
	//this is generating an error right now.
	time.Sleep(1000000000)
}
