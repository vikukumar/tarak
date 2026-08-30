# Security Policy

## Supported Versions

The following versions of Tarak receive active security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.6   | ✅ Current release  |
| 1.0.x   | ✅ Patch updates    |
| < 1.0.0 | ❌ Not supported   |

---

## Reporting a Vulnerability

We take the security of Tarak very seriously. If you discover a security vulnerability, please report it **privately** rather than opening a public GitHub issue.

### How to Report

Email the maintainers directly with the subject line: **[SECURITY] Tarak Vulnerability Report**

You should receive a response within **48 hours**. If the issue is confirmed, a patch will be released as soon as possible (usually within a few days depending on complexity).

### What to Include

- A clear description of the vulnerability and its potential impact
- Detailed steps to reproduce the issue
- The Tarak version affected (`tarakctl version`)
- Platform and OS details
- Any proof-of-concept code or output, if available

---

## Security Features in Tarak

Tarak includes several built-in security mechanisms:

| Feature | Description |
|---|---|
| **TarakSecurityPolicy** | Native network policy engine — define `Allow`/`Deny` rules per namespace and port |
| **Admission Validation** | All resource create/update requests are validated before persisting |
| **Immutable Field Protection** | `metadata.name`, `metadata.namespace`, and `kind` cannot be changed after creation |
| **mTLS Support** | Client certificate authentication between CLI and API server |
| **Exec Credential Plugins** | Supports EKS/GKE/AKS exec-based auth providers (no credential leakage in kubeconfig) |
| **Token File Auth** | Supports projected service account token files (short-lived, rotatable) |
| **RBAC-ready API** | API server enforces `Authorization` header on all requests |

---

## Thank You

Thank you for helping keep Tarak and its users safe!
