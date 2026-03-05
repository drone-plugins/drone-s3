# Leading-Slash Regression Audit Report: drone-s3 Plugin

**Repo:** `/Users/dhiraj/Code/drone-plugins/drone-s3`  
**Branch:** CI-21342 (commit 92ec948 on top of master 7cb43d6)  
**Context:** AWS SDK v1 → v2 migration; leading-slash bug fix

---

## 1. Executive Summary

The v1 code added leading slashes to S3 keys via `if !strings.HasPrefix(key, "/") { key = "/" + key }`. AWS SDK v1 silently stripped these; SDK v2 preserves them literally, causing 404s when other tools access objects without the leading slash. The fix removes the leading-slash-adding logic and instead strips leading slashes.

**All S3 key construction paths now produce keys without leading slashes.** The download path does not construct keys—it uses keys from ListObjects responses—so the fix does not change download behavior.

---

## 2. Code Paths That Construct S3 Keys

### 2.1 Upload Path (Exec, PutObject)

| Location | Condition | Key Construction |
|----------|-----------|------------------|
| `plugin.go:211-215` | `normalizedStrip != "" && !HasPrefix(normalizedStrip, "/")` (relative strip_prefix) | `target = resolveKey(p.Target, filepath.ToSlash(match), p.StripPrefix)` |
| `plugin.go:214-215` | else (absolute strip_prefix or no strip_prefix) | `rel := strings.TrimPrefix(filepath.ToSlash(stripped), "/"); target = filepath.ToSlash(filepath.Join(p.Target, rel))` |

### 2.2 resolveKey() Function

**v1 (7cb43d6):**
```go
func resolveKey(target, srcPath, stripPrefix string) string {
	key := filepath.Join(target, strings.TrimPrefix(srcPath, filepath.ToSlash(stripPrefix)))
	key = filepath.ToSlash(key)
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	return key
}
```

**v2 Fixed (current):**
```go
func resolveKey(target, srcPath, stripPrefix string) string {
	key := filepath.Join(target, strings.TrimPrefix(srcPath, filepath.ToSlash(stripPrefix)))
	key = filepath.ToSlash(key)
	key = strings.TrimPrefix(key, "/")
	return key
}
```

### 2.3 Exec Upload Branch (Absolute Strip Prefix)

**v1 (7cb43d6):**
```go
rel := strings.TrimPrefix(filepath.ToSlash(stripped), "/")
target = filepath.ToSlash(filepath.Join(p.Target, rel))
if !strings.HasPrefix(target, "/") {
	target = "/" + target
}
```

**v2 Fixed (current):**
```go
rel := strings.TrimPrefix(filepath.ToSlash(stripped), "/")
target = filepath.ToSlash(filepath.Join(p.Target, rel))
```

---

## 3. Comparison Table: Key Values by Code Path

| Code Path | Input Example | v1.5.8 Effective Key (SDK v1 normalized) | v1.5.9 Original Key (before fix, SDK v2) | v1.5.9 Fixed Key (current) |
|-----------|---------------|-------------------------------------------|------------------------------------------|----------------------------|
| **resolveKey** (relative strip) | target=`hello`, src=`/foo/bar`, strip=`/foo` | `hello/bar` | `/hello/bar` | `hello/bar` |
| **resolveKey** (target empty) | target=``, src=`/foo/bar`, strip=`/foo` | `bar` | `/bar` | `bar` |
| **resolveKey** (no strip) | target=`hello`, src=`/foo/bar`, strip=`` | `hello/foo/bar` | `/hello/foo/bar` | `hello/foo/bar` |
| **Exec absolute strip** | target=`deployment`, stripped=`module1/app.zip` | `deployment/module1/app.zip` | `/deployment/module1/app.zip` | `deployment/module1/app.zip` |
| **Exec absolute strip** | target=`releases`, stripped=`auth/v1.0/auth-service.zip` | `releases/auth/v1.0/auth-service.zip` | `/releases/auth/v1.0/auth-service.zip` | `releases/auth/v1.0/auth-service.zip` |

**Note:** v1.5.8 "Effective Key" = what SDK v1 actually stored (it stripped leading slashes). v1.5.9 "Original Key" = what the broken v2 migration would have sent (causing 404s). v1.5.9 "Fixed Key" = correct behavior matching v1.5.8 effective keys.

---

## 4. Download Path Analysis

### 4.1 Does download construct keys or use ListObjects responses?

