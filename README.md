# SSH Proxy

A secure SSH proxy that sits between SSH clients and an upstream SSH server, logging all client activity and optionally analyzing session security using AI.

## Quick Start (Docker)


>> When you create the config, make sure that you set the parameters in the config before moving ahead
```bash
# Clone the repository
git clone https://github.com/devashar13/ssh-proxy.git
cd ssh-proxy

# Create your config.yaml from the example
cp configs/config.example.yaml configs/config.yaml

# Generate required SSH keys
chmod +x setup-keys.sh
./setup-keys.sh

# Start the Docker environment (includes both proxy and upstream SSH server)
docker-compose up -d

# View logs
docker-compose logs -f
```

## Connect to the Proxy

### Using password authentication:
```bash
ssh -p 2022 user1@localhost
# Enter password as set in the config
```

### Using key-based authentication:
```bash
# Use the key generated by setup-keys.sh
ssh -i test_key -p 2022 keyuser@localhost
```

## Features

- Acts as an intermediary between SSH clients and an upstream SSH server
- Logs all client input to timestamped files
- Handles terminal resizing and special characters correctly
- Supports both password and public key authentication
- Optional Security analysis using OpenAI's API

## Setup and Configuration

### 1. Create Configuration File

The repository includes a sample configuration file. You must create your own `config.yaml` file:

```bash
# Copy the example config
cp configs/config.example.yaml configs/config.yaml

# Edit as needed
vim configs/config.yaml
```

### 2. Generate Required SSH Keys

We provide a script that sets up all required keys:

```bash
chmod +x setup-keys.sh
./setup-keys.sh
```

This script:
- Creates the SSH host key for your proxy server
- Generates a test key pair for authenticating with key-based authentication
- Sets proper permissions on all keys
- Configures the authorized_keys file

### 3. Configure the Proxy

Edit your `configs/config.yaml` file to customize your setup:

```yaml
# Server configuration
server:
  port: 2022                               # The port the proxy listens on
  host_key_path: "./configs/ssh_host_ed25519_key"  # Created by setup-keys.sh

# Upstream SSH server (where connections are forwarded)
upstream:
  host: "ssh-server"                       # Docker service name from docker-compose.yml
  port: 22                                 # Default SSH port in the container
  username: "admin"                        # Username on the upstream server
  auth:
    type: "password"                       # Authentication type for upstream
    password: "admin_password"             # Password for the upstream server

# Users allowed to connect to your proxy
users:
  - username: "user1"                      # Password-based user
    auth:
      type: "password"
      password: "user1pass"
  - username: "keyuser"                    # Key-based user
    auth:
      type: "publickey"
      key_path: "./configs/authorized_keys"  # Created by setup-keys.sh

# Where to store session logs
logging:
  directory: "./logs"

# LLM security analysis
llm:
  enabled: true                           # Set to true to enable
  api_key: ""                              # Your OpenAI API key
  provider: "openai"
  model: "gpt-3.5-turbo"                   # or "gpt-4" for better analysis
```

### 4. Docker Management

```bash
# Start both the proxy and upstream SSH server
docker-compose up -d

# View logs from both containers
docker-compose logs -f

# Stop all containers
docker-compose down
```

## Testing Different Authentication Methods

### Password Authentication

The proxy supports standard username/password authentication:

```bash
# Connect with password auth
ssh -p 2022 user1@localhost
# Enter the password when prompted
```

### Public Key Authentication

For key-based authentication (more secure):

```bash
# Connect using the key generated by setup-keys.sh
ssh -i test_key -p 2022 keyuser@localhost
```

## Session Logs and Security Analysis

### Viewing Session Logs

All session logs are stored in the `logs` directory with format `username_timestamp.log`:

```bash
# List all session logs
ls -la logs/

# View a specific log file
cat logs/user1_20250310-140839.log
```

### Security Analysis (Optional)

If you enable the LLM integration, each session will be analyzed for security risks:

1. Enable the feature in `configs/config.yaml`:
   ```yaml
   llm:
     enabled: true
     api_key: "your-openai-api-key"
     provider: "openai"
     model: "gpt-4"  # Recommended for security analysis
   ```

2. After each session ends, a summary is generated and saved alongside the original log:
   ```bash
   cat logs/user1_20250310-140839.log.summary
   ```

## Troubleshooting

### SSH Host Key Verification Issues

If you see a warning about host key changes when connecting:

```
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
```

Simply remove the old host key from your known_hosts file:

```bash
ssh-keygen -R "[localhost]:2022"
```

Then try connecting again.

This was based on an issue i was facing while testing, should not occur in case you delete the host key and re-generate it without removing it from the known-hosts 
