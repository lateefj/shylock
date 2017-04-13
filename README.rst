=======
shylock
=======

"we must treat the data center itself as one massive warehouse-scale computer" - Luiz André Barroso

At the core I created shylock because I feel there is little sanity programming in a modern distributed system. Shylock basically provides two core features.

#. Everything is a file (file system): key / value (object stores), message queues ect
   #. Examples: S3, Etcd, redis pub/sub, kafka ect
#. Quality of Service (QOS): given a key or path limit the resource based on read/write bytes or operations
   #. Example: User (/mnt/lhj) can only write 1024K per second, read 512K per second

This would not be possible if it wasn't for the great work done by the `bazil/fuse <https://bazil.org/fuse/>`_ project. By making writing a `fuse <https://github.com/libfuse/libfuse>`_ file system in Go amazingly easy I can't thank the developers enough. 

Status
------

Currently shylock is in Proof of Concept (POC). The next step is to get enough integration's to be useful for users. Before an integration goes to beta status it will need:

* QOS if applicable
* Unit Tests
* Functional Tests
* Benchmarks

Integration`s
"""""""""""""

+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| Name           | Type          | QOS | Status | Notes                                                                                      |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| Local Path     | File System   | Yes | POC    | Mounts a local file system directory to provide QOS                                        |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| Etcd           | Key Value     | No  | POC    | Low footprint distributed key value store. Basically configuration store for microservices |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| Kafka          | Message Queue | No  | POC    | Distributed streaming system                                                               |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| Redis MQ       | Message Queue | No  | Idea   | Simple Pub / Sub Message queue system                                                      |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| AWS S3         | Object Store  | No  | Idea   | Distribtued Object Store                                                                   |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| Google Storage | Object Store  | No  | Idea   | Distribtued Object Store                                                                   |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| AWS SQS        | Message Queue | No  | Idea   | AWS Message Queue                                                                          |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| Zookeeper      | Key Value     | No  | Idea   | Heavy footprint key value                                                                  |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+
| Google Drive   | Object Store  | No  | Idea   | Consumer Object Store                                                                      |
+----------------+---------------+-----+--------+--------------------------------------------------------------------------------------------+



Name
----

In Shakespearean time [Shylock](https://en.wikipedia.org/wiki/Shylock) common meaning was "white hair" which I am going to ~= meaning "gray beard" in Unix. This is the inspiration for the name.


Development Notes
-----------------

* Kafka defaults environment variables KAFKA_BROKERS=127.0.0.1:9092 and KAFKA_TOPIC="my_topic"
* Kafka and or Zookeeper will probably crash or heavily abuse resources don't run it always in the background
  * Zookeeper has used up all my disk space
  * Kafka eventually staves out my VM's without it sending or receiving any messages
 
Usage
------

Etcd
````

.. code-block:: bash

   umount $HOME/tmp/localhost/etcd; rm -f shylock; go build; ./shylock etcd $HOME/tmp/localhost/etcd/

   # Mount as read only
   umount $HOME/tmp/localhost/etcd; rm -f shylock; go build; env ETC_READ_ONLY="true" ./shylock etcd $HOME/tmp/localhost/etcd/



PathQOS 
```````
.. code-block:: bash

  umount /mnt/a; rm -f shylock; go build; env IOC_FILE=/tmp/shylock.csv PATHQOS_DIR=/mnt/b ./shylock pathqos /mnt/a

With this csv as an example:

.. code-block:: text

  /mnt/b/foo/foo/,1000,2,2
  /mnt/b/bar/foo/,2000,3,3
  /mnt/b/bar/bar/,3000,4,4

Kafka 
`````
.. code-block:: bash

  umount $HOME/mnt/localhost; rm -f shylock; go build; ./shylock kafka $HOME/mnt/localhost


Rest API Examples
`````````````````

Create a new path configuration:

.. code-block:: bash

  curl -H "Content-Type: application/json" -X POST -d '{"key":"/home/lhj/mnt/b/foo/monkey/","duration":1000,"read_limit":10,"write_limit":10}' http://localhost:7070/key/home/lhj/mnt/b/foo/monkey/

  curl http://localhost:7070/key/home/lhj/mnt/b/foo/monkey/

  {"key":"/home/lhj/mnt/b/foo/monkey/","duration":1000,"read_limit":10,"write_limit":10}

Update a configuration:

.. code-block:: bash

   curl -H "Content-Type: application/json" -X PUT -d '{"key":"/home/lhj/mnt/b/foo/monkey/","duration":1000,"read_limit":20,"write_limit":20}' http://localhost:7070/key/home/lhj/mnt/b/foo/monkey/

   http://localhost:7070/key/home/lhj/mnt/b/foo/monkey/

   {"key":"/home/lhj/mnt/b/foo/monkey/","duration":1000,"read_limit":20,"write_limit":20}
