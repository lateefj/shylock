QOS_FILE=/tmp/qos.csv PATHQOS_DIR=/mnt/b shylock pathqos /mnt/a

mkdir -p /mnt/a/foo/

touch /mnt/a/foo/monkey

dd if=/dev/zero of=/mnt/a/foo/monkey bs=32 count=10

dd if=/dev/zero of=/mnt/a/foo/monkey bs=64 count=10

dd if=/dev/zero of=/mnt/a/foo/monkey bs=128 count=10


mkdir -p /mnt/a/bar/

touch /mnt/a/bar/monkey

dd if=/dev/zero of=/mnt/a/bar/monkey bs=32 count=10

dd if=/dev/zero of=/mnt/a/bar/monkey bs=64 count=10

dd if=/dev/zero of=/mnt/a/bar/monkey bs=128 count=10


qos.csv

"/mnt/b/foo/",1024,1024

"/mnt/b/bar/",2048,2048

