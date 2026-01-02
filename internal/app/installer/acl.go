//go:build windows
// +build windows

package installer

// Placeholder for ACL helpers. In a fuller implementation this would adjust
// filesystem permissions for service and tray artifacts.
func EnsurePaths() error {
	return nil
}
