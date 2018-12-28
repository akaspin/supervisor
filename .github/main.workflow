action "action a" {
  uses = "./.github/mod-outdated/"
}

workflow "test" {
  on = "push"
  resolves = ["Go outtdated modules check"]
}

action "Go outtdated modules check" {
  uses = "./.github/mod-outdated"
}
