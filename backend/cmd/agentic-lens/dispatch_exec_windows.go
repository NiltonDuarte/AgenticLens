package main

// Windows has no process-image replacement. The outer PowerShell entrypoint
// owns its inherited-handle transport and bounded advisory process tree.
func execManualNative(path string, args []string) (bool, error) { return false, nil }
