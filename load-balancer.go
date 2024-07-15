package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

const (
	Tries          = "Tries"
	Compute        = "Compute"
	Storage        = "Storage"
	Compute_Vector = "Compute-Vector"
	Storage_Vector = "Storage-Vector"
)

// for http context reasons

// // DEFINING JOB
type Job struct {
	compute int
	storage int
}

//// DEFINING BACKEND SERVER

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
	backend_map  map[*url.URL]uint32
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
func (sl *ServerList) get_next_alive_server(j *Job) *BackendServer {
	// var dom int
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

// every 100 s, send pingacks to the entire server pool to update status
func refresh_alive_servers() {
	ticker := time.NewTicker(100 * time.Second)

	go func() {
		for {
			select {
			case _ = <-ticker.C:
				pool.pool_ping_acks()
			}
		}
	}()

}

// use http context to store and get the number of tries
//
//	to make the request succeed
//
// if no context value, that means no tries have been made, thus return 0
func get_tries_from_req(req *http.Request) int {
	if tries, ok := req.Context().Value(Tries).(int); ok {
		return tries
	}
	return 0
}

var pool ServerList

// http handler proper
func balance(w http.ResponseWriter, r *http.Request) {
	num_tries := get_tries_from_req(r)
	if num_tries > 5*len(pool.backend_list) {
		log.Printf("Request %s from %s failed", r.URL.Path, r.RemoteAddr)
		http.Error(w, "Request still failed after 5 tries", http.StatusRequestTimeout)
	}

	compute_reqs := rand.Intn(10) + 1
	storage_reqs := rand.Intn(10) + 1
	compute_ctx := context.WithValue(r.Context(), Compute, compute_reqs)
	full_ctx := context.WithValue(compute_ctx, Storage, storage_reqs)
	r = r.WithContext(full_ctx)
	r.Header.Set(Compute, fmt.Sprintf("%d", compute_reqs))
	r.Header.Set(Storage, fmt.Sprintf("%d", storage_reqs))

	job := Job{compute: compute_reqs, storage: storage_reqs}

	target := pool.get_next_alive_server(&job)
	if target != nil {
		target.r_proxy.ServeHTTP(w, r)
		return
	}

	http.Error(w, "No available server", http.StatusServiceUnavailable)
}

func setup_incoming_server(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		log.Fatal(err)
	}

	// cv := r.Header.Get(Compute_Vector)
	// sv := r.Header.Get(Storage_Vector)

	s_url, err := url.Parse(fmt.Sprintf("http://%s:80", host))
	log.Printf("%s:%s\n", host, "80")
	if err != nil {
		log.Fatal(err)
	}

	reverse_proxy := httputil.NewSingleHostReverseProxy(s_url)
	reverse_proxy.ErrorHandler = func(wr http.ResponseWriter, req *http.Request, e error) {
		log.Printf("Error found. Backend host: %s\nError:%s", req.Host, e.Error())
		tries := get_tries_from_req(req)
		if tries < 5 { // increment tries
			ctx := context.WithValue(req.Context(), Tries, tries+1)
			reverse_proxy.ServeHTTP(wr, req.WithContext(ctx))
			return
		}

		//if > 5 tries, fully retry request
		// mark server as down as well
		pool.backend_list[pool.backend_map[s_url]].set_alive(false)
		log.Printf("%s: Retrying request on other backend", req.RemoteAddr)
		ctx := context.WithValue(req.Context(), Tries, tries+1)
		reverse_proxy.ServeHTTP(wr, req.WithContext(ctx))
		balance(wr, req.WithContext(ctx))
	}

	pool.add_server(&BackendServer{
		URL:     s_url,
		alive:   true,
		r_proxy: reverse_proxy,
	})
	log.Printf("New backend server %s\n", s_url)
	fmt.Fprintf(w, "Server accepted\n")
}

func main() {
	var in_port = flag.String("i", "9797", "lb connection acceptor")
	var lb_port = flag.String("o", "5252", "host port of lb")
	flag.Parse()

	// set up wait group
	// process isn't terminated until all entries in wg are marked as done
	var wg sync.WaitGroup

	in_s := http.Server{
		Addr:    *in_port,
		Handler: http.HandlerFunc(setup_incoming_server),
	}

	// start up input server - collect backend incoming
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := in_s.ListenAndServe()

		if err != nil {
			log.Fatal(err)
		}
	}(&wg)

	// start ping acks - assume connections have come in
	go refresh_alive_servers()

	lb_s := http.Server{
		Addr:    *lb_port,
		Handler: http.HandlerFunc(balance),
	}

	// start load balancing server
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := lb_s.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}(&wg)

	fmt.Printf("Both servers started\n")
	wg.Wait()
}
