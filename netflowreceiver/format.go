package netflowreceiver

import "encoding/json"

type JsonFormat struct {
}

func (d *JsonFormat) Prepare() error {
	return nil
}

func (d *JsonFormat) Init() error {
	return nil
}

func (d *JsonFormat) Format(data interface{}) ([]byte, []byte, error) {
	var key []byte
	if dataIf, ok := data.(interface{ Key() []byte }); ok {
		key = dataIf.Key()
	}
	output, err := json.Marshal(data)
	return key, output, err
}
