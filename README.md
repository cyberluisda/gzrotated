# gzrotated
Compress files that was rotated for other app with gzip algorithm based on modification time

# Requirements

This project requires golang 1.16 version or higher

# Build

Use go command to build executable file. For example:

```bash
go build -o build/gzrotated main.go
```

# Run

Call with `-h` option to see all available options.

For example:
```bash
go run main.go -h
```

or, assuming you previously built with `-o gzrotated` option:

```bash
./build/gzrotated -h
```
