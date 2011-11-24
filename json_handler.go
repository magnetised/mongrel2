package mongrel2

type M2JsonHandler interface {
	ReadJson() (map[string]interface{}, error)
	WriteJson(map[string]interface{}) error
}

type M2JsonHandlerDefault struct {
	*M2RawHandlerDefault
}



func (self *M2JsonHandlerDefault) ReadJson() (map[string]interface{}, error) {
	return nil,nil
}

func (self *M2JsonHandlerDefault) WriteJson(map[string]interface{}) error {
	return nil
}


