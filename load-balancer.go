package main

import (
	"log"
	"net"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

//// DEFINING INSTANCE OF BACKEND SERVER

type BackendServer struct {
	URL      *url.URL
	is_alive bool
	r_proxy  *httputil.ReverseProxy
	mutex    sync.RWMutex
}

// writes state of backend server
// explicitly uses write lock
func (backend *BackendServer) setAlive(isalive bool) {
	backend.mutex.Lock()
	backend.is_alive = isalive
	backend.mutex.Unlock()
}

func (backend *BackendServer) isAlive() bool {
	backend.mutex.RLock()
	is_alive := backend.is_alive
	backend.mutex.RUnlock()
	return is_alive
}

//// DEFINING SERVER POOL

// rotating index to send requests to
type ServerList struct {
	backend_list []*BackendServer
	backend_map  map[*url.URL]uint64
	index        uint32
}

// updates a server's alive/timed out status via server list
func (sl *ServerList) UpdateServerStatus(url *url.URL, is_alive bool) {
	index := sl.backend_map[url]
	sl.backend_list[index].setAlive(is_alive)
}

// gets the "next" alive server
// for simplicity, next is defined as the closest possible server
//
//	to the right of the current server index as maintained by
//	the backend list.
//
// Updates the serverlist index iff a match is found
// returns the server matched
func (sl *ServerList) GetNextAliveServer() *BackendServer {
	serverlist_len := uint32(len(sl.backend_list))
	for i := sl.index + 1; i < sl.index+serverlist_len; i++ {
		next := i % serverlist_len
		if sl.backend_list[next].isAlive() {
			sl.index = next
			return sl.backend_list[next]
		}
	}
	return nil
}

// send heartbeats to get status of current server
// return true if server accepts ping-ack aka is alive
//
//	otherwise false
func (bs *BackendServer) send_ping_ack() bool {
	timeout := 1 * time.Second
	dial := net.Dialer{Timeout: timeout}
	conn, err := dial.Dial("tcp", bs.URL.Host)
	if err != nil {
		log.Printf("Server %s down - error: %s", bs.URL.String(), err.Error())
		return false
	}
	defer conn.Close()
	return true
}

// check and update status of all servers
func (sl *ServerList) pool_ping_acks() {
	for i, server := range sl.backend_list {
		is_alive := server.send_ping_ack()
		server.setAlive(is_alive)
		log.Printf("Server %d : %s status: %t", i, sl.backend_list[i].URL.String(), is_alive)
	}
}
