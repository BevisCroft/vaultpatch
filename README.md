# vaultpatch

> CLI tool to diff and apply HashiCorp Vault secret changes across environments safely

---

## Installation

```bash
go install github.com/yourusername/vaultpatch@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourusername/vaultpatch/releases).

---

## Usage

**Diff secrets between two environments:**

```bash
vaultpatch diff --src secret/staging/app --dst secret/production/app
```

**Apply changes from a patch file:**

```bash
vaultpatch apply --patch changes.patch --env production
```

**Generate a patch and review before applying:**

```bash
vaultpatch diff --src secret/staging/app --dst secret/production/app --output changes.patch
vaultpatch apply --patch changes.patch --dry-run
```

### Common Flags

| Flag | Description |
|------|-------------|
| `--addr` | Vault server address (default: `$VAULT_ADDR`) |
| `--token` | Vault token (default: `$VAULT_TOKEN`) |
| `--dry-run` | Preview changes without applying them |
| `--output` | Write diff output to a file |

---

## Requirements

- Go 1.21+
- HashiCorp Vault 1.12+
- `VAULT_ADDR` and `VAULT_TOKEN` environment variables set

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

---

## License

[MIT](LICENSE)