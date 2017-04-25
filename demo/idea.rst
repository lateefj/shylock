=========
Demo Idea
=========

The basic idea is to create an IRC web based system that uses etcd for configuration, redis for pub / sub and Postgresql QOS. Demo using etcd to configuration multiple instances of a Go application that uses websockets for basic chat. It also uses redis pub/sub for sending messages to specific rooms. Single IRC example we can see how etcd configuration for the application and redis pub / sub message queue as just a file. Also query the database for the room. Then we create 1K databases from listing on ports 5432 - 6432. Using a simple python script to generate all the configuration files will all standard library packages. This script will also make calls to the REST QOS endpoint to populate all the directories. Then all the directories will be inited and processes started in the background. Finally run a load tool that connects to the API of the APP and for every 1K rooms it tries to send the same amount of messages. 

Goals
-----

* Etcd integration is clear how it works
* Redis MQ integration is very simple to use for pub / sub
* Path QOS using many databases

Tasks
-----

App Configuration 
"""""""""""""""""

#. Start a VM
#. Copy shylock to vm
#. Run script that mounts etcd as readonly
#. Based on hostname run script in etcd (/mnt/etcd/$HOST/setup.sh)

   #. Downloads go application
   #. Copies systemd configuration
   #. Mounts redis
   #. Starts go Application pointing to config (/mnt/etcd/app_name/config.json)

Configuration Generation
""""""""""""""""""""""""

#. Mount etcd as writable
#. App Generation

   #. Takes list of hosts to create
   #. Generates script to start hosts

      #. Sets up each host instance
#. Database configuration

   #. Generates database configuration
   #. Takes a size for the number to generate
   #. Takes a base path to generate configuration in
   #. Generates the QOS based on params

Database Configuration
""""""""""""""""""""""

#. Start a vm
#. Copy shylock to vm
#. Runs etcd as readonly
#. Setup databases for all the folders in a /mnt/etcd/pg/

   #. Run the script that should

      #. Create the directory
      #. Init the directory with dbinit
      #. Start a postgresql process with the configuration / port (systemd?)

Display
-------

Show IRC basic app interface with javascript. Have a way to count the total number of messages.

Some interface that can show the number of messages for all 1K chat rooms to show how the QOS limited writes.

