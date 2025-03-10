#!/bin/bash
# Generate SSH keys for testing

mkdir -p configs
ssh-keygen -t ed25519 -f configs/ssh_host_ed25519_key -N ""
ssh-keygen -t ed25519 -f test_key -N ""
cat test_key.pub > configs/authorized_keys

# Set proper permissions
chmod 600 configs/ssh_host_ed25519_key
chmod 600 test_key
chmod 600 configs/authorized_keys
