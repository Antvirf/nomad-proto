namespace = "default"
path      = "nomadops/v1/gitrepository/testrepo"

items {
  url               = "https://github.com/Antvirf/nomad-proto"
  branch            = "main"
  relative_path     = "gitops-controller-draft"
  regex_path_filter = "job-.*.nomad.hcl"
  controller_name   = "nomadops"
  recurse           = false // WIP - currently does nothing even if true
}