**Answer: Uses keys from ListObjects responses. Does NOT construct keys.**

Flow:
1. `downloadS3Objects(ctx, client, sourceDir)` — `sourceDir = normalizePath(p.Source)` (no leading slash)
2. `ListObjectsV2(Bucket, Prefix: &sourceDir)` — lists objects under prefix
3. For each `item` in `list.Contents`: `key = *item.Key` (from S3)
4. `downloadS3Object(ctx, client, sourceDir, *item.Key, target)` — passes `*item.Key` to GetObject
5. `GetObject(Bucket, Key: &key)` — uses S3 key as-is

S3 object keys in ListObjects/GetObject are always stored without leading slashes. The plugin never constructs keys for download; it uses whatever S3 returns.

### 4.2 Does the fix change download behavior?

**No.** The fix only affects:
- `resolveKey()` — used only in upload path when `normalizedStrip != "" && !HasPrefix(normalizedStrip, "/")`
- Exec upload branch — when using absolute strip_prefix or no strip_prefix

The download path uses:
- `sourceDir` (ListObjects Prefix) — from `normalizePath(p.Source)` which already strips leading slashes
- `*item.Key` (GetObject Key) — from S3 ListObjects response, not constructed

---

## 5. ListObjects Prefix (sourceDir)

For download mode:
- `p.Source = normalizePath(p.Source)` at Exec start
- `sourceDir := normalizePath(p.Source)` (redundant but safe)
- `normalizePath(path) = strings.TrimPrefix(filepath.ToSlash(path), "/")` — always produces path without leading slash

So the ListObjects `Prefix` is never given a leading slash. Correct.

---

## 6. Test Expectations vs Fixed Behavior

### 6.1 Tests That Reference S3 Keys

| Test File | Test | Key-Related Expectations | Matches Fixed? |
|-----------|------|--------------------------|----------------|
| `plugin_unix_test.go` | TestResolveUnixKey | `bar`, `hello/foo/bar`, `hello/bar` | ✅ All no leading slash |
| `plugin_windows_test.go` | TestResolveWinKey | `bar`, `hello/foo/bar`, `hello/bar`, `hello/world` | ✅ All no leading slash |
| `wildcard_strip_test.go` | TestBuildKeyWithWildcards | `deployment/module1/app.zip`, `releases/auth/v1.0/auth-service.zip`, `upload/app.zip`, `backup/build123/lib.zip`, `fallback/different/location/file.zip` | ✅ All no leading slash |
| `wildcard_strip_test.go` | TestResolveKey_BackCompat | `hello/bar.zip`, `hello/foo/bar.zip`, `hello/bar.zip`, `hello/foo/bar.zip` | ✅ All no leading slash |

### 6.2 Tests That Reference Paths (Not S3 Keys)

| Test File | Test | Expectations | Notes |
|-----------|------|--------------|-------|
| `plugin_unix_test.go` | TestNormalizePath | `path/to/file.txt`, etc. | Strips leading slash |
| `plugin_unix_test.go` | TestResolveSource | `output-file.txt`, `images/logo.png`, etc. | Local file paths, not S3 keys |
| `wildcard_strip_test.go` | TestStripWildcardPrefix | Various stripped paths | Path manipulation, not S3 keys |

### 6.3 Summary

All test expectations that reference S3 keys expect keys **without** leading slashes. The fixed implementation produces these values. All tests pass.

---

## 7. Evidence Summary

| Item | Evidence |
|------|----------|
| v1 code added leading slash | `if !strings.HasPrefix(key, "/") { key = "/" + key }` in resolveKey; `if !strings.HasPrefix(target, "/") { target = "/" + target }` in Exec |
| v2 fix strips leading slash | `key = strings.TrimPrefix(key, "/")` in resolveKey; removed leading-slash addition in Exec |
| Download uses ListObjects keys | `downloadS3Object(..., *item.Key, target)` — key from `item.Key` |
| Download unaffected by fix | No code path in download constructs keys; all keys come from S3 |
| Tests match fixed behavior | All resolveKey/BuildKey expectations expect keys without leading slash |

---

## 8. Conclusion

The leading-slash fix is correct and complete. All code paths that construct S3 keys now produce keys without leading slashes, matching the effective behavior of SDK v1. The download path is unaffected because it uses keys directly from S3 ListObjects responses. All tests pass and their expectations align with the fixed behavior.
