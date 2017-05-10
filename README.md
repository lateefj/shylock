

# shylock

_"we must treat the data center itself as one massive warehouse-scale computer"_ - Luiz André Barroso

At the core I created shylock because I feel there is little sanity programming in a modern distributed system. To get more sanity it was going to require decoupling in some sepcific places, integrated with enough systems and provide configuration for quality of service.

1. Everything is a file (file system): key / value (object stores), message queues ect
   1. Examples: S3, Etcd, redis pub/sub, kafka ect
1. Quality of Service (QOS): given a key or path limit the resource based on read/write bytes or operations
   1. Example: User (/mnt/lhj) can only write 1024K per second, read 512K per second

This would not be possible if it wasn't for the great work done by the [bazil/fuse](https://bazil.org/fuse/) project. By making writing a [fuse](https://github.com/libfuse/libfuse) file system in Go amazingly easy I can't thank the developers enough. 

## Status

Currently shylock is in Proof of Concept (POC). The next step is to get enough integration's to be useful for users. Before an integration goes to beta status it will need:

* QOS if applicable
* Unit Tests
* Functional Tests
* Benchmarks

### Integrated Systems

| Name           | Type          | QOS | Status | Notes |
| -------------- |:-------------:|:---:|:------:| :---- |
| Local Path     | File System   | Yes | POC    | Mounts a local file system directory to provide QOS |
| Etcd           | Key Value     | No  | POC    | Low footprint distributed key value store. Basically configuration store for microservices |
| Redis MQ       | Message Queue | No  | POC    | Simple Pub / Sub Message queue system |
| Kafka          | Message Queue | No  | POC    | Distributed streaming system |
| AWS S3         | Object Store  | No  | Idea   | Distribtued Object Store |
| Google Storage | Object Store  | No  | Idea   | Distribtued Object Store |
| AWS SQS        | Message Queue | No  | Idea   | AWS Message Queue  |
| Zookeeper      | Key Value     | No  | Idea   | Heavy footprint key value |
| Google Drive   | Object Store  | No  | Idea   | Consumer Object Store |


## Name

In Shakespearean time [Shylock](https://en.wikipedia.org/wiki/Shylock) common meaning was "white hair" which I am going to ~= meaning "gray beard" in Unix. This is the inspiration for the name.


## Development Notes

* Kafka defaults environment variables KAFKA_BROKERS=127.0.0.1:9092 and KAFKA_TOPIC="my_topic"
* Kafka and or Zookeeper will probably crash or heavily abuse resources don't run it always in the background
  * Zookeeper has used up all my disk space
  * Kafka eventually staves out my VM's without it sending or receiving any messages
 
### Usage

#### Etcd

  _For more details [shylock etcd docs](docs/etcd.rst)_

```bash
shylock etcd /mnt/localhost/etcd/

# Mount as read only

ETC_READ_ONLY="true" shylock etcd /mnt/localhost/etcd/
```

### Redis

Redis currently just uses pubsub message system. This makes it very easy to broadcast messages out.

```bash
shylock redis /mnt/localhost/redis/
```
Environment variables with the defaults:  REDIS_HOST=localhost:6379 REDIS_PASSWORD="" REDIS_DB=0 shylock redis /mnt/localhost/redis/

TODO:

* Key / Value store
* Task queue system

####  PathQOS 

```bash
IOC_FILE=/tmp/shylock.csv PATHQOS_DIR=/mnt/b shylock pathqos /mnt/a
```

With this csv as an example:

```
  /mnt/b/foo/foo/,1000,2,2
  /mnt/b/bar/foo/,2000,3,3
  /mnt/b/bar/bar/,3000,4,4
```

#### Kafka 

  shylock kafka $HOME/mnt/localhost


### Rest API Examples

Create a new path configuration:


  ```
  curl -H "Content-Type: application/json" -X POST -d '{"key":"/home/lhj/mnt/b/foo/monkey/","duration":1000,"read_limit":10,"write_limit":10}' http://localhost:7070/key/home/lhj/mnt/b/foo/monkey/

  curl http://localhost:7070/key/home/lhj/mnt/b/foo/monkey/

  {"key":"/home/lhj/mnt/b/foo/monkey/","duration":1000,"read_limit":10,"write_limit":10}
   ```
Update a configuration:

   ```
   curl -H "Content-Type: application/json" -X PUT -d '{"key":"/home/lhj/mnt/b/foo/monkey/","duration":1000,"read_limit":20,"write_limit":20}' http://localhost:7070/key/home/lhj/mnt/b/foo/monkey/

   http://localhost:7070/key/home/lhj/mnt/b/foo/monkey/

   {"key":"/home/lhj/mnt/b/foo/monkey/","duration":1000,"read_limit":20,"write_limit":20}
```
