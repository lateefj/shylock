KafkaFS
=======

Kafka has a great streaming architecture. However I have been very frustrated at the interface to accessing it (Java, JVM scala?). 

Partition Interface
-------------------

In Unix a pipe or even a file seems like a great representation to how to interact with Kafka. Each Kafka partition could be a file / pipe for reading and or writing so something like /kafka/topic/partition/start or /kafka/topic/partition/end depending on reading from the start or the end of the partition, also an offset path /kafka/topic/partition/offset/123. 

Cluster Consumer / Consumer Group
---------------------------------

Cluster consumer (consumer group) could also easily be read based on a /kafka/topic/consumer_name/messages that provides a stream of kafka messages. A message would contain a single byte with the size of the payload ex: [size:payload].

Another interface could be lines (/kafka/topic/consumer_name/lines) which would take the message and turn it into lines. Ideally this would be used for log or data that is in a format that is in a lines separated format (csv, logs, ect).
