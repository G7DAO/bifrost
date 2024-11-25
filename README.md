<p align="center">
<a href="https://game7.io/">
<img width="4335" alt="GAME7_LOGO_RED" src="https://github.com/G7DAO/chainprof/assets/38267570/600f533f-782d-49ef-97f5-11c096c2e13b">
</a>
</p>

<h1 align="center">G7 Bifrost</h1>

The G7 Bifrost is a cross-chain messaging CLI.

## What this repository contains

The [`G7DAO/bifrost`](https://github.com/G7DAO/bifrost) repository contains the CLI for the G7 Bifrost.

It also contains:

1. Go bindings to these contracts
2. The `bifrost` command line tool which can be used to deploy and interact with these contracts

## Development

### Requirements

- [Node.js](https://nodejs.org/en) (version >= 20)
- [Go](https://go.dev/) (version >= 1.21), for the `bifrost` CLI, and other developmental and operational tools
- [`seer`](https://github.com/G7DAO/seer), which we use to generate Go bindings and command-line interfaces

### Building this code

The [`Makefile`](./Makefile) for this project can be used to build all the code in the repository.

Build everything using:

```bash
make
```
