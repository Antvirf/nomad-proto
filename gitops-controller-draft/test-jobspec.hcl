namespace = "default"
path      = "nomadops/v1/nomadjobgroup/testjob"

items {
  controller_name     = "nomadops"
  git_repository_name = "nomadops/v1/gitrepository/testrepo" // refers to the Nomad Variable Path of the GitRepository
  relative_path       = "gitops-controller-draft"
  regex_path_filter   = "job-.*.nomad.hcl"
  recurse             = false // WIP - currently does nothing even if true
  spec                = ""    // WIP
  status              = ""    // WIP
}
