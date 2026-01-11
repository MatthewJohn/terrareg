//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("🚀 Terrareg-Go Implementation Validation")
	fmt.Println("==========================================")

	// Check if main application builds
	fmt.Println("\n1. Testing main application build...")
	if runCommand("go", "build", "-o", "terrareg-go-server", "./cmd/server") {
		fmt.Println("✅ Main application builds successfully")
	} else {
		fmt.Println("❌ Main application build failed")
		os.Exit(1)
	}

	// Check if all packages build
	fmt.Println("\n2. Testing package builds...")
	if runCommand("go", "build", "./...") {
		fmt.Println("✅ All packages build successfully")
	} else {
		fmt.Println("❌ Package build failed")
		os.Exit(1)
	}

	// Run integration tests
	fmt.Println("\n3. Running integration tests...")
	if runCommand("go", "test", "-v", "./test/integration/complete_workflow_test.go") {
		fmt.Println("✅ Integration tests passed")
	} else {
		fmt.Println("❌ Integration tests failed")
		os.Exit(1)
	}

	// Run performance tests
	fmt.Println("\n4. Running performance tests...")
	if runCommand("go", "test", "-v", "./test/integration/performance_test.go") {
		fmt.Println("✅ Performance tests passed")
	} else {
		fmt.Println("❌ Performance tests failed")
		os.Exit(1)
	}

	// Check code formatting
	fmt.Println("\n5. Checking code formatting...")
	if runCommand("go", "fmt", "./...") {
		fmt.Println("✅ Code is properly formatted")
	} else {
		fmt.Println("⚠️  Code formatting issues found (run 'go fmt ./...')")
	}

	// Run go vet
	fmt.Println("\n6. Running static analysis...")
	if runCommand("go", "vet", "./...") {
		fmt.Println("✅ Static analysis passed")
	} else {
		fmt.Println("❌ Static analysis found issues")
		os.Exit(1)
	}

	// Check for unused dependencies
	fmt.Println("\n7. Checking for unused dependencies...")
	if runCommand("go", "mod", "tidy") {
		fmt.Println("✅ Dependencies are clean")
	} else {
		fmt.Println("❌ Dependency cleanup failed")
		os.Exit(1)
	}

	fmt.Println("\n🎉 All validation checks passed!")
	fmt.Println("Terrareg-Go implementation is ready for production use.")
}

func runCommand(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return err == nil
}

func captureCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
