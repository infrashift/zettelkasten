#!/bin/bash
set -e

SSH_PORT="${SSH_PORT:-2222}"

# Apply custom port
if [[ "$SSH_PORT" != "2222" ]]; then
    sed -i "s/^Port 2222$/Port ${SSH_PORT}/" /etc/ssh/sshd_config
fi

mkdir -p /run/sshd

# Fix .ssh permissions (volume mounts often have wrong ownership)
# Use || true to handle read-only mounts gracefully
if [[ -d /home/user/.ssh ]]; then
    chown -R user:0 /home/user/.ssh 2>/dev/null || true
    chmod 700 /home/user/.ssh 2>/dev/null || true
    [[ -f /home/user/.ssh/authorized_keys ]] && chmod 600 /home/user/.ssh/authorized_keys 2>/dev/null || true
fi

echo "Starting sshd on port ${SSH_PORT}..."
exec /usr/sbin/sshd -D -e
