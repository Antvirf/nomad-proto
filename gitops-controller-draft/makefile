purgeTestVars:
	nomad var purge nomadops/v1/gitrepository/testrepo
	nomad var purge nomadops/v1/nomadjobgroup/testjob

putvars:
	nomad var put -force @manifests/test-jobspec.hcl
	nomad var put -force @manifests/test-gitrepo.hcl

install:
	cd nomad-gitops-operator && go install .

run:
	go run ./nomad-gitops-operator

deploy:
	nomad job run manifests/job-nomadops.nomad.hcl

