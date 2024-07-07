package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

const (
	Tries int = 1
)

// for http context reasons

//// DEFINING INSTANCE OF BACKEND SERVER

type BackendServer struct {
	URL     *url.URL
	alive   bool
	r_proxy *httputil.ReverseProxy
	mutex   sync.RWMutex
}

// writes state of backend server
// explicitly uses write lock
func (backend *BackendServer) set_alive(isalive bool) {
	backend.mutex.Lock()
	backend.alive = isalive
	backend.mutex.Unlock()
}

func (backend *BackendServer) is_alive() bool {
	backend.mutex.RLock()
	alive := backend.alive
	backend.mutex.RUnlock()
	return alive
}

//// DEFINING SERVER POOL

// rotating index to send requests to
type ServerList struct {
	backend_list []*BackendServer
	backend_map  map[*url.URL]uint64
	index        uint32
	mutex        sync.RWMutex
}

// updates a server's alive/timed out status via server list
func (sl *ServerList) update_server_status(url *url.URL, is_alive bool) {
	index := sl.backend_map[url]
	sl.backend_list[index].set_alive(is_alive)
}

func (sl *ServerList) add_server(b *BackendServer) {
	sl.backend_list = append(sl.backend_list, b)
}

// gets the "next" alive server
// for simplicity, next is defined as the closest possible server
// to the right of the current server index as maintained by
// the backend list.
//
// Updates the serverlist index iff a match is found
// returns the server matched
func (sl *ServerList) get_next_alive_server() *BackendServer {
	sl.mutex.RLock()
	serverlist_len := uint32(len(sl.backend_list))
	sl_index := sl.index
	sl.mutex.RUnlock()
	for i := sl_index + 1; i < sl_index+serverlist_len; i++ {
		next := i % serverlist_len
		if sl.backend_list[next].is_alive() {
			sl.mutex.Lock()
			sl.index = next
			sl.mutex.Unlock()
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
		server.set_alive(is_alive)
		log.Printf("Server %d : %s status: %t", i, sl.backend_list[i].URL.String(), is_alive)
	}
}

// use http context to store and get the number of tries
//
//	to make the request succeed
//
// if no context value, that means current round in the first try
func get_tries_from_req(req *http.Request) int {
	if tries, ok := req.Context().Value(Tries).(int); ok {
		return tries
	}
	return 1
}

var pool ServerList

// http handler proper
func balance(w http.ResponseWriter, r *http.Request) {
	num_tries := get_tries_from_req(r)
	if num_tries > 5 {
		log.Printf("Request %s from %s failed", r.URL.Path, r.RemoteAddr)
		http.Error(w, "Request still failed after 5 tries", http.StatusRequestTimeout)
	}

	target := pool.get_next_alive_server()
	if target != nil {
		target.r_proxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "No available server", http.StatusServiceUnavailable)
}
