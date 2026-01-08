// package contract defines the stable, stdlib-only API surface shared by the Onyx Database client.
//
// The contract package owns the data structures and interfaces that other packages implement.
// It purposefully avoids any runtime behavior so the API remains deterministic and easy to audit.
// Implementations and side effects live in sibling packages such as onyx or internal; this
// package stays dependency-free beyond the Go standard library to keep the surface small and
// portable across environments.
//
// Stability follows semantic versioning with an additive-only policy for v1.x:
//
//   - new symbols may be added in minor versions as long as existing behavior is preserved
//   - existing public types, functions, and method signatures must not change in incompatible ways
//   - breaking changes require a major version increment and an updated STABILITY.md
package contract
