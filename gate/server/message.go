package server

import "github.com/bytedance/sonic"

// Message represents a message to be broadcasted in the cluster.
type Message struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ParseMessage deserializes a JSON byte slice into a Message.
// It returns the parsed message and a boolean indicating success.
func ParseMessage(data []byte) *Message {
	msg := new(Message)
	_ = sonic.Unmarshal(data, msg)
	return msg
}

type ChanMessage struct {
	Pipe chan *Message
}

var chanMessage chan *ChanMessage
