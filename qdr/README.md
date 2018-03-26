# QPID Dispatch Router RPM Build

The `./builtit.sh` script will result in a set of RPMs that can be consumed by
`docker build` to create the `nfvpe/qpid-dispatch-router` container image.

## Requirements

It is expected that `mock` is installed on your system and that your current
user is in the `mock` system group.
