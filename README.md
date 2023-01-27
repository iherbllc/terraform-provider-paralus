# Paralus terraform provider

[![Action](https://github.com/iherbllc/terraform-provider-paralus/workflows/build/badge.svg)](https://github.com/iherbllc/terraform-provider-paralus/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/iherbllc/terraform-provider-paralus)](https://goreportcard.com/badge/github.com/iherbllc/terraform-provider-paralus)
[![Go Version](https://img.shields.io/github/go-mod/go-version/iherbllc/terraform-provider-paralus)](https://img.shields.io/github/go-mod/go-version/iherbllc/terraform-provider-paralus)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/iherbllc/terraform-provider-paralus?tab=overview)

Terraform Provided for Paralus

## Documentation

See [docs](/docs) page for a full explanation of the various datasource/resources

## Acceptance Tests

Acceptance tests (deployed under the /internal/acctest directory) can be run one of two ways

### Single Test

If you wish to run a specific acceptance test, do the following:

(Note: this assumes you are using vscode and TF is deployed as a plugin)

1. Download the config.json from Paralus UI. See [CLI](https://www.paralus.io/docs/usage/cli)
2. Put the json into the same directory as the make file
3. Create a launch.json file with the following contents

    ```json
    {
        "version": "0.2.0",
        "configurations": [
            {
                "name": "Launch a test function",
                "type": "go",
                "request": "launch",
                "mode": "auto",
                "program": "${fileDirname}",
                "env": {
                    "PKG_NAME": "${relativeFileDirname}",
                    "TF_ACC": "1",
                    "TF_LOG": "INFO",
                    "GOFLAGS": "-mod=readonly",
                    "CONFIG_JSON": "${workspaceFolder}/config.json"
                }, 
                "args": [
                    "-test.v",
                    "-test.run",
                    "^${selectedText}$"
                ],
                "showLog": true
            }
        ]
    }
    ```

4. Go into the individual test case within your go file and highlight the func name
5. Go to `Run and Debug` on the left and select `Launch a test function` from the top
6. Look at the `DEBUG_CONSOLE` window to see the result

### All Tests

To run all acceptance tests, use the make command by doing the following:

1. Download the config.json from Paralus UI. See [CLI](https://www.paralus.io/docs/usage/cli)
2. Put the json into the same directory as the make file
3. Run the command `make testacc`

Note: This will run all acceptance tests in the `internal/acctest` directory
