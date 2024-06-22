#!/bin/bash

# Create certs in case this is the first time
nomad tls ca create
nomad tls cert create -server -region global

sudo nomad agent -dev -config=./config/nomad.hcl
