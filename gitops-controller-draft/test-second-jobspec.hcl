namespace = "default"
path      = "nomadops/v1/nomadjobgroup/second-testjob"

items {
  controller_name     = "nomadops"
  git_repository_name = "nomadops/v1/gitrepository/testrepo" // refers to the Nomad Variable Path of the GitRepository

  // Where to find .hcl files that describe Nomad Jobs
  nomad_job_relative_path     = "deployments"
  nomad_job_regex_path_filter = "job-.*.nomad.hcl"

  // Where to find .hcl files that describe NomadJobGroups
  nomad_job_group_relative_path     = "gitops-controller-draft"
  nomad_job_group_regex_path_filter = ".*.-jobspec.hcl"

  spec   = "something" // WIP
  status = "something" // WIP
}
