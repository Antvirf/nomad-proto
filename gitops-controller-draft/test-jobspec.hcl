namespace = "default"
path      = "nomadops/v1/nomadjob/testjob"

items {
  git_repository_name = "nomadops/v1/gitrepository/testrepo" // refers to the Nomad Variable Path of the GitRepository
  controller_name     = "nomadops"
  spec                = "" // WIP
  status              = "" // WIP
}
