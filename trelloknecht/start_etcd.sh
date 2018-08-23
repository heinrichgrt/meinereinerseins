#!/bin/bash
set -e

DATA_DIR=$(pwd)\etcd_data_dir
export MASTER=$(ifconfig -a |grep inet  | grep -v fe80  | grep -v 127.0.0.1 | grep -v inet6 | tr -s "[:blank:]" | tr -d "\t" | cut -f 2 -d" ")
docker run \
  -p 2379:2379 \
  -p 2380:2380 \
  --volume=${DATA_DIR}:/etcd-data \
  --name etcd quay.io/coreos/etcd:latest \
  /usr/local/bin/etcd \
  --data-dir=/etcd-data --name master \
  --initial-advertise-peer-urls http://${MASTER}:2380 --listen-peer-urls http://${MASTER}:2380 \
  --advertise-client-urls http://${MASTER}:2379 --listen-client-urls http://${MASTER}:2379 \
  --initial-cluster MASTER=http://${MASTER}:2380
