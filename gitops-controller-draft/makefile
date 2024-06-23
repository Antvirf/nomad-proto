purgeTestVars:
	nomad var purge nomadops/v1/gitrepository/testrepo
	nomad var purge nomadops/v1/nomadjob/testjob

putvars:
	nomad var put -force @test-jobspec.hcl
	nomad var put -force @test-gitrepo.hcl

install:
	go install antvirf/nomadops

deploy:
	nomad job run job-nomadops.nomad.hcl

