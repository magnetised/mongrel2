package mongrel2

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/gozmq"
	"io"
	//"os"
)

//HttpHandler is an interface that allows communication with the mongrel2 for serving
//HTTP requests in the particular format that mongrel2 uses.  This format is represented
//by the HttpRequest and HttpResponse types in this package.  This level does not mess
//with the channels supplied to it at the ReadLoop() and WriteLoop() methods; this
//object deals ONLY with sockets.  Anyone using ReadLoop() or WriteLoop() should close
//the channels themselves AND close the ZMQ context so that any goroutines blocked
//on a read of a socket will get ETERM and die.
type HttpHandler interface {
	ReadMessage() (*HttpRequest, error)
	WriteMessage(*HttpResponse) error
}

//HttpRequest structs are the "raw" information sent to the handler by the Mongrel2 server.
//The primary fields of the mongrel2 protocol are broken out in this struct and the
//headers (supplied by the client, passed through by Mongrel2) are included as a map.
//The RawRequest slice is the byte slice that holds all the data.  The Body byte slice
//points to the same underlying storage.  The other fields, for convenience have been
//parsed and _copied_ out of the RawRequest byte slice.
type HttpRequest struct {
	RawRequest []byte
	Body       []byte
	ServerId   string
	ClientId   int
	BodySize   int
	Path       string
	Header     map[string]string
}

//HttpResponse structss are sent back to Mongrel2 servers. The Mongrel2 server you wish
//to target should be specified with the UUID and the client of that server you wish
//to target should be in the ClientId field.  Note that this is a slice since you
//can target up to 128 clients with a single HttpResponse struct.  The other fields are
//passed through at the HTTP level to the client or clients.  The easiest way
//to correctly target a HttpResponse is by looking at the values supplied in a Request
//struct.
type HttpResponse struct {
	ServerId      string
	ClientId      []int
	Body          io.ReadCloser
	ContentLength int64
	StatusCode    int
	StatusMsg     string
	Header        map[string]string
	Stream        bool
}

//HttpHandlerDefault is a basic implementation of the HttpHandler that knows about channels.
//You can use the ReadLoop() and WriteLoop() to launch goroutines that interact correctly
//with the channels, although it never closes them.
type HttpHandlerDefault struct {
	*RawHandlerDefault
}

// ReadLoop is a loop that reads mongrel2 message until it gets an error.  This useful if
// you want to launch a goroutine that reads forever from mongrel2 and makes the read
// messages available on the supplied channel.
func (self *HttpHandlerDefault) ReadLoop(in chan *HttpRequest) {
	for {
		r, err := self.ReadMessage()
		if err != nil {
			//e := err.(gozmq.ZmqErrno)
			if err == gozmq.ETERM {
				//fmt.Printf("HTTP socket ignoring ETERM in read, signaling higher level and assuming shutdown...%p\n",self)
				self.InSocket.Close()
				//concurrency claim: we DONT want to close the in channel because the ETERM will happen only
				//at the end of shutdown processing and thus the in channel is already closed...
				return
			}
			panic(err)
		}
		select {
		case x, ok := <-in:
			//this case happens if somebody manages to close the channel right as you are reading something
			//from the server. Rare, but possible.
			fmt.Printf("discard received message because channel is closed: %v %v %p\n", x, ok, self)
			return //closin time
		case in <- r:
		}
	}
}

// WriteLoop is a loop that sends mongrel two message until it gets an error
// or a message to close.  This is useful when you want to launch a goroutine
//that runs forever just taking messages from the out channel supplied and pushing them
//to mongrel2.
func (self *HttpHandlerDefault) WriteLoop(out chan *HttpResponse) {
	for {
		//coming from higher layer to us
		m := <-out
		if m == nil {
			//fmt.Printf("HTTP socket read nil in write loop, assuming shutdown...%p\n",self)
			self.OutSocket.Close()
			return //end of goroutine b/c of shutdown
		}

		err := self.WriteMessage(m)
		if err != nil {
			//e := err.(gozmq.ZmqErrno)
			if err == gozmq.ETERM {
				//fmt.Printf("HTTP socket ignoring ETERM in write loop, assuming shutdown of %p...\n",self)
				self.OutSocket.Close()
				return
			}
			panic(err)
		}
	}

}

//ReadMessage creates a new Request struct based on the values sent from a Mongrel2
//instance. This call blocks until it receives a Request.  Note that you can have
//several different goroutines all waiting on messages from the same server and they
// will be delivered in a round-robin fashion.  This call tries to be efficient and look
//at each byte only when necessary.  The body of the request is not examined by
//this method.
func (self *HttpHandlerDefault) ReadMessage() (*HttpRequest, error) {

	req, err := self.InSocket.Recv(0)
	if err != nil {
		return nil, err
	}

	serverId, clientId, path, jsonMap, bodyStart, bodySize, err := DecodePayloadStart(req)

	result := new(HttpRequest)
	result.RawRequest = req
	result.Path = path
	result.BodySize = bodySize
	result.ServerId = serverId
	result.ClientId = clientId
	result.Header = jsonMap

	if bodySize > 0 {
		result.Body = req[bodyStart : bodyStart+bodySize]
	}

	return result, nil
}

//WriteMessage takes an HttpResponse structs and enques it for transmission.  This call
//does _not_ block.  The Response struct must be targeted for a specific server
//(ServerId) and one or more clients (ClientID).  The HttpResponse struct may be received
//by many Mongrel2 server instances, but only the server addressed in the serverId
//will transmit process the response --sending the result on to the client or clients.
func (self *HttpHandlerDefault) WriteMessage(response *HttpResponse) error {

	//create the properly mangled body in HTTP format
	buffer := new(bytes.Buffer)
	if response.StatusMsg == "" {
		buffer.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", 200, "OK"))
	} else {
		buffer.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", response.StatusCode, response.StatusMsg))
	}

	if !response.Stream && response.ContentLength == 0 && response.Body != nil {
		panic("content length set to zero but body is not nil!")
	}
	buffer.WriteString(fmt.Sprintf("Content-Length: %d\r\n", response.ContentLength))

	for k, v := range response.Header {
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	//critical, separating extra newline
	buffer.WriteString("\r\n")
	//then the body, if it exists
	if response.Body != nil {
		_, e := buffer.ReadFrom(response.Body)
		if e != nil {
			return e
		}
	}

	_, err := self.Write(response.ServerId, response.ClientId, buffer.Bytes())

	return err
}
