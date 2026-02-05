# Fix S3 XAmzContentSHA256Mismatch for S3-Compatible Stores

**Goal:** Add a configuration option to disable S3 payload signing (use `UNSIGNED-PAYLOAD`), fixing uploads to S3-compatible stores like Cellar/Ceph that fail with `XAmzContentSHA256Mismatch`.

**Architecture:** The AWS SDK Go v2 computes a SHA256 hash of the request body and sends it in the `X-Amz-Content-Sha256` header for SigV4 signing. Some S3-compatible stores (Ceph RadosGW, Cellar, etc.) don't verify this correctly, causing `XAmzContentSHA256Mismatch` errors. The SDK provides `v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware` to replace the hash with `UNSIGNED-PAYLOAD`, which is safe over HTTPS. We add a `files.s3.disablesigning` config option that applies this middleware when creating the S3 client.

**Tech Stack:** Go, AWS SDK Go v2 (`aws/signer/v4`), viper config

---

### Task 1: Add the config key

**Files:**
- Modify: `pkg/config/config.go:170` (add key constant)
- Modify: `pkg/config/config.go:448` (add default)

**Step 1: Add the config key constant**

Add after `FilesS3UsePathStyle` (line 170) in `pkg/config/config.go`:

```go
FilesS3UsePathStyle Key = `files.s3.usepathstyle`
FilesS3DisableSigning Key = `files.s3.disablesigning`
FilesS3TempDir      Key = `files.s3.tempdir`
```

**Step 2: Add the default value**

Add after `FilesS3UsePathStyle.setDefault(false)` (line 448) in `pkg/config/config.go`:

```go
FilesS3UsePathStyle.setDefault(false)
FilesS3DisableSigning.setDefault(false)
FilesS3TempDir.setDefault("")
```

**Step 3: Commit**

```bash
git add pkg/config/config.go
git commit -m "feat: add files.s3.disablesigning config key"
```

---

### Task 2: Add config to `config-raw.json`

This file generates `config.yml.sample` and documents the option for users.

**Files:**
- Modify: `config-raw.json` (add entry after `usepathstyle` in the `s3` children array, around line 531)

**Step 1: Add the config entry**

Insert after the `usepathstyle` entry (line 527-531) in `config-raw.json`:

```json
{
    "key": "disablesigning",
    "default_value": "false",
    "comment": "When enabled, the S3 client will send UNSIGNED-PAYLOAD instead of computing a SHA256 hash for request signing. Some S3-compatible providers (such as Ceph RadosGW, Clever Cloud Cellar) do not correctly verify payload signatures and return XAmzContentSHA256Mismatch errors. Enabling this option works around the issue. Only applies over HTTPS."
}
```

**Step 2: Commit**

```bash
git add config-raw.json
git commit -m "docs: add files.s3.disablesigning to config template"
```

---

### Task 3: Write the failing test

**Files:**
- Modify: `pkg/files/s3_test.go` (add test at the end)

**Step 1: Write the test**

Add to the end of `pkg/files/s3_test.go`:

```go
func TestInitS3FileHandler_DisableSigningOption(t *testing.T) {
	// Save original config
	originalType := config.FilesType.GetString()
	originalEndpoint := config.FilesS3Endpoint.GetString()
	originalBucket := config.FilesS3Bucket.GetString()
	originalAccessKey := config.FilesS3AccessKey.GetString()
	originalSecretKey := config.FilesS3SecretKey.GetString()
	originalDisableSigning := config.FilesS3DisableSigning.GetBool()
	originalClient := s3Client

	defer func() {
		config.FilesType.Set(originalType)
		config.FilesS3Endpoint.Set(originalEndpoint)
		config.FilesS3Bucket.Set(originalBucket)
		config.FilesS3AccessKey.Set(originalAccessKey)
		config.FilesS3SecretKey.Set(originalSecretKey)
		config.FilesS3DisableSigning.Set(originalDisableSigning)
		s3Client = originalClient
	}()

	config.FilesS3Endpoint.Set("https://fake-endpoint.example.com")
	config.FilesS3Bucket.Set("test-bucket")
	config.FilesS3AccessKey.Set("test-key")
	config.FilesS3SecretKey.Set("test-secret")
	config.FilesS3DisableSigning.Set(true)

	// initS3FileHandler should succeed (validation will fail because the
	// endpoint is fake, but that happens later in InitFileHandler).
	// We call initS3FileHandler directly to test the client is created.
	err := initS3FileHandler()
	// The error will be from ValidateFileStorage or nil -- either way,
	// initS3FileHandler itself should not error for the signing option.
	// We can't easily verify the middleware was applied without a real
	// request, so we verify no panic or init error.
	assert.NoError(t, err)
	assert.NotNil(t, s3Client)
}
```

