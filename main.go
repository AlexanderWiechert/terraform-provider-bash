// terraform-provider-bash is a small utility Terraform provider which aims
// to make it easier to integrate values from Terraform into a bash script,
// automatically handling the necessarily escaping to translate values
// faithfully from Terraform's type system into Bash's type system.
//
// This provider can use features that require Bash 4, but if you need to work
// with an earlier version of Bash then you can avoid the Bash 4 requirement
// by not passing in maps, and thus avoiding this provider trying to generate
// associative arrays.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/apparentlymart/terraform-provider-bash/internal/bash"
)

func main() {
	provider := bash.NewProvider()
	err := provider.Serve(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start provider: %s", err)
		os.Exit(1)
	}
}
