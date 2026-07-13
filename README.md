# Check Scripts

## Templates

The `templates/` directory contains reusable Rice script templates organized per folder.

### Structure

```
templates/
├── bundle.json                 # Auto-generated bundle of all templates
├── schema.json                 # Manifest JSON schema
└── {template-name}/
    ├── manifest.json           # Template metadata + variant list
    └── scripts/
        └── {variant}.rice      # Rice script per variant
```

### Manifest schema

Each template folder contains a `manifest.json` defining the template name, description, and variants. Each variant references a `scriptFile` — the relative path to its Rice script.

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
