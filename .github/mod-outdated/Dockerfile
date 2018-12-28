FROM golang:alpine

LABEL "name"="Go outtdated modules check"
LABEL "maintainer"="Aleksandr Dorofeev <aka.spn@gmail.com>"
LABEL "version"="1.0.0"

LABEL "com.github.actions.name"="Go outtdated modules check"
LABEL "com.github.actions.description"="Checks outtdated modules."
LABEL "com.github.actions.icon"="package"
LABEL "com.github.actions.color"="blue"

COPY README.md /
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
