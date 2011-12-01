package main

import (
	"fmt"
	"mongrel2"
	"os"
	"time"
	"strings"
)

//Simple demo program that processes one request and returns one response to
//the server and client that sent the request.  This does not use goroutines
//it is the most raw possible interface to talking with mongrel2 at the
//http protocol level.
//
//This expects that you have configured your mongrel two server with a handler
//like this:
//handler_test = Handler(	send_spec='tcp://127.0.0.1:10070',
//                      	send_ident='34f9ceee-cd52-4b7f-b197-88bf2f0ec378',
//                      	recv_spec='tcp://127.0.0.1:10071',
//							recv_ident='') 
// Also, somewhere in your config you need to bind that handler to a path.


func main() {

	//initialize zmq... only once per address space
	ctx := mongrel2.MustCreateContext()
	//remember to close it
	defer func() {
		ctx.Close()
	}()

	//Note: the HttpHandlerDefault is *both* an RawHandler and HttpHandler because
	//it borrows the implementation of mongrel2.RawHandlerDefault.  These interfaces
	//are separated so that one can combine them any way you want, such as an object
	//"foo" that is RawHandler, HttpHandler, and JSHandler at the same time.  However
	//the default implementations combine these together for convenience.
	
	var implementation *mongrel2.HttpHandlerDefault
	var httpInterface mongrel2.HttpHandler
	var socketInterface mongrel2.RawHandler
	var err error

	implementation = &mongrel2.HttpHandlerDefault{&mongrel2.RawHandlerDefault{}}
	httpInterface = implementation    // to illustrate the types
	socketInterface = implementation  // to illustrate the types

	//we want to work with the socket layer first
	err = socketInterface.Bind("sample1",ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing mongrel connection (Bind):%s\n", err)
		return
	}

	// don't forget to clean up various socket resources when done
	defer socketInterface.Shutdown()

	fmt.Printf("waiting on a message from the mongrel2 server...\n")

	//now we want to work with the http layer... this blocks waiting for a read
	req, err := httpInterface.ReadMessage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading from mongrel connection:%s\n", err)
		return
	}

	// there are many interesting fields in req, but we just print out a couple
	fmt.Printf("server %s sent %s from client %d\n", req.ServerId, req.Path, req.ClientId)

	//create a response to go back to the client
	response := new(mongrel2.HttpResponse)

	//note: copying the serverid and clientid to target the appropriate browser!
	response.ServerId = req.ServerId
	response.ClientId = []int{req.ClientId}

	//make up a simple body for the user to see
	b:= fmt.Sprintf("<pre>hello there, %s with client %d!</pre>", req.ServerId, req.ClientId)
	response.Body=strings.NewReader(b)
	response.ContentLength=len(b)

	//send to the mongrel2 server (via the http interface) and it eventually ends up at the browser. 
	//This does NOT block waiting to send!
	err = httpInterface.WriteMessage(response)

	//this is what we have to do to make sure the sent message gets delivered
	//before we shut down. this waits 1.5 secs.  if you want to get fancy you
	//can access the ZMQ out socket and set the linger time to -1.  then, when
	//you close the context (due to the defer above) it will wait until the message
	//is delivered
	time.Sleep(1500000)
}
