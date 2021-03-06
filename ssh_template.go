package gitssh

var sshTemplate = `#!/bin/sh
unset SSH_AUTH_SOCK
ssh -o CheckHostIP=no \
    -o IdentitiesOnly=yes \
    -o LogLevel=INFO \
    -o StrictHostKeyChecking=no \
    -o PasswordAuthentication=no \
    -o UserKnownHostsFile={{.HostsFile}} \
    -o IdentityFile={{.KeyFile}} $*
`
