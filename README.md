> inspired from [github/qdm12](https://github.com/qdm12/golibs)

## Go Project Guide

### Build

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -tags=nomsgpack .
```