// Package policy defines the public contract for ArchGuard rules: the Rule interface,
// the ScanContext a rule receives, and the Issue/ScanResult types it produces.
//
// Third-party rules implement policy.Rule and can be registered with the engine.
package policy