**Step 2: Run the test to verify it fails**

Run: `mage test:filter TestInitS3FileHandler_DisableSigningOption`
Expected: FAIL -- `config.FilesS3DisableSigning` does not exist yet if Task 1 isn't done, or passes if Task 1 is already done (in which case the test serves as a regression guard).

**Step 3: Commit**

```bash
git add pkg/files/s3_test.go
git commit -m "test: add test for S3 disable signing config option"
```

---

### Task 4: Apply the middleware in `initS3FileHandler`

**Files:**
- Modify: `pkg/files/filehandling.go:96-100` (add middleware option)
- Modify: `pkg/files/filehandling.go` imports (add `v4` signer import)

**Step 1: Add the v4 signer import**

Add to the imports in `pkg/files/filehandling.go`:

```go
v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
```

**Step 2: Apply the middleware when config is enabled**

Replace the S3 client creation block (lines 97-100) in `pkg/files/filehandling.go`:

```go
// Before:
client := s3.NewFromConfig(cfg, func(o *s3.Options) {
    o.BaseEndpoint = aws.String(endpoint)
    o.UsePathStyle = config.FilesS3UsePathStyle.GetBool()
})

// After:
client := s3.NewFromConfig(cfg, func(o *s3.Options) {
    o.BaseEndpoint = aws.String(endpoint)
    o.UsePathStyle = config.FilesS3UsePathStyle.GetBool()
    if config.FilesS3DisableSigning.GetBool() {
        o.APIOptions = append(o.APIOptions, v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
    }
})
```

This replaces the `ComputePayloadSHA256` middleware with `UnsignedPayload` for all requests made by this client. Instead of computing the body's SHA256 and setting it in `X-Amz-Content-Sha256`, the client sends `UNSIGNED-PAYLOAD`. This is safe over HTTPS since TLS provides integrity guarantees.

**Step 3: Run the test**

Run: `mage test:filter TestInitS3FileHandler_DisableSigningOption`
Expected: PASS

**Step 4: Run the full file test suite**

Run: `mage test:filter TestFileSave_S3`
Expected: PASS (existing tests unaffected since the config defaults to false)

**Step 5: Run lint**

Run: `mage lint`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/files/filehandling.go
git commit -m "feat: support UNSIGNED-PAYLOAD for S3-compatible stores

Add files.s3.disablesigning config option that replaces the SHA256 payload
hash with UNSIGNED-PAYLOAD in S3 requests. This fixes XAmzContentSHA256Mismatch
errors with S3-compatible providers like Ceph RadosGW and Clever Cloud Cellar."
```

---

### Task 5: Run the config init tests

**Files:** (no changes, verification only)

**Step 1: Run all S3 config tests**

Run: `mage test:filter TestInitFileHandler_S3`
Expected: PASS -- existing config validation tests should still pass since `disablesigning` is only read, not validated as required.

**Step 2: Run full lint**

Run: `mage lint`
Expected: PASS

---

### Summary of changes

| File | Change |
|------|--------|
| `pkg/config/config.go` | Add `FilesS3DisableSigning` key constant and `false` default |
| `config-raw.json` | Add `disablesigning` entry with documentation comment |
| `pkg/files/filehandling.go` | Import `v4` signer, apply `SwapComputePayloadSHA256ForUnsignedPayloadMiddleware` when config is true |
| `pkg/files/s3_test.go` | Add `TestInitS3FileHandler_DisableSigningOption` test |

### User-facing documentation

Users who encounter `XAmzContentSHA256Mismatch` errors should add to their `config.yml`:

```yaml
files:
  type: s3
  s3:
    endpoint: "https://..."
    bucket: "..."
    accesskey: "..."
    secretkey: "..."
    disablesigning: true
```
