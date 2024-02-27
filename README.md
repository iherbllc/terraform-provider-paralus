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

#### Setting Up Environment

In order to run the tests, you will need a functioning paralus environment. You can either use one existing, or install the development environment outlined [here](https://github.com/paralus/paralus/blob/main/CONTRIBUTING.md) and [here](https://github.com/paralus/dashboard/blob/v0.1.6/CONTRIBUTING.md), for the paralus server and dashboard respectively. If installing a local one modify the `webpack.dev.js` file to use port `80` instead of `3000`. Otherwise, the pctl commands will try to use HTTPS.

*Note* if while trying to start up the dashboard you get the error: `Node Sass does not yet support your current environment: OS X 64-bit with Unsupported runtime (120)`, then downgrade nodejs to version 14 using [asdf](https://github.com/asdf-vm/asdf-nodejs). After you'll need to run `sudo yarn install` before running `yarn run start` again.

*Note*: Make sure `node -v` comes back as version 14.

#### Generating Credentials File

The provider needs to connect to the paralus environment via a config json file. You can download the file for the user from the dashboard or use the command as defined above for the test resources. See See [CLI](https://www.paralus.io/docs/usage/cli)

*Note:* Put the json into the same directory as the make file and name it `paralus.local.json`.

#### Setting Up Required Resources

Some of the tests require certain items already be setup. These are as follows:

- New Project:
        - name: acctest-donotdelete
        - description: Project used for acceptance testing
- New Cluster
        - project: acctest-donotdelete
        - name: man-acctest
        - description: Manually created for acceptance testing
- New Group:
        - name: acctest-group
        - description: For acceptance testing
- New User:
        - name: `acctest-user@example.com`
        - first name: acctest
        - last name: user
- New User:
        - name: `acctest2-user@example.com`
        - first name: acctest2
        - last name: user

Either manually create them within the dashboard, or install the [Paralus CLI](https://github.com/paralus/cli) and run the following commands to setup the resources:

1. `pctl config download http://<SERVER_URL> --to-file paralus.local.json` # Optional if you downloaded it manually earlier.
    *Note*: If you are using a local environment, you will need to change the config.json URLs to be the same as the <SERVER_URL> value above and also change the `project` value in the config file to `acctest-donotdelete`
2. `pctl create project acctest-donotdelete --desc 'Project used for acceptance testing' -c ./paralus.local.json`
3. `pctl create cluster imported man-acctest -c ./conparalus.localfig.json`
4. `pctl create group acctest-group --desc "For acceptance testing" -c ./paralus.local.json`
5. `pctl create user acctest-user@example.com --groups acctest-group -c ./paralus.local.json`
6. `pctl create user acctest2-user@example.com --groups acctest-group -c ./paralus.local.json`

*Note*: Make sure that the CLI downloaded matches the release date of the environment you will be testing against

### Deploying a test cluster

You can deploy a test cluster by following the below steps:

1. Install minikube via the steps [here](https://minikube.sigs.k8s.io/docs/start/)
2. Start minikube with desired Kubernetes version via command `minikube start --kubernetes-version=<K8S_VERSION>`
3. Grab the IP address minikube users to see the host via the command `minikube ssh "nc -vz host.minikube.internal 80"`
4. Create a new cluster manually within Paralus and download the bootstrap config
5. Update the bootstrap config by replacing the replays `addr` value for the relay-agent-config ConfigMap with the IP address you found above
    - For example, replace `"addr":"console.paralus.dev:443"` with `"addr":"192.168.65.254:443"`
6. Apply the bootstrap against the minikube instance

### Running the tests

Acceptance tests (deployed under the /internal/acctest directory) can be run one of two ways.

#### Single Test

If you wish to run a specific acceptance test, do the following:

(Note: this assumes you are using vscode and TF is deployed as a plugin)

1. Create a launch.json file with the following contents

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
                    "CONFIG_JSON": "${workspaceFolder}/paralus.local.json"
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

2. Go into the individual test case within your go file and highlight the func name
3. Go to the `Run and Debug` on the far left and select `Launch a test function` from the top (the green button)
4. Look at the `DEBUG CONSOLE` window to see the result

#### All Tests

To run all acceptance tests, use the make command by running the command `make testacc`

Note: This will run all acceptance tests in the `internal/acctest` directory

### Important Note

If when running your tests you find that you keep getting 404 errors, try removing the port from the REST_ENDPOINT and OPS_ENDPOINT.
