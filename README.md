# KVRaft [![Go Report Card](https://goreportcard.com/badge/github.com/tagniam/KVRaft)](https://goreportcard.com/report/github.com/tagniam/KVRaft)
**KVRaft** is a fault-tolerant distributed key-value store, built on the [Raft consensus algorithm](https://raft.github.io/).

## Demo
![](demo.gif)

## Background
KVRaft uses the Raft protocol to implement a replicated key-value store that can withstand node failure and network partitions, exposed through a REST API on the client side.

Raft is a distributed consensus algorithm that manages replicated logs for finite state machines (such as key-value stores). The algorithm handles *leader election*, *log replication*, and *safety* to ensure data survives in case of node failure. More information on Raft can be found at the [wonderful paper written by Diego Ongaro and John Ousterhout](https://raft.github.io/raft.pdf).


## Usage
### Building & Running
The service starts with 3 Raft nodes operating in individual Docker containers. To bring them up, run Docker Compose:
```sh
$ docker-compose up --build
```

### API
From here, the service is exposed through a REST API on `localhost:3000`, that supports two operations:
```http
GET /kvraft/:key
```

and

```http
POST /kvraft/:key
```

where the `POST` request body is in the form `{"value": "<your-value-here>"}`.

Sample usage of the API:
```sh
$ curl -XPOST -d '{"value":"baz"}' http://localhost:3000/kvraft/foo
{"success": true}
$ curl http://localhost:3000/kvraft/foo
{"value": "baz"}
```

### Tests
The project is also bundled with integration tests that simulate a network of Raft nodes under various conditions (concurrent clients, follower/leader crashes, partitions, etc.). To run the tests, run `go test` on the `raft` and `kvraft` packages:

```sh
$ go test raft
ok  	raft	159.893s
$ go test kvraft
ok  	kvraft	214.245s
```

## Acknowledgements
This project is adapted from Princeton's [COS 418 Distributed Systems](https://www.cs.princeton.edu/courses/archive/fall19/cos418/schedule.html) course, which in turn is adapted from MIT's [6.824 Distributed Systems](https://pdos.csail.mit.edu/6.824/) course. Many thanks to the instructors for the amazing course material.
