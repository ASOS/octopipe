<<<<<<< HEAD
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
```sh
$ octopipe create
```
**_See below for more information on the yaml schema_**

Find and replace Octopus Deploy formatted variables (`#{variablevalue}`) in deploy script files:

- All files in folder `scripts`, replacing values for the environment named `DevTest`:
    
```sh
$ octopipe sub in scripts/ DevTest
```
- Specific file(s) in folder `scripts`, replacing values for the environment named `Pre-Production`
```sh
$ octopipe sub in -f deploystep1.ps1 scripts/ Pre-Production
$ octopipe sub in -f deploystep1.ps1,deploystep2.sh scripts/ Pre-Production
```
- Only check and report back for all files in folder `scripts`, those variables which have no matching values in **octopipe.yaml** for environment `Production`
```sh
$ octopipe sub in -c scripts/ Production
```

Before subbing, octopipe creates a backup of each file in the same location with `.octopipe` appended to the name.  After making changes to your scripts, use your favourite merge tool to merge in your changes and bring back the unsubbed variables

- Clear out .octopipe files in folder `scripts` after merging
```sh
$ octopipe sub clear scripts/
```
Write project configuration data to the server:
```sh
$ octopipe put
```
### Yaml schema
```Yaml
project:
  name: Octopipe.Test.Project # the name of the project (will create new if the slug does not resolve)
  description: Project for testing the Octopipe tool # project description
  group: My Octopus Project Group # the Octopus Project Group this project will belong to
  lifecycle: Default.Lifecycle # the OCtopus Lifecycle for the Deployment Process

variables:
- name: processName
  value: kubelet # 'value' for a single variable value

- name: environmentName
  values: # 'values' for a variable with environment scopings (separated by comma), 'default' being an optional unscoped value
    DevTest,Pre-Production: devtest
    Production: pd
    default: env

- name: deployAccount
  value: azureserviceprincipal-azuresuba
  type: AzureAccount # types supported are "AzureAccount", "AWSAccount", "Certificate", String" (default)
  description: Account used for deployment to the subscription # variable description

process:
  steps:
  - name: Init
    type: PowerShell # types supported are "PowerShell", "Bash", "CSharp", "FSharp"
    file: scripts/init.ps1 # file location relative to octopipe.yaml
    
  - name: Deploy Kubernetes
    type: PowerShell
    file: scripts/deploystep1.ps1
```
### Todos

 - Add support for external secrets storage (Hashicorp Vault, Azure KeyVault)
 - Add support for script modules within the deployment process
 - Add support for selecting the Worker Pool a deployment step will run against

[License]


   [Octopus Deploy]: <https://octopus.com>
   [Cobra CLI]: <https://github.com/spf13/cobra>
   [License]: <https://github.com/ASOS/octopipe/blob/master/LICENSE>
=======
# Octopipe

Octopipe is a pipeline-as-code utility for [Octopus Deploy].  It supports writing project configuration data to Octopus before release creation as well as subbing out variables in local deploy scripts for local debugging

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
```sh
$ octopipe create
```
**_See below for more information on the yaml schema_**

Find and replace Octopus Deploy formatted variables (`#{variablevalue}`) in deploy script files:

- All files in folder `scripts`, replacing values for the environment named `DevTest`:
    
```sh
$ octopipe sub in scripts/ DevTest
```
- Specific file in folder `scripts`, replacing values for the environment named `Pre-Production` (multiple specific files separated with a comma)
```sh
$ octopipe sub in -f deploystep1.ps1 scripts/ Pre-Production
```
- Only check and report back for all files in folder `scripts`, those variables which have no matching values in **octopipe.yaml** for environment `Production`
```sh
$ octopipe sub in -c scripts/ Production
```

Before subbing, octopipe creates a backup of each file in the same location with `.octopipe` appended to the name.  After making changes to your scripts, use your favourite merge tool to merge in your changes and bring back the unsubbed variables

- Clear out .octopipe files in folder `scripts` after merging
```sh
$ octopipe sub clear scripts/
```
Write configuration data to the server:
```sh
$ octopipe put
```
### Yaml schema
```Yaml
project:
  name: Octopipe.Test.Project # the name of the project (will create new if the slug does not resolve)
  description: Project for testing the Octopipe tool # project description
  group: My Octopus Project Group # the Octopus Project Group this project will belong to
  lifecycle: Default.Lifecycle # the OCtopus Lifecycle for the Deployment Process

variables:
- name: processName
  value: kubelet # 'value' for a single variable value

- name: environmentName
  values: # 'values' for a variable with environment scopings (separated by comma), 'default' being an optional unscoped value
    DevTest,Pre-Production: devtest
    Production: pd
    default: env

- name: deployAccount
  value: azureserviceprincipal-azuresuba
  type: AzureAccount # types supported are "AzureAccount", "AWSAccount", "Certificate", String" (default)
  description: Account used for deployment to the subscription # variable description

process:
  steps:
  - name: Init
    type: PowerShell # types supported are "PowerShell", "Bash", "CSharp", "FSharp"
    file: scripts/init.ps1 # file location relative to octopipe.yaml
    
  - name: Deploy Kubernetes
    type: PowerShell
    file: scripts/deploystep1.ps1
```
### Todos

 - Add support for external secrets storage (Hashicorp Vault, Azure KeyVault)
 - Add support for script modules within the deployment process

[License]


   [Octopus Deploy]: <https://octopus.com>
   [Cobra CLI]: <https://github.com/spf13/cobra>
   [License]: <https://github.com/ASOS/octopipe/blob/master/LICENSE>
>>>>>>> e1043df166d1abb9afaf436a2e1de194d678e715
