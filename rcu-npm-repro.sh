#!/bin/bash
 
concurrent_runs=10
interval=1
iterations=20

start_docker_containers() {
  for i in $(seq 1 $concurrent_runs); do
    docker run telescope.azurecr.io/issue-repro/zombie:v1.1.11 &
  done
}
 
for ((i=1; i<=iterations; i++)); do
  start_docker_containers
  sleep $interval
done
