FROM golang:alpine

RUN apk --no-cache add git

LABEL "name"="Go outtdated modules"
LABEL "maintainer"="Aleksandr Dorofeev <aka.spn@gmail.com>"
LABEL "version"="1.0.0"

LABEL "com.github.actions.name"="Go outtdated modules"
LABEL "com.github.actions.description"="Checks outtdated Go modules"
LABEL "com.github.actions.icon"="package"
LABEL "com.github.actions.color"="blue"

COPY README.md /
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
