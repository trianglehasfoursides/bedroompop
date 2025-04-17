package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	case "delete_db":
		var mutex = new(sync.Mutex)
		var name, category string = gjson.Get(string(message.Value), "name").String(), gjson.Get(string(message.Value), "category").String()
		if err := database.DeleteDatabase(name, category, mutex); err != nil {
			var err = fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
			// Send error message back to the gate
			node.Gosip.SendToAddress(node.Gate, []byte(err))
			// Log the error
			zap.L().Error("Failed to create database", zap.String("name", name), zap.String("category", category), zap.String("error", err))
			return
		}
		return
	case "get_db":
		var name, category string = gjson.Get(string(message.Value), "name").String(), gjson.Get(string(message.Value), "category").String()
		var db, err = database.GetDatabase(name, category)
		if err != nil {
			var err = fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
			// Send error message back to the gate
			node.Gosip.SendToAddress(node.Gate, []byte(err))
			// Log the error
			zap.L().Error("Failed to create database", zap.String("name", name), zap.String("category", category), zap.String("error", err))
			return
		}
		node.Gosip.SendToAddress(node.Gate, []byte(db))
		return
	case "list_db":
		var category string = gjson.Get(string(message.Value), "category").String()
		var db, err = database.ListDatabases(category)
		if err != nil {
			var err = fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
			// Send error message back to the gate
			node.Gosip.SendToAddress(node.Gate, []byte(err))
			// Log the error
			zap.L().Error("Failed to create database", zap.String("category", category), zap.String("error", err))
			return
		}
		listDb, err := sonic.Marshal(db)
		if err != nil {
			var err = fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
			// Send error message back to the gate
			node.Gosip.SendToAddress(node.Gate, []byte(err))
			// Log the error
			zap.L().Error("Failed to create database", zap.String("category", category), zap.String("error", err))
			return
		}
		node.Gosip.SendToAddress(node.Gate, listDb)
		return
	case "update_config":
		var mutex = new(sync.Mutex)
		var newConfig = new(database.DatabaseConfiguration)
		var name, category string = gjson.Get(string(message.Value), "name").String(), gjson.Get(string(message.Value), "category").String()
		var config = gjson.Get(string(message.Value), "config").Value()
		sonic.Unmarshal(config.([]byte), newConfig)
		if err := database.UpdateDatabaseConfiguration(name, category, newConfig, mutex); err != nil {
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
// The memberlist is configured with the provided address, name, and port.
func Start(addr string, name string, port int) {
	// Create a channel for incoming messages
	messageChannel := make(chan []byte)

	// Initialize the GossipDelegate for message handling
	gossipDelegate = &GossipDelegate{
		MessageChannel: messageChannel,
		Queue: &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				if node.Gosip != nil {
					return node.Gosip.NumMembers()
				}
				return 0
			},
			RetransmitMult: 3, // Number of times to retransmit messages
		},
	}

	// Configure the memberlist node
	config := memberlist.DefaultLocalConfig()
	config.Name = name // Set the node name to avoid conflicts
	config.BindAddr = addr
	config.BindPort = port
	config.AdvertiseAddr = addr
	config.Delegate = gossipDelegate

	// Create the memberlist instance
	var err error
	node = &Node{}
	node.Gosip, err = memberlist.Create(config)
	if err != nil {
		zap.L().Fatal("Failed to create memberlist", zap.Error(err))
	}

	// Log successful initialization
	zap.L().Info("Memberlist initialized", zap.String("name", name), zap.String("address", addr), zap.Int("port", port))

	// Context for stopping the gossip protocol
	stopCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Goroutine to handle termination signals
	go func(cancel context.CancelFunc) {
		signal_chan := make(chan os.Signal, 1)
		signal.Notify(signal_chan, syscall.SIGINT)
		for {
			select {
			case s := <-signal_chan:
				log.Printf("signal %s happen", s.String())
				cancel()
			}
		}
	}(cancel)

	// Start processing incoming messages
	run := true
	for run {
		select {
		case <-stopCtx.Done():
			// Stop the gossip protocol
			zap.L().Info("Stopping gossip protocol")
			run = false
		}
	}

	// Cleanup and shutdown
	zap.L().Info("Gossip protocol stopped")
}

func Join(addr string) {
	// Join the cluster using the provided address
	_, err = node.Gosip.Join([]string{addr})
	if err != nil {
		zap.L().Fatal("Failed to join cluster", zap.Error(err))
	}
	zap.L().Info("Cluster joined successfully", zap.String("address", addr))
}
