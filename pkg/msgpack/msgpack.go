package msgpack

import "github.com/shamaton/msgpack/v2"

// Serializer implements the serialize.Serializer interface
type Serializer struct{}

// NewSerializer returns a new Serializer.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Marshal returns the msgpack encoding of v.
func (s *Serializer) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal parses the msgpack-encoded data and stores the result
// in the value pointed to by v.
func (s *Serializer) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// GetName returns the name of the serializer.
func (s *Serializer) GetName() string {
	return "msgpack"
}
