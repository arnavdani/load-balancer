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

## usage


## other concerns

1. Serverside in python? -> Yes, the GIL sucks. However, it is at its worst when multiple threads share a resource. Take a look at the backend code here; there are no resources being shared between any threads here. Therefore, a single cpu core + all possible threads are enough. Used python for backend its the easiest to debug (ChatGPT wrote it + I fixed its mistakes).