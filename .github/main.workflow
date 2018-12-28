action "action a" {
  uses = "./.github/mod-outdated/"
}

workflow "test" {
  on = "push"
  resolves = ["Go outdated modules"]
}

action "Go outdated modules" {
  uses = "./.github/mod-outdated"
}
