# lambdactl

CLI for [Lambda AI](https://lambda.ai/) cloud GPU instances.

## Install

```
go install github.com/dims/lambdactl@latest
```

Or build from source:

```
git clone https://github.com/dims/lambdactl.git
cd lambdactl
make build        # binary in bin/lambdactl
make install      # installs to $GOPATH/bin
```

## Authentication

lambdactl looks for an API key in this order:

1. `LAMBDA_API_KEY` environment variable
2. File at path in `LAMBDA_API_KEY_FILE` environment variable
3. `~/.config/lambda/.key` (default)

Get your API key from the [Lambda dashboard](https://cloud.lambdalabs.com/api-keys).

```
mkdir -p ~/.config/lambda
echo "your-api-key" > ~/.config/lambda/.key
```

Verify:

```
$ lambdactl
API key is valid.
```

## Usage

### List GPU types

```
$ lambdactl types
NAME                   DESCRIPTION              $/HR    VCPUS  RAM     DISK     REGIONS
gpu_1x_a100_sxm4       1x A100 (40 GB SXM4)     $1.48   30     200GB   512GB    1 available
gpu_1x_h100_sxm5       1x H100 (80 GB SXM5)     $3.78   26     225GB   2816GB   0 available
gpu_8x_b200_sxm6       8x B200 (180 GB SXM6)    $45.92  208    2900GB  22528GB  0 available
...
```

### Launch an instance

```
$ lambdactl start -g gpu_1x_a100_sxm4 -s my-ssh-key
Launched instance f34fea36... in us-west-2. Waiting for IP...
  status: booting
  status: booting
Ready! ssh ubuntu@161.153.31.243
```

### Watch for availability and auto-launch

```
$ lambdactl watch -g gpu_1x_h100_sxm5 -s my-ssh-key --region us-east-1
Watching for gpu_1x_h100_sxm5 availability (every 10s)...
  [14:20:15] no availability
  [14:20:25] no availability
  [14:20:35] found in us-east-1! Launching...
```

### SSH into an instance

```
$ lambdactl ssh f34fea36
Connecting to ubuntu@161.153.31.243...
```

Accepts instance ID or name:

```
$ lambdactl ssh my-instance-name
```

### List instances

```
$ lambdactl instances
ID                                NAME            STATUS  IP              TYPE              REGION
f34fea366052433a8e37e9a5897b1b3e  lambdactl-test  active  161.153.31.243  gpu_1x_a100_sxm4  us-west-2
```

### Stop an instance

```
$ lambdactl stop lambdactl-test
Terminate instance "lambdactl-test" (f34fea36...)? [y/N] y
Instance "lambdactl-test" (f34fea36...) terminated.
```

Skip confirmation with `--yes`:

```
$ lambdactl stop f34fea36 --yes
```

### Restart an instance

```
$ lambdactl restart my-instance-name
```

### Rename an instance

```
$ lambdactl rename f34fea36 new-name
```

### Manage SSH keys

```
$ lambdactl ssh-keys
ID                                NAME
168fe8923af3489082ae4e86c39025bb  my-key

$ lambdactl ssh-keys add work-key ~/.ssh/id_ed25519.pub
Added SSH key "work-key" (9db4a5a8...)

$ lambdactl ssh-keys rm 9db4a5a832ef4dffbb38ab22ccde97a2
SSH key 9db4a5a8... deleted.
```

### List OS images

```
$ lambdactl images
FAMILY                     NAME                         VERSION          REGIONS
gpu-base-24-04             GPU Base 24.04               24.4.4-2141      16
lambda-stack-24-04         Lambda Stack 24.04           24.4.4-2141      16
ubuntu-24-04               Ubuntu 24.04.2 LTS           24.4.2-20250626  15

$ lambdactl images --family ubuntu-24-04 --region us-west-2
```

### List firewall rules

```
$ lambdactl firewall
PROTOCOL  PORTS  SOURCE     DESCRIPTION
tcp       22     0.0.0.0/0  Allow SSH connections from any IP
icmp      all    0.0.0.0/0  Allow Ping from any IP
```

## JSON output

All commands support `--json`:

```
$ lambdactl instances --json
$ lambdactl types --json | jq '.[] | select(.regions_with_capacity_available | length > 0)'
```

## Shell completions

```
lambdactl completion bash > /etc/bash_completion.d/lambdactl
lambdactl completion zsh > "${fpath[1]}/_lambdactl"
lambdactl completion fish > ~/.config/fish/completions/lambdactl.fish
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
