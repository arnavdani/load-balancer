# load-balancer

## goals

1. more Go practice
2. implement core backend service
3. learn about proxys, handlers, etc and how go stl supports them
4. scheduling algorithms in load balancers

building a basic load balancer in Go

initially based on this article: https://kasvith.me/posts/lets-create-a-simple-lb-go/

sources used (outside of go stl):
https://hub.docker.com/r/strm/helloworld-http - basis for dockerfile

also stole dockerfile from the first article

## deviations from the article

1. my own backend server - I maintain my own backend servers so that I can relay information that could be useful for balancing tasks.
2. At uptime, the backend servers to balance between are not known at initialization to the load balancer. These servers send pings to the load balancer to notify it of its existence and availability. This is important because this makes an application easy to scale. Now, connections can be accepted on the fly and servers can be added to scale up and scale down infra as needed (instead of keeping it static).
3. Better balancing algorithm - this uses a self designed algorithm based on the Hadoop DRF algorithm. Each load is matched to the best suited backend based on aligning compute and storage requirements. This is done by finding the ratio of compute:storage and matching it. For simluation purposes, the ratios of jobs and the backend servers are backed by RNG.

## usage

prerequisites: Docker, Docker Desktop (ideally), Internet Conn

1. By default, this load balancer can be simulated by running `deploy.sh`. This will deploy the load balancer (defined by `load-balancer.go`) and 3 separate backend servers (all are the same, defined by `backend.py`) to 4 different docker containers (all running Alpine).

Refer to docker-compose for more details - main load balancer is deployed to `localhost:5252` (shoutout Patrick Willis). Python servers are deployed to `:5453` (thank you Fred Warner + Navarro Bowman), `:5454` (Fred Warner x2), `:5455` (Fred Warner, Ahmad Brooks).

To use, visit `localhost:5252` in browser - request will be automatically balanced between the 3 available backends.

To see basic UI about load balanced across all servers, visit `localhost:9494` (shoutout Justin Smith).

## Implementation Details

The 3 backends alert the load balancer of its existence via the server set up on `load-balancer:9797` (shoutout Nick Bosa). In addition, ping-acks are sent every 100s to verify the uptime of all participating backend servers.

The strategy used to balance these loads uses the same rational as the Hadoop DRF scheduling algorithm. While that is designed to schedule multiple jobs on one unit, I will apply the same logic of leveraging the ratios between expected resource demand to best balance the load.

Since all my testing is simulated anyway, I am using rng at initialization to quantify the requirements of the job and the behavior of each backend server.

## other concerns

1. Serverside in python? -> Yes, the GIL sucks. However, it is at its worst when multiple threads share a resource. Take a look at the backend code here; there are no resources being shared between any threads here. Therefore, a single cpu core + all possible threads are enough. Used python for backend its the easiest to debug (ChatGPT wrote it + I fixed its mistakes).

## cool things i learned about (wip)

go sync primitives (rw locks, wait groups)
reverse proxy
http request header, context, request manipulation
nested functions

## next steps (if i ever come back)

spam testing