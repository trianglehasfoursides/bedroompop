package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/memberlist"
	"github.com/trianglehasfoursides/bedroompop/consist"
)

type MyDelegate struct {
	meta []byte
}

func (d *MyDelegate) NodeMeta(limit int) []byte {
	if len(d.meta) > limit {
		return d.meta[:limit]
	}

	return d.meta
}

func (d *MyDelegate) NotifyMsg([]byte)                           {}
func (d *MyDelegate) GetBroadcasts(overhead, limit int) [][]byte { return nil }
func (d *MyDelegate) LocalState(join bool) []byte                { return nil }
func (d *MyDelegate) MergeRemoteState(buf []byte, join bool)     {}

type NotifyDelegate struct{}

// join while still running
func (n *NotifyDelegate) NotifyJoin(node *memberlist.Node) {
	consist.Consist.Add(consist.Member(node.Meta))
}

func (n *NotifyDelegate) NotifyLeave(node *memberlist.Node)  {}
func (n *NotifyDelegate) NotifyUpdate(node *memberlist.Node) {}

type Gossip struct {
	Node *memberlist.Memberlist
}

func CreateGossip(grpc string, goss string, name string) (gossip *Gossip, err error) {
	delegate := &MyDelegate{meta: []byte(grpc)}

	config := memberlist.DefaultLocalConfig()
	config.Name = name
	config.Delegate = delegate
	config.Events = new(NotifyDelegate)

	gss := strings.Split(goss, ":")
	config.BindAddr = gss[0]
	config.BindPort, _ = strconv.Atoi(gss[1])

	config.AdvertiseAddr = config.BindAddr
	config.AdvertisePort = config.BindPort

	gossip = new(Gossip)
	gossip.Node, err = memberlist.Create(config)
	if err != nil {
		return
	}

	return
}

// first time join
func (g *Gossip) Join(nodes []string) (err error) {
	if _, err = g.Node.Join(nodes); err != nil {
		return
	}

	for _, n := range g.Node.Members() {
		fmt.Println(n.Meta)
		consist.Consist.Add(consist.Member(n.Meta))
	}

	return
}
