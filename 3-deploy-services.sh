#!/bin/bash
export NOMAD_ADDR=https://localhost:4646
export NOMAD_SKIP_VERIFY=True

nomad run deployments/*
