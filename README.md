> inspired from [github/qdm12](https://github.com/qdm12/golibs)

## Go Project Guide

### Build

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -tags=nomsgpack .
```

### Recommend packages

#### Leveled logging

- [Zerolog](https://github.com/rs/zerolog)
- [zap](https://github.com/uber-go/zap)
- log/slog(go 1.21.0+)
    > Startup: https://stackoverflow.com/questions/16895651/how-to-implement-level-based-logging-in-golang/76867161#76867161