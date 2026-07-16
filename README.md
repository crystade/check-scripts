# Check Scripts

## Templates

The `templates/` directory contains reusable Rice script templates organized per folder.

### Structure

```
templates/
├── bundle.json                 # Auto-generated bundle of all templates
├── schema.json                 # Manifest JSON schema
└── {template-name}/
    ├── manifest.yml            # Template metadata + variant list
    └── scripts/
        └── {variant}.rice      # Rice script per variant
```

### Manifest schema

Each template folder contains a `manifest.yml` (YAML) defining the template name, description, and variants. Each variant references a `scriptFile` — the relative path to its Rice script.

A JSON Schema for the manifest is available at `templates/schema.json`.

### Bundling templates

Bundle all templates into a single self-contained `bundle.json`, embedding each script's content directly:

```
go run ./tools/bundle_templates.go
```

The output replaces `scriptFile` with `script` (inline content) for each variant.

## Test Samples

The `test-samples/` directory contains mock check results used to validate Rice alerting scripts.

### Structure

```
test-samples/
├── bundle.json          # Auto-generated bundle of all samples (inlined metadata + payload)
├── index.json           # Auto-generated index of all samples
└── src/{protocol}/      # Sample files organized by protocol (http, https, tcp, tls, minecraft)
    ├── {sample}.json    # Metadata (check result payload)
    └── {sample}.txt     # Payload (request debug dump)
```

### Indexing samples

Scan all sample files and regenerate `index.json`:

```
go run ./tools/index_test_samples.go
```

### Bundling samples

Bundle all samples into a single self-contained `bundle.json`, embedding each sample's metadata and payload inline as Base64:

```
go run ./tools/bundle_test_samples.go
```

## Template Testing

The `template-testing/` directory contains the test runner that validates every template variant against all test samples.

### Structure

```
template-testing/
├── eat.rice          # Test script template — uses `#{{__script__}}` placeholder injected at runtime
├── main.go           # Minimal entry point
└── main_test.go      # Test runner — executes eat.rice + each variant script against all indexed samples
```

### How it works

`main_test.go` loads `templates/bundle.json` and `test-samples/index.json`, then for each template variant injects its script into the `eat.rice` template (replacing the `#{{__script__}}` placeholder) and executes the combined script against every test sample.

### Running tests

Execute the Rice test script against every sample:

```
go test ./template-testing/... -v
```

## Dependencies

Managed via Go modules (`go.mod`). Requires Go 1.24+.

```
go mod tidy
```
