#!/bin/bash
export NOMAD_ADDR=https://localhost:4646
export NOMAD_SKIP_VERIFY=True

cd deployments
nomad run job-traefik.nomad.hcl
nomad run job-monitoring.nomad.hcl

cd "$(dirname "$0")"
