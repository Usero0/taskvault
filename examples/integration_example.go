package examples

import (
	"fmt"
	"os"
)

// ExampleIntegration shows basic integration patterns
func ExampleIntegration() {
	fmt.Println("TaskVault Integration Example")
	fmt.Println("============================")
	fmt.Println("This example shows how to integrate TaskVault into a workflow")
	fmt.Println()
	fmt.Println("Basic workflow:")
	fmt.Println("1. taskvault init          - Create .taskvault/config.yaml")
	fmt.Println("2. taskvault cache save    - Store task output")
	fmt.Println("3. taskvault cache get     - Retrieve cached result")
	fmt.Println("4. taskvault cache stats   - Monitor cache health")
	fmt.Println()
	fmt.Println("For SDK usage, see examples/sdk_examples.go")
	fmt.Println()

	// This function can serve as a template for custom integrations
	args := os.Args
	if len(args) > 1 {
		fmt.Printf("Arguments: %v\n", args[1:])
	}
}
