#!/bin/sh

if [ -d /gitssh-secret ]; then
    install -d -m 0755 -o katafygio -g katafygio /home/katafygio/.ssh/
    install -D -m 0400 -o katafygio -g katafygio /gitssh-secret/* /home/katafygio/.ssh/
fi

su-exec katafygio /usr/bin/katafygio "$@"
