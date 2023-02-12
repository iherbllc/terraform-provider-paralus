# Paralus terraform provider

[![Action](https://github.com/iherbllc/terraform-provider-paralus/workflows/build/badge.svg)](https://github.com/iherbllc/terraform-provider-paralus/workflows/build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/iherbllc/terraform-provider-paralus)](https://goreportcard.com/badge/github.com/iherbllc/terraform-provider-paralus)
[![Go Version](https://img.shields.io/github/go-mod/go-version/iherbllc/terraform-provider-paralus)](https://img.shields.io/github/go-mod/go-version/iherbllc/terraform-provider-paralus)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/iherbllc/terraform-provider-paralus?tab=overview)

Terraform Provided for Paralus

## Documentation

See [docs](/docs) page for a full explanation of the various datasource/resources

## Acceptance Tests

### Pre-requisites

In order to run the tests, you will need a functioning paralus environment. You can either use one existing, or install the development environment outlined [here](https://github.com/paralus/paralus/blob/main/CONTRIBUTING.md) and [here](https://github.com/paralus/dashboard/blob/v0.1.6/CONTRIBUTING.md), for the paralus server and dashboard respectively.

### Running the tests

Acceptance tests (deployed under the /internal/acctest directory) can be run one of two ways.

*Note:* Some of the tests require certain items already be setup. These are as follows:

- New Project:
        - name: acctest-donotdelete
        - description: Project used for acceptance testing
- New Cluster
        - project: acctest-donotdelete
        - name: man-acctest
        - description: Manually created for acceptance testing
- New User:
        - name: acctest-user@example.com
        - first name: acctest
        - last name: user
- New User:
        - name: acctest2-user@example.com
        - first name: acctest2
        - last name: user
- New Group:
        - name: acctest-group
        - description: For acceptance testing

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
2. Put the json into the same directory as the make file. Make sure to name it `config.json` or update the name in the makefile
3. Run the command `make testacc`

Note: This will run all acceptance tests in the `internal/acctest` directory
