#######
shylock
#######

Stateful Isolation Idea that provides scale down isolation for IO



Name
====

In Shakespearean time [Shylock](https://en.wikipedia.org/wiki/Shylock) common meaning was "white hair" which I am going to ~= meaning "gray beard" in Unix. This is the inspiration for the name.


Development Notes
=================

* When testing it is important to remember not keep remounting the same paths
* Kafka defaults environment variables KAFKA_BROKERS=127.0.0.1:9092 and KAFKA_TOPIC="my_topic"
 
Usage
------

Etcd
````

.. code-block:: bash

   umount $HOME/tmp/localhost/etcd; rm -f shylock; go build; ./shylock etcd $HOME/tmp/localhost/etcd/

   # Mount as read only
   umount $HOME/tmp/localhost/etcd; rm -f shylock; go build; env ETC_READ_ONLY="true" ./shylock etcd $HOME/tmp/localhost/etcd/



PathIOC 
```````
.. code-block:: bash

  umount /mnt/a; rm -f shylock; go build; env IOC_FILE=/tmp/shylock.csv PATHIOC_DIR=/mnt/b ./shylock pathqos /mnt/a

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
