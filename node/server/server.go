package server

import (
	"fmt"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/hashicorp/memberlist"
	"github.com/tidwall/gjson"
	"github.com/trianglehasfoursides/mathrock/node/database"
	"go.uber.org/zap"
)

var (
	node           *Node
	gossipDelegate *GossipDelegate
	err            error
)

type Node struct {
	Gate  memberlist.Address
	Gosip *memberlist.Memberlist
}

// GossipDelegate handles message broadcasting and receiving in the cluster.
type GossipDelegate struct {
	MessageChannel chan []byte
	Queue          *memberlist.TransmitLimitedQueue
}

// NotifyMsg is called when a message is received from the cluster.
// It forwards the message to the message channel.
// TODO: "Implement message handling logic here."
func (d *GossipDelegate) NotifyMsg(msg []byte) {
	var message = ParseMessage(msg)
	switch message.Key {
	case "ping":
		_ = sonic.Unmarshal([]byte(message.Value), node.Gate)
		return
	case "create_db":
		var mutex = new(sync.Mutex)
		var name, category string = gjson.Get(string(message.Value), "name").String(), gjson.Get(string(message.Value), "category").String()
		if err := database.CreateDatabase(name, category, mutex); err != nil {
			var err = fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
			// Send error message back to the gate
			node.Gosip.SendToAddress(node.Gate, []byte(err))
			// Log the error
			zap.L().Error("Failed to create database", zap.String("name", name), zap.String("category", category), zap.String("error", err))
			return
		}
		return
	}
}

// GetBroadcasts retrieves messages to be broadcasted to the cluster.
// It limits the number of messages based on the provided overhead and limit.
func (d *GossipDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return d.Queue.GetBroadcasts(overhead, limit)
}

// NodeMeta provides metadata about the local node.
// This implementation returns an empty byte slice as metadata is not used.
func (d *GossipDelegate) NodeMeta(limit int) []byte {
	return []byte("")
}

// LocalState provides the local node's state to other nodes during a join.
// This implementation returns an empty byte slice as state sharing is not used.
func (d *GossipDelegate) LocalState(join bool) []byte {
	return []byte("")
}

// MergeRemoteState merges the state received from a remote node.
// This implementation does nothing as state sharing is not used.
func (d *GossipDelegate) MergeRemoteState(buf []byte, join bool) {
	// No operation
}

// Message represents a message to be broadcasted in the cluster.
type Message struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

// ParseMessage deserializes a JSON byte slice into a Message.
// It returns the parsed message and a boolean indicating success.
func ParseMessage(data []byte) *Message {
	msg := new(Message)
	_ = sonic.Unmarshal(data, msg)
	return msg
}

// Start initializes the memberlist and starts the gossip protocol.
// It creates a channel for incoming messages and sets up the event delegate.
// The memberlist is configured with the provided address and name.
func Start(addr string, name string, port int) {
	// Create a channel for incoming messages
	messageChannel := make(chan []byte)

	// Initialize delegate for message handling
	gossipDelegate = &GossipDelegate{
		MessageChannel: messageChannel,
		Queue:          &memberlist.TransmitLimitedQueue{},
	}
	// Configure the memberlist node
	config := memberlist.DefaultLocalConfig()
	config.Name = name // Avoid name conflicts
	config.BindAddr = addr
	config.BindPort = port
	config.AdvertiseAddr = addr
	config.Delegate = gossipDelegate

	// Create the memberlist
	node.Gosip, err = memberlist.Create(config)
	if err != nil {
		zap.L().Fatal("Failed to create memberlist", zap.Error(err))
	}
}

func Join(addr string) {}
