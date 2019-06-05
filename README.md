# Octopipe

Octopipe is a pipeline-as-code utility for [Octopus Deploy].  It supports writing project configuration data to Octopus before release creation as well as subbing out variables in deploy scripts for local debugging

### Build

Octopipe is based on the [Cobra CLI] framework, all other dependencies ship with Go

```sh
$ cd octopipe
$ go get
$ go build main.go
```

### Usage

Octopipe expects the presence of OCTOPUS_URI and OCTOPUS_API_KEY environment variables in order to know which server to talk to and for authorisation
```sh
$ export OCTOPUS_URI=https://myoctopus.server.com
$ export OCTOPUS_API_KEY=API-1A2B3C4D5E6F7G8H9I0J
$ octopipe --help
```
Create an octopipe.yaml file:

- From scratch:
```sh
$ octopipe create
```
- From an existing Octopus project:
```sh
$ octopipe create -i My.Octopus.Project
```
**_See below for more information on the yaml schema_**

Find and replace Octopus Deploy formatted variables (`#{variablevalue}`) in deploy script files:

- All files in folder `scripts`, replacing values for the environment named `DevTest`:
    
```sh
$ octopipe sub scripts/ 'Environment=DevTest'
```
- Specific file(s) in folder `scripts`, replacing values for the environment named `Pre-Production`:
```sh
$ octopipe sub -f deploystep1.ps1 scripts/ 'Environment=Pre-Production'
```
- Specific file(s) in folder `scripts`, replacing values for the environment named `Pre-Production` and the machine named `deploynode01`:
```sh
$ octopipe sub -f deploystep1.ps1,deploystep2.sh scripts/ 'Environment=Pre-Production,Machine=deploynode01'
```
- Only check and report back for all files in folder `scripts`, those variables which have no matching values in **octopipe.yaml** for environment `Production`:
```sh
$ octopipe sub -c scripts/ 'Environment=Production'
```

Before subbing, octopipe creates a backup of each file in the same location with `.octopipe` appended to the name.  After making changes to your scripts, use your favourite merge tool to merge in your changes and bring back the unsubbed variables

Clear out .octopipe files in folder `scripts` after merging:
```sh
$ octopipe sub clear scripts/
```
Write project configuration data to the server:
```sh
$ octopipe put
```
### Yaml schema

**_For interoperability with the Octopus API, types are case sensitive_**

```Yaml
project:
  name: Octopipe.Test.Project # the name of the project (will create new if the slug does not resolve)
  description: Project for testing the Octopipe tool # project description
  group: My Octopus Project Group # the Octopus Project Group this project will belong to
  lifecycle: Default.Lifecycle # the Octopus Lifecycle for the Deployment Process
  tenanted: TenantedOrUntenanted # if not specified in octopipe.yaml, the default is Untenanted. Valid tenancy types are Tenanted, Untenanted, TenantedOrUntenanted

variables:
- name: processName
  value: kubelet # 'value' for a single variable value

- name: environmentName
  scopedValues: # 'scopedValues' for a variable with scopings
    - value: devtest
      Environment: DevTest,Pre-Production
    - value: pd
      Environment: Production
      TenantTag: Azure Regions/West Europe # valid scope types are Environment, Machine, TenantTag, Channel, Action, Role
      Role: web-server
    - value: env # default unscoped value for the variable

- name: deployAccount
  value: azureserviceprincipal-azuresub
  type: AzureAccount # valid variable types are AzureAccount, AWSAccount, Certificate, String (default)
  description: Account used for deployment to the subscription # variable description

process:
  steps:
  - name: Init
    type: PowerShell # valid script types are PowerShell, Bash, CSharp, FSharp
    file: scripts/init.ps1 # file location relative to octopipe.yaml
    
  - name: Deploy Kubernetes
    type: PowerShell
    file: scripts/deploystep1.ps1
```
### Todos

 - Add support for external secrets storage (Hashicorp Vault, Azure KeyVault)
 - Add support for script modules within the deployment process
 - Add support for selecting the Worker Pool a deployment step will run against
 - Add support for step sub-actions

[License]


   [Octopus Deploy]: <https://octopus.com>
   [Cobra CLI]: <https://github.com/spf13/cobra>
   [License]: <https://github.com/ASOS/octopipe/blob/master/LICENSE>