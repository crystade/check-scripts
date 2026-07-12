# Check Scripts

## Test Samples

The `test-samples/` directory contains mock check results used to validate Rice alerting scripts.

### Structure

```
test-samples/
├── index.json          # Auto-generated index of all samples
├── src/{protocol}/     # Sample files organized by protocol (http, https, tcp, tls, minecraft)
│   └── {sample}.json   # Metadata (check result payload)
│   └── {sample}.txt    # Payload (request debug dump)
└── test/
    ├── test.rice       # Alerting/incident test script
    ├── main.go         # Minimal entry point
    └── main_test.go    # Test runner — executes test.rice against all indexed samples
```

### Indexing samples

Scan all sample files and regenerate `index.json`:

```
go run ./tools/index_test_samples.go
```

### Running tests

Execute the Rice test script against every sample:

```
go test ./test-samples/test/... -v
```
