#######
shylock
#######

Stateful Isolation Idea that provides scale down isolation for IO



Name
====

In Shakespearean time [Shylock](https://en.wikipedia.org/wiki/Shylock) common meaning was "white hair" which I am going to ~= meaning "gray beard" in Unix. This is the inspiration for the name.


Development Notes
=================


Freebsd Testing
---------------

Kafka 
`````

.. code-block:: bash

  sudo umount /dev/fuse; rm -f shylock; go build; sudo ./shylock kafka $HOME/mnt 


