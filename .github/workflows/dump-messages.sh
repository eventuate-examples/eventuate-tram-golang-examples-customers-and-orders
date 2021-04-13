#! /bin/bash

set -e

echo 'select * from eventuate.message\G' | ./mysql-cli.sh -i
echo 'select * from eventuate.received_messages\G' | ./mysql-cli.sh -i
