// Testing utils package
package utils

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/pkg/errors"
)

// AssertStringNotEmpty asserts when the string is not empty
func AssertStringNotEmpty(message, str string) diag.Diagnostics {
	var diags diag.Diagnostics
	str = strings.TrimSpace(str)
	if str != "" {
		return diags
	}

	if message != "" {
		return diag.FromErr(errors.New(fmt.Sprintf("%s: expected not empty string.", message)))
	} else {
		return diag.FromErr(errors.New("Expected not empty string."))
	}
}
