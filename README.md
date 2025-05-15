# Terraform Policy Plugin Framework

terraform-policy-plugin-framework is a module for building Terraform Policy plugins.
It is built on [go-plugin](https://github.com/hashicorp/go-plugin).

Terraform Policy is a new experimental approach to managing policy within HCP Terraform.
This repository enables Terraform Policy users to write and configure their own functions for consumption within the Terraform Policy runtime.

## Status

Terraform Policy is still considered to be in development and is not generally available. 
This repository should only be used after communication with Hashicorp Product Managers and Engineers.

### Experimental Disclaimer

By using the software in this repository (the "Software"), you acknowledge that: (1) the Software is still in development, may change, and has not been released as a commercial product by HashiCorp and is not currently supported in any way by HashiCorp; (2) the Software is provided on an "as-is" basis, and may include bugs, errors, or other issues; (3) the Software is NOT INTENDED FOR PRODUCTION USE, use of the Software may result in unexpected results, loss of data, or other unexpected results, and HashiCorp disclaims any and all liability resulting from use of the Software; and (4) HashiCorp reserves all rights to make all decisions about the features, functionality and commercial release (or non-release) of the Software, at any time and without any obligation or liability whatsoever.

### Contributing

Given the experimental nature of this project, we are not currently accepting external contributions to this repository.

## Getting Started

The `policy-plugin/plugins` package is the main entry point for plugin development.

New functions can be registered using the `plugins.RegisterFunction` and `plugins.RegisterFunctionDirect` functions.

The `RegisterFunction` function accepts a name, and a Go `func` implementation. 
This implementation can accept any number of function arguments, but must have exactly two return types with the second return type being a Go `error`.
The plugin framework will attempt to convert between the required Go types and the underlying Terraform type system automatically and will panic if this is not possible.

More experienced users can use the `RegisterFunctionDirect` function, which accepts a [go-cty](https://github.com/zclconf/go-cty) `function.Function` directly.
This allows direct control over the concrete Terraform types that will be accepted and returned by the function when used within Terraform Policy.

Once all functions have been registered, `plugins.Serve` should be called. 
The `Serve` function will block and wait for external connections from Terraform Policy. 
This function must be the last operation called by the `main` function, and uses `go-plugin` to start an RPC server that can interface with Terraform Policy.  

You can use the `plugins.CallFunction` function from Go test files to test your functions after they have been registered.
This is important for the `RegisterFunction` function in particular, as it will make sure the automatic conversion process has succeeded.

### Example plugin

```go
// main.go

package main

import "github.com/hashicorp/terraform-policy-plugin-framework/policy-plugin/plugins"

func main() {
	plugins.RegisterFunction("echo", func(input string) (string, error) {
		// simple echo function, just return the string input directly.
		return input, nil
    })
	plugins.Serve()
}
```

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/terraform-policy-plugin-framework/blob/main/LICENSE)
