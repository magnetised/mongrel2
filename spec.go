package mongrel2

import (
	"fmt"
	"hash/fnv"
)

// HandlerSpec is returned in response a request for the location (in 
// 0mq terms) of a particular named handler.  It contains the mongrel2
// necessary specifications of the Pull and Pub sockets, plus the unique
// id of the handler.  The Pull socket is assigned the lower of the two
// port numbers.
type HandlerSpec struct {
	Name     string
	PubSpec  string
	PullSpec string
	Identity string
}

var (
	//handler is the private mapping that keps the binding between names and
	//the handler addresses.
	handler = make(map[string]*HandlerSpec)

	//currentPort is the next port number to be assigned by the GetAssignment
	//function.  It never decreases.
	currentPort = 10070
	
	//fnv hash is reused
	hasher = fnv.New64()
)

//GetAssignment is used to find a spec for a handler of a given name.  If
//name has been previously assigned a HandlerAddr the previously allocated
//address is returned, otherwise a new HandlerAddr is created and returned.
func GetHandlerSpec(name string) (*HandlerSpec, error) {
	a := handler[name]
	if a != nil {
		return a, nil
	}
	result := new(HandlerSpec)
	result.Name= name
	result.PullSpec = fmt.Sprintf("tcp://127.0.0.1:%d", currentPort)
	currentPort++
	result.PubSpec = fmt.Sprintf("tcp://127.0.0.1:%d", currentPort)
	currentPort++
	result.Identity = Checksum(name)
	handler[name] = result
	return result, nil
}

//This is a cheap and cheerful way to generate something that looks like a
//unique ID but is always the same for a given string.  It computes a 
//fnv64 has and then uses that as both bytes 0-7 and 8-15.  It outputs a 
//string of bytes the same way that Type4UUID() does, in RFC 4122 format.
//
// Based on posting by Russ Cox to go-nuts mailing list
//http://groups.google.com/group/golang-nuts/msg/5ebbdd72e2d40c09
func Checksum(s string) string {
	b := make([]byte, 16)
	
	hasher.Reset()
	hasher.Write([]byte(s))
	hasher.Sum(b)
	
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}
