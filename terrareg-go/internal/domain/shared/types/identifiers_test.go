package types

import "testing"

func TestTypeSafety(t *testing.T) {
	// Demonstrate that the types work as expected
	ns := NamespaceName("test-namespace")
	mn := ModuleName("test-module")
	pn := ProviderName("aws")

	if string(ns) != "test-namespace" {
		t.Errorf("expected 'test-namespace', got '%s'", ns)
	}
	if string(mn) != "test-module" {
		t.Errorf("expected 'test-module', got '%s'", mn)
	}
	if string(pn) != "aws" {
		t.Errorf("expected 'aws', got '%s'", pn)
	}

	// Demonstrate that these types prevent accidental swapping
	// This would be a compile error if we tried:
	// func Process(namespace string, module NamespaceName) {}
	// Process(mn, ns) // Compile error! Can't pass ModuleName where string expected
}

// Demonstrate that we can't accidentally mix up types at compile time
func ProcessModule(ns NamespaceName, mn ModuleName, pn ProviderName) string {
	return string(ns) + "/" + string(mn) + "/" + string(pn)
}

func TestTypePrevention(t *testing.T) {
	// This works correctly
	result := ProcessModule(
		NamespaceName("my-ns"),
		ModuleName("my-mod"),
		ProviderName("aws"),
	)
	expected := "my-ns/my-mod/aws"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}

	// If we tried to swap arguments, it would be a compile error:
	// ProcessModule(
	//     ModuleName("my-mod"),  // Wrong! NamespaceName expected
	//     NamespaceName("my-ns"), // Wrong! ModuleName expected
	//     ProviderName("aws"),
	// )
}
