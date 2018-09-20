CloudWatch Monitor
==================

[![Build Status](https://travis-ci.org/dedalusj/cwmonitor.svg?branch=master)](https://travis-ci.org/dedalusj/cwmonitor) [![codecov](https://codecov.io/gh/dedalusj/cwmonitor/branch/master/graph/badge.svg)](https://codecov.io/gh/dedalusj/cwmonitor)

Monitoring tool to collect basic statistics from a machine and from docker statistics and send them to CloudWatch metrics.

Available metrics for collection:

- CPU
- Memory
- Swap
- Disk
- Docker stats
- Docker health status

# How to

### Binary

Download the binary from the [GitHub release](https://github.com/dedalusj/cwmonitor/releases).

Run it with `./cwmonitor --metrics cpu,memory --interval 60 --namespace a_namespace --hostid "$(hostname)"`

Available metrics are: `cpu, memory, swap, disk, docker-health, docker-stats`.

Use `./cwmonitor --help` to see a description of the other command line arguments.

### Docker

CWMonitor is also available as a docker image and can be run with

    docker run --rm --name=cwmonitor -v /var/run/docker.sock:/var/run/docker.sock \
        dedalusj/cwmonitor --metrics cpu,memory --interval 60 --namespace a_namespace --hostid "$(hostname)"

