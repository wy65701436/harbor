#!/bin/bash

set -e

docker pull vmware/harbor-e2e-engine:1.41

make install GOBUILDIMAGE=golang:1.9.2 COMPILETAG=compile_golangimage CLARITYIMAGE=goharbor/harbor-clarity-ui-builder:1.6.0