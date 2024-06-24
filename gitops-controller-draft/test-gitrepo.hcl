namespace = "default"
path      = "nomadops/v1/gitrepository/testrepo"

items {
  controller_name = "nomadops"

  // For remote usage
  url  = "https://github.com/Antvirf/nomad-proto"
  type = "remote-repository" // or local directory

  // For local usage, e.g. dev - point url any local directory, do not sync from a remote. That local url is then copied to another tmp location for operations
  // url  = "/home/antti/dev/nomad-proto/"
  // type = "local-directory"

  branch = "main"
}

