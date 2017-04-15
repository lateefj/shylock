============
shylock etcd
============

If we treat the data center like a _Unix_ system first we will need an /etc. etcd_ was created for just this purpose and to quote the website:

  "etcd is a distributed key value store that provides a reliable way to store data across a cluster of machines."

Why
---

 * Simplify managing configuration of a data center
 * Allow for a diversity of different methods of building, generating or creating configuration
 * Configuration consumers don't need a system or special tools to get configuration

Idea
----

Just like configuring a single hosts services in /etc we should store configuration for software and hosts in a distributed store like etcd_. This allows configuration for applications like an cluster of `nginx <https://nginx.org/en/>`_ instances to share the same configuration in one location. 

Why is container configuration so hard?
---------------------------------------

One of the things that I have noticed about containers is that the configuration tends to be built in a combination of build and runtime. Often what happens is a bunch of environment variables are passed into the container at runtime and then unix tools like sed, awk or just echo are used to populate a configuration file. Worse if configuration management tools are installed and or configuration generation stuff (scripts and templates). It seems a better strategy is to generate the configuration, store it in etcd_ and restart the container. Usually there is a new problem which is how to get that configuration out of etcd_ configuration into the docker container. Why not just use the copy (cp) command? If we mount etcd_ on as a directory we can both generate the files and mount them in the container it become really easy to copy configuration on container start. 

Why etcd
--------

So the reason etcd_ is key to a sane way to manage a distributed system. Centrally managing configuration files is a lot more sane. This is also the location for load balancer, database connection information ect. Compared to something like `redis <https://redis.io/>`_ it is replicated so there is not single point of failure. Also it had a couple features compared to `zookeeper <https://zookeeper.apache.org/>`_ in that etcd_ footprint is really small (CPU, memory) and basic usage is trivially easy.

Read only option
----------------

Specifically there is a read only option since the producer of the configuration is probably not also the consumer. There is a ton of different ways to generate configuration from simple to complex. This doesn't tightly couple configuration to requiring using a Puppet, Chef but instead just forces those configuration to be writing out to a file in so that the can be accessed by the action hosts (containers) that need them. 


Examples
-------

Read only::

  ETC_HOSTS="http://etcd-01.lhj.me:2379" shylock etcd /mnt/localhost/etcd/

Writable::

  ETC_HOSTS="http://etcd-01.lhj.me:2379" ETC_READ_ONLY="true" shylock etcd /mnt/localhost/etcd/


.. _etcd: https://coreos.com/etcd
