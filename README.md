# cooltower

Industrial cooling tower controller coordinating fan VFD banks, spray header flow, drift eliminator windows, and fan/spray interlocks.

## Build

```bash
make build
make test
```

## Run

```bash
go run ./cmd/cooltower
```

## HMI

With `-web :8080` the embedded dashboard exposes tower state, drift readings, and alarm history.

## Benzhi

See [BENZHI_README.md](BENZHI_README.md).
