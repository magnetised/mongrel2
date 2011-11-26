//The mongrel2 package provides a way to write a handler in go for Mongrel2.
package mongrel2

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alecthomas/gozmq"
	"os"
	"strconv"
)

//RawHandler is the basic type for an object that communicate with mongrel2.  This interface
//just knows about sockets, nothing about what to do with communication on those sockets.
//Developers should not need this type.
type RawHandler interface {
	Bind(name string, ctx gozmq.Context) error
	Shutdown() error
}

//Handler is a low-level an implementation of the interface RawHandler
//for connecting, via 0MQ, to a mongrel2 server. Developers should not need
//this type, it is part of the implementation of this package.
type RawHandlerDefault struct {
	InSocket, OutSocket         gozmq.Socket
	PullSpec, PubSpec, Identity string
}

//initZMQ creates the necessary ZMQ machinery and sets the fields of the
//Mongrel2 struct.  This is normally called via the Init() method.
func (self *RawHandlerDefault) InitZMQ(ctx gozmq.Context) error {

	s, err := ctx.NewSocket(gozmq.PULL)
	if err != nil {
		return err
	}
	self.InSocket = s

	err = self.InSocket.Connect(self.PullSpec)
	if err != nil {
		return err
	}

	err = self.InSocket.SetSockOptInt(gozmq.LINGER, 0)
	if err != nil {
		return err
	}

	s, err = ctx.NewSocket(gozmq.PUB)
	if err != nil {
		return err
	}
	self.OutSocket = s

	err = self.OutSocket.SetSockOptString(gozmq.IDENTITY, self.Identity)
	if err != nil {
		return err
	}

	err = self.OutSocket.SetSockOptInt(gozmq.LINGER, 0)
	if err != nil {
		return err
	}

	err = self.OutSocket.Connect(self.PubSpec)
	if err != nil {
		return err
	}

	return nil
}

//Bind is a method that allocates the zmq resources needed for a connection
//to mongrel2.  It uses the supplied context to allocate the resources and allocates
//an address based on the name and uses that for the send and receive sockets.  If
//called multiple times, it has no effect.
func (self *RawHandlerDefault) Bind(name string, ctx gozmq.Context) error {
	//this only needs to be done once for a particular name, even if you call
	//Shutdown() and Bind() again.
	if self.Identity == "" {
		address, err := GetHandlerSpec(name)
		if err != nil {
			return err
		}

		self.PullSpec = address.PullSpec
		self.PubSpec = address.PubSpec
		self.Identity = address.Identity
	}

	if self.InSocket == nil {
		err := self.InitZMQ(ctx)
		if err != nil {
			return errors.New("0mq init:" + err.Error())
		}
	}
	return nil
}

//Shutdown cleans up the resources associated with this mongrel2 connection.
//Normally this function should be part of a defer call that is immediately after
//allocating the resources, like this:
//	mongrel:=new(RawHandlerDefault)
//  defer mongrel.Shutdown()
// Note that this does not close the context because the context is supplied from
// outside the handler.
func (self *RawHandlerDefault) Shutdown() error {

	//dump the ZMQ level sockets
	if self.InSocket != nil {
		if err := self.InSocket.Close(); err != nil {
			return err
		}
		if err := self.OutSocket.Close(); err != nil {
			return err
		}
		self.InSocket = nil
		self.OutSocket = nil
	}
	return nil
}

//MustCreateContext is a function that creates a ZMQ context or panics trying to do so.
//Useful if you can't do any work without a ZMQ context.
func MustCreateContext() gozmq.Context {
	// do a version check
	x, y, z := gozmq.Version()
	if x != 2 && y != 1 {
		fmt.Fprintf(os.Stderr, "version of zmq is %d.%d.%d and this code was tested primarily on 2.1.10\n", x, y, z)
	}

	//initialize zmq... only once per address space
	ctx, err := gozmq.NewContext()
	if err != nil {
		panic(fmt.Sprintf("unable to initialize zmq context:%s\n", err))
	}
	return ctx
}

//DecodePayloadStart decodes the front of a packet from the mongrel2 server destined for
//a backend.  The actual bytes of the body are not decoded because they differet between
//different types of handlers.
func DecodePayloadStart(req []byte) (serverId string, clientId int, path string, jsonmap map[string]string, bodyStart int, bodySize int, err error) {

	endOfServerId := readSome(' ', req, 0)
	serverId = string(req[0:endOfServerId])

	endOfClientId := readSome(' ', req, endOfServerId+1)
	clientId, err = strconv.Atoi(string(req[endOfServerId+1 : endOfClientId]))
	if err != nil {
		return
	}

	endOfPath := readSome(' ', req, endOfClientId+1)
	path = string(req[endOfClientId+1 : endOfPath])

	var jsonSize int
	endOfJsonSize := readSome(':', req, endOfPath+1)

	jsonSize, err = strconv.Atoi(string(req[endOfPath+1 : endOfJsonSize]))
	if err != nil {
		return
	}

	jsonmap = make(map[string]string)
	jsonStart := endOfJsonSize + 1

	if jsonSize > 0 {
		err = json.Unmarshal(req[jsonStart:jsonStart+jsonSize], &jsonmap)
		if err != nil {
			return
		}
	}

	bodySizeStart := (jsonSize + 1) + jsonStart
	bodySizeEnd := readSome(':', req, bodySizeStart)
	bodyStart = bodySizeEnd + 1

	bodySize, err = strconv.Atoi(string(req[bodySizeStart:bodySizeEnd]))

	if err != nil {
		return
	}

	if bodySize > 0 {
		//bodySize -= 1
	}
	return
}

func readSome(terminationChar byte, req []byte, start int) int {
	result := start
	for {
		if req[result] == terminationChar {
			break
		}
		result++
	}
	return result
}
