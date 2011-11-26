package mongrel2

import (
	"fmt"
	"os"
	"github.com/alecthomas/gozmq"
	"encoding/json"
	"strings"
	"strconv"
)

type M2JsonRequest struct {
	ServerId string
	ClientId int
	ServicePath string
	MongrelInfo map[string] string
	Json map[string]interface{}
}

type M2JsonResponse struct {
	ServerId string
	ClientId []int
	Json map[string]interface{}
}

type M2JsonHandler interface {
	ReadJson() (*M2JsonRequest, error)
	WriteJson(*M2JsonResponse) error
}

type M2JsonHandlerDefault struct {
	*M2RawHandlerDefault
}

func (self *M2JsonHandlerDefault) ReadJson() (*M2JsonRequest, error) {
	
	payload, err := self.InSocket.Recv(0)
	if err != nil {
		return nil, err
	}
	serverId, clientId, path, m2, bodyStart, bodySize, error := DecodeM2PayloadStart(payload)
	
	if error!=nil {
		return nil,error
	}
	
	var content map[string]interface{}
	
	if bodySize>0 {
		err = json.Unmarshal(payload[bodyStart:bodyStart+bodySize], &content)
		if err!=nil {
			return nil, err
		}
	}
	
	result := new(M2JsonRequest)
	result.ServerId = serverId
	result.ClientId = clientId
	result.MongrelInfo = m2
	result.ServicePath = path
	result.Json = content
	
	fmt.Fprintf(os.Stderr,"client %d json=%v\n",result.ClientId, result.Json)
	
	return result,nil
}

func (self *M2JsonHandlerDefault) WriteJson(resp *M2JsonResponse) error {
	c := make([]string, len(resp.ClientId), len(resp.ClientId))
	for i, x := range resp.ClientId {
		c[i] = strconv.Itoa(x)
	}
	clientList := strings.Join(c, " ")
	b,err:=json.Marshal(resp.Json)
	if err!=nil {
		return err
	}
	body:=string(b)
	payload := FormatForMongrel2(resp.ServerId,  200,  "",  clientList,  nil, body)
	return self.OutSocket.Send([]byte(payload),0)
	
}

// ReadLoop is a loop that reads mongrel2 messages until it gets an error.  This useful if
// you want to launch a goroutine that reads forever from mongrel2 and makes the read 
// messages available on the supplied channel.
func (self *M2JsonHandlerDefault) ReadLoop(in chan *M2JsonRequest) {
	fmt.Fprintf(os.Stderr,"in read loop... about to go to read json\n")
	for {
		r, err := self.ReadJson()
		if err != nil {
			if err == gozmq.ETERM {
				return
			}
			panic(err)
		}
		in <- r
	}
}
// WriteLoop is a loop that sends mongrel two message until it gets an error
// or a message to close.  This is useful when you want to launch a goroutine
//that runs forever just taking messages from the out channel supplied and pushing them
//to mongrel2.
func (self *M2JsonHandlerDefault) WriteLoop(out chan *M2JsonResponse) {
	for {
		m := <-out
		if m == nil {
			return //end of goroutine b/c of shutdown
		}

		err := self.WriteJson(m)
		if err != nil {
			if err == gozmq.ETERM {
				return
			}
			panic(err)
		}
	}

}
