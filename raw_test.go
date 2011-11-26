package mongrel2

import (
	"launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the default gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type MongrelSuite struct {

}

// hook up suite to gocheck
var _ = gocheck.Suite(&MongrelSuite{})

var (
	JSON_SAMPLE = `1ccef67e-f118-413b-9cce-f67ef118d13b 164 @chat 80:{"PATH":"@chat","x-forwarded-for":"127.0.0.1","METHOD":"JSON","PATTERN":"@chat"},46:{"type":"msg",
"msg":"foo",
"user":"lamenick"},`
	GET_SAMPLE = `0de9b17e-e958-4502-8de9-b17ee958d502 235 /echo/50285a0c-d1e3-4deb-9028-5a0cd1e35deb 268:{"PATH":"/echo/50285a0c-d1e3-4deb-9028-5a0cd1e35deb","x-forwarded-for":"127.0.0.1","accept-encoding":"gzip","user-agent":"Go http package","host":"localhost:6767","METHOD":"GET","VERSION":"HTTP/1.1","URI":"/echo/50285a0c-d1e3-4deb-9028-5a0cd1e35deb","PATTERN":"/echo"},0:,`	
)

func (self *MongrelSuite) SetUpSuite(c *gocheck.C) {
}

func (self *MongrelSuite) TearDownSuite(c *gocheck.C) {
}

//little example provided by the gocheck author
func (s *MongrelSuite) TestPayloadDecodingJson(c *gocheck.C) {
	req := []byte(string(JSON_SAMPLE))
	serverId, clientId, path, jsonmap, bodyStart, bodySize, error := DecodeM2PayloadStart(req)

	c.Check(error, gocheck.Equals, nil)
	c.Check(serverId, gocheck.Equals, "1ccef67e-f118-413b-9cce-f67ef118d13b") 
	c.Check(clientId, gocheck.Equals, 164)
	c.Check(path, gocheck.Equals, "@chat")
	c.Check(jsonmap["PATH"], gocheck.Equals, "@chat")
	c.Check(jsonmap["METHOD"], gocheck.Equals, "JSON")
	c.Check(int(req[bodyStart]), gocheck.Equals,int('{'))
	c.Check(int(req[bodyStart+bodySize]), gocheck.Equals,int(','))
	c.Check(int(req[bodyStart+bodySize-1]), gocheck.Equals,int('}'))
}
func (s *MongrelSuite) TestPayloadDecodingGet(c *gocheck.C) {
	req := []byte(string(GET_SAMPLE))
	serverId, clientId, path, jsonmap, _, bodySize, error := DecodeM2PayloadStart(req)

	c.Check(error, gocheck.Equals, nil)
	c.Check(serverId, gocheck.Equals, "0de9b17e-e958-4502-8de9-b17ee958d502") 
	c.Check(clientId, gocheck.Equals, 235)
	p:="/echo/50285a0c-d1e3-4deb-9028-5a0cd1e35deb"
	c.Check(path, gocheck.Equals, p)
	c.Check(jsonmap["PATH"], gocheck.Equals, p)
	c.Check(jsonmap["METHOD"], gocheck.Equals, "GET")
	c.Check(0, gocheck.Equals, bodySize)
}
