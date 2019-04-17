package cmd

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/memberlist"
)

var (
	mtx          sync.RWMutex
	AliveMembers []*memberlist.Node
	members      = flag.String("members", "", "comma seperated list of members")
	items        = map[string]string{}
	broadcasts   *memberlist.TransmitLimitedQueue
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type delegate struct{}

type update struct {
	Action string
	Data   map[string]string
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	switch b[0] {
	case 'd':
		var updates []*update
		if err := json.Unmarshal(b[1:], &updates); err != nil {
			return
		}
		mtx.Lock()
		for _, u := range updates {
			for k, v := range u.Data {
				switch u.Action {
				case "add":
					items[k] = v
				case "del":
					delete(items, k)
				}
			}
		}
		mtx.Unlock()
	}
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	mtx.RLock()
	m := items
	mtx.RUnlock()
	b, _ := json.Marshal(m)
	return b
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}

	if !join {
		return
	}

	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	mtx.Lock()
	for k, v := range m {
		items[k] = v
	}
	mtx.Unlock()

}

func StartMemberlist() error {
	hostname, _ := os.Hostname()
	v_ip := CFG.Volume.Ip
	v_port, _ := strconv.Atoi(CFG.Volume.Port)

	g_port := v_port - 1000
	c := memberlist.DefaultLocalConfig()
	c.Delegate = &delegate{}
	c.BindAddr = v_ip
	c.BindPort = g_port
	c.Name = strings.Join([]string{hostname, v_ip, strconv.Itoa(g_port)}, "-")

	m, err := memberlist.Create(c)
	if err != nil {
		return err
	}

	if len(*members) > 0 {
		parts := strings.Split(*members, ",")
		_, err := m.Join(parts)
		if err != nil {
			return err
		}
	}
	broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return m.NumMembers()
		},
		RetransmitMult: 3,
	}
	node := m.LocalNode()
	Logger.Printf("Local member %s:%d\n", node.Addr, node.Port)

	AliveMembers = m.Members()
	for _, member := range m.Members() {
		Logger.Printf("Member: %s %s\n", member.Name, member.Addr)
	}
	return nil
}

func PrintMemberList() error {

	for _, member := range AliveMembers {
		Logger.Printf("Member: %s %s %s\n", member.Name, member.Addr, member.Port)
	}
	return nil
}
