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

  umount $HOME/mnt/a; rm -f shylock; go build; env PATHIOC_DIR=$HOME/mnt/b ./shylock pathioc $HOME/mnt/a


Kafka 
`````
.. code-block:: bash

  umount $HOME/mnt/localhost; rm -f shylock; go build; ./shylock kafka $HOME/mnt/localhost


