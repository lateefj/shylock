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

PathIOC 
```````
.. code-block:: bash

  umount /mnt/a; rm -f shylock; go build; env IOC_FILE=/tmp/shylock.csv PATHIOC_DIR=/mnt/b ./shylock pathioc /mnt/a

With this csv as an example:

.. code-block:: text

  /mnt/b/foo/foo/,1000,2,2
  /mnt/b/bar/foo/,2000,3,3
  /mnt/b/bar/bar/,3000,4,4

Kafka 
`````
.. code-block:: bash

  umount $HOME/mnt/localhost; rm -f shylock; go build; ./shylock kafka $HOME/mnt/localhost


