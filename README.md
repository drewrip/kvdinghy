# Dinghy :boat:


### What is this?
This is an implemenation of the [hashicorp/raft](https://github.com/hashicorp/raft) Raft library. At a fundemental level this dinghy allows you to start a Raft server, that enables multiple nodes to maintain distributed consensus.

### Start Dinghy


Starting the leader(first) node

```
./dinghy -r 7000 -h 8000 --bootstrap
```

Starting the other nodes

```
./dinghy -r 7001 -h 8001 --join="127.0.0.1:8000"
```

* It is important to note that if you want your nodes to have the ability of forwarding ```/set``` requests to the leader, you need to make the http port of your nodes 1000 greater than your raft port.

### Interact with your cluster

For this we will be demonstrating using ```curl```. All requests should be able to be made to any node given you have correctly configured your ports, but our examples will just use the address of the leader node.

**Get a key**

Gets the value of a given key:

```
curl -d '{"key": "keyname"}' -X GET http://localhost:8000/get
```

**Set a value**

Set the value of a given key:

```
curl -d '{"key": "keyname", "value": someinteger}' -X POST http://localhost:8000/set
```

**Dump**

Responds with the entire key-value store in JSON form:

```
curl -X GET http://localhost:8000/dump
```


#### Thanks
Thanks to [jen20](https://github.com/jen20) for writing the [barebones implementation](http://github.com/jen20/hashiconf-raft) this is based on.
