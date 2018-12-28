workflow "check" {
  on = "push"
  resolves = ["Go outdated modules"]
}

action "Go outdated modules" {
  uses = "./.github/mod-outdated"
}
