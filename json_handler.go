package mongrel2

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/gozmq"
	"strconv"
	"strings"
)

type JsonRequest struct {
	ServerId    string
	ClientId    int
	ServicePath string
	MongrelInfo map[string]string
	Json        map[string]interface{}
}

type JsonResponse struct {
	ServerId string
	ClientId []int
	Json     map[string]interface{}
}

type JsonHandler interface {
	ReadJson() (*JsonRequest, error)
	WriteJson(*JsonResponse) error
}

type JsonHandlerDefault struct {
	*RawHandlerDefault
}

func (self *JsonHandlerDefault) ReadJson() (*JsonRequest, error) {

	payload, err := self.InSocket.Recv(0)
	if err != nil {
		return nil, err
	}
	serverId, clientId, path, info, bodyStart, bodySize, error := DecodePayloadStart(payload)

	if error != nil {
		return nil, error
	}

	var content map[string]interface{}

	if bodySize > 0 {
		err = json.Unmarshal(payload[bodyStart:bodyStart+bodySize], &content)
		if err != nil {
			return nil, err
		}
	}

	result := new(JsonRequest)
	result.ServerId = serverId
	result.ClientId = clientId
	result.MongrelInfo = info
	result.ServicePath = path
	result.Json = content

	return result, nil
}

func (self *JsonHandlerDefault) WriteJson(resp *JsonResponse) error {
	c := make([]string, len(resp.ClientId), len(resp.ClientId))
	for i, x := range resp.ClientId {
		c[i] = strconv.Itoa(x)
	}
	clientList := strings.Join(c, " ")
	b, err := json.Marshal(resp.Json)
	if err != nil {
		return err
	}

	body := string(b)
	payload := fmt.Sprintf("%s %d:%s, %s", resp.ServerId, len(clientList), clientList, body)
	return self.OutSocket.Send([]byte(payload), 0)

}

// ReadLoop is a loop that reads mongrel2 messages until it gets an error.  This useful if
// you want to launch a goroutine that reads forever from mongrel2 and makes the read 
// messages available on the supplied channel.
func (self *JsonHandlerDefault) ReadLoop(in chan *JsonRequest) {
	for {
		r, err := self.ReadJson()
		if err != nil {
			if err == gozmq.ETERM {
				fmt.Printf("JSON socket ignoring ETERM on read, assuming shutdown...\n")
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
func (self *JsonHandlerDefault) WriteLoop(out chan *JsonResponse) {
	for {
		m := <-out
		if m == nil {
			return //end of goroutine b/c of shutdown
		}

		err := self.WriteJson(m)
		if err != nil {
			if err == gozmq.ETERM {
				fmt.Printf("JSON socket ignoring ETERM on write, assuming shutdown...\n")
				return
			}
			panic(err)
		}
	}

}
