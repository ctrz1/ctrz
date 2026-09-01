
<img src="./docs/examples/assets/logo.png" width="200" style="display: block; margin: auto;">


# ctrz

`ctrz` is an image-less ad-hoc containerisation runtime. It lets you run statically linked binaries in an isolated environment. Find some [documentation](/docs/examples/README.md) on how to use `ctrz`.

## Installation

> Note that `ctrz` is only available for Linux

To install `ctrz`, you have different options:

1) Run the following command that will install `ctrz` on your PATH for you via the [install script](/install):

```bash
curl -fsSL https://raw.githubusercontent.com/ctrz1/ctrz/main/install | sudo bash
```

2) Manually download a binary from the [releases](https://github.com/ctrz1/ctrz/releases)
3) Clone the repo and build the project yourself using `make build`.

## Requirements

1. To fully run and use `ctrz`, root privileges are required. Rootless containers are not supported. It is recommended to use `sudo ctrz ...` for all commands.
2. `ctrz` is currently still using `iptables` to define networking rules. An effort to migrate to `nftables` is being made [here](https://github.com/ctrz1/ctrz/tree/nftables).
