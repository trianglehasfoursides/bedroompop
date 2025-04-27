package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

var (
	gate           *memberlist.Memberlist
	gossipDelegate *GossipDelegate
	err            error
)

// GossipDelegate handles message broadcasting and receiving in the cluster.
type GossipDelegate struct {
	MessageChannel chan []byte
	Queue          *memberlist.TransmitLimitedQueue
}

// NotifyMsg is called when a message is received from the cluster.
// It forwards the message to the message channel.
// TODO: "Implement message handling logic here."
func (d *GossipDelegate) NotifyMsg(msg []byte) {
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

// EventDelegate handles cluster events such as node join, leave, and update.
type EventDelegate struct{}

// NotifyJoin is called when a node joins the cluster.
func (e *EventDelegate) NotifyJoin(node *memberlist.Node) {
	zap.L().Info("Node joined the cluster", zap.String("node", node.Name), zap.String("address", node.Address()))
}

// NotifyLeave is called when a node leaves the cluster.
func (e *EventDelegate) NotifyLeave(node *memberlist.Node) {
	zap.L().Info("Node left the cluster", zap.String("node", node.Name), zap.String("address", node.Address()))
}

// NotifyUpdate is called when a node in the cluster is updated.
func (e *EventDelegate) NotifyUpdate(node *memberlist.Node) {
}

// Start initializes the memberlist and starts the gossip protocol.
// It creates a channel for incoming messages and sets up the event delegate.
// The memberlist is configured with the provided address, name, and port.
func Start(addr string, name string, port int, pipe chan ChanMessage) {
	// Create a channel for incoming messages
	messageChannel := make(chan []byte)

	// Initialize the GossipDelegate for message handling
	gossipDelegate = &GossipDelegate{
		MessageChannel: messageChannel,
		Queue: &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				if len(gate.Members()) > 0 {
					return gate.NumMembers()
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
	gate, err = memberlist.Create(config)
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
		for s := range signal_chan {
			log.Printf("signal %s happen", s.String())
			cancel()
		}
	}(cancel)

	// Start processing incoming messages
	for {
		select {
		case msg := <-pipe:
			// Forward the message to the channel for processing
			h := <-msg.Pipe
			switch h.Key {
			case "list_db":
				// Handle listing databases
			case "create_db":
				// Handle creating a database
			}
		case <-stopCtx.Done():
			zap.L().Info("Stopping gossip protocol")
			// Handle graceful shutdown
			return
		}
	}
}

func Join(addr string) {
	// Join the cluster using the provided address
	_, err = gate.Join([]string{addr})
	if err != nil {
		zap.L().Fatal("Failed to join cluster", zap.Error(err))
	}
	zap.L().Info("Cluster joined successfully", zap.String("address", addr))
}
