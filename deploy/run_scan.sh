#!/bin/bash

go build &&

source ~/.aws/credentials &&

./deploy -add_node 1 &&
./deploy -setup      &&
./deploy -deploy     &&

go build ../scream  &&
./scream -host $(./deploy -server_dns 2>/dev/null) -start 1025 -end 65535 | tee portscan.log &&
./deploy -remove_node 1 &&
rm deploy scream
