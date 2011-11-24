package main

import (
	"fmt"
	"mongrel2"
	"os"
	"time"
)

//Simple demo program that processes requests one by one forever.  It uses the
//go channel mechanism.
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

	//allocate the in and out channels that we'll be using
	in:=make(chan *mongrel2.M2HttpRequest)
	out:=make(chan *mongrel2.M2HttpResponse)

	// this allocates the "raw" abstraction for talking to a mongrel server	
	// mongrel doc refers to this as a "handler"
	handler := new(mongrel2.M2HttpHandlerDefault)
	err := handler.Bind("sample2",ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing mongrel connection (Bind):%s\n", err)
		return
	}
	// don't forget to clean up various resources when done
	defer handler.Shutdown()

	//loop forever taking anything mongrel2 sends to us and putting on a channel
	go handler.ReadLoop(in)
	//loop forever taking anything we put on the out channel and sending to mongrel
	go handler.WriteLoop(out)
	

	//lets read 3 messages
	for i:=0; i<3;i++ {
		fmt.Printf("waiting on a message from on the in channel...\n")
		
		//blocks!
		req := <- in
		
		//create a response to go back to the client
		response := new(mongrel2.M2HttpResponse)

		//note: copying the serverid and clientid to target the appropriate browser!
		response.ServerId = req.ServerId
		response.ClientId = []int{req.ClientId}

		//make up a simple body for the user to see
		response.Body = fmt.Sprintf("<pre>howdy %s, with client %d!</pre>", req.ServerId, req.ClientId)
		
		//send it via the channel
		out <- response
	}

	//this is what we have to do to make sure the sent message gets delivered
	//before we shut down. this waits 1.5 secs.  if you want to get fancy you
	//can access the ZMQ out socket and set the linger time to -1.  then, when
	//you close the context (due to the defer above) it will wait until the message
	//is delivered
	time.Sleep(1500000)
}
