package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
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
