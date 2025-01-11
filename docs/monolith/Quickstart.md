---
title: Quickstart
weight: 1
---

# Quickstart

This guide walks through the process to install Upbound Universal Crossplane and install the AzAPI official provider.

To use AzAPI official provider, install Upbound Universal Crossplane into your Kubernetes cluster, install the `Provider`, apply a `ProviderConfig`, and create a *managed resource* in AzAPI via Kubernetes.

## Install the Up command-line
Download and install the Upbound `up` command-line.

```shell
curl -sL "https://cli.upbound.io" | sh
mv up /usr/local/bin/
```

Verify the version of `up` with `up --version`

```shell
$ up --version
v0.36.2
```

_Note_: official providers only support `up` command-line versions v0.13.0 or later.

## Install Universal Crossplane
Install Upbound Universal Crossplane with the Up command-line.

```shell
$ up uxp install
UXP 1.18.0-up.1 installed
```

Verify the UXP pods are running with `kubectl get pods -n upbound-system`

```shell
> kubectl get pods -n upbound-system
NAME                                       READY   STATUS    RESTARTS   AGE
crossplane-649f76c8db-fnhqm                1/1     Running   0          24s
crossplane-rbac-manager-645fdf89d6-dnkgj   1/1     Running   0          24s
```

## Install the official Azure provider

Install the official `provider-azapi` into the Kubernetes cluster with a Kubernetes configuration file.

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-azapi
spec:
  package: xpkg.upbound.io/upbound/provider-azapi:v0.1.0
```

Apply this configuration with `kubectl apply -f`.

After installing the provider, verify the install with `kubectl get providers`.   

```shell
> kubectl get providers
NAME             INSTALLED   HEALTHY   PACKAGE                                         AGE
provider-azapi   True        True      xpkg.upbound.io/upbound/provider-azapi:v0.1.0   39s
```

It may take up to 5 minutes to report `HEALTHY`.

## Create a Kubernetes secret
The `provider-azapi` requires credentials to create and manage AzAPI resources.

### Install the Azure command-line
Generating an [authentication file](https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authorization#use-file-based-authentication) requires the Azure command-line. Follow the documentation from Microsoft to [Download and install the Azure command-line](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli).

### Create an Azure service principal
Follow the Azure documentation to [find your Subscription ID](https://docs.microsoft.com/en-us/azure/azure-portal/get-subscription-tenant-id) from the Azure Portal.

Log in to the Azure command-line.

```command
az login
```

Using the Azure command-line and provide your Subscription ID create a service principal and authentication file.

```command
az ad sp create-for-rbac --sdk-auth --role Owner --scopes /subscriptions/<Subscription ID> 
```

The command generates a JSON file like this:
```json
{
  "clientId": "5d73973c-1933-4621-9f6a-9642db949768",
  "clientSecret": "24O8Q~db2DFJ123MBpB25hdESvV3Zy8bfeGYGcSd",
  "subscriptionId": "c02e2b27-21ef-48e3-96b9-a91305e9e010",
  "tenantId": "7060afec-1db7-4b6f-a44f-82c9c6d8762a",
  "activeDirectoryEndpointUrl": "https://login.microsoftonline.com",
  "resourceManagerEndpointUrl": "https://management.azure.com/",
  "activeDirectoryGraphResourceId": "https://graph.windows.net/",
  "sqlManagementEndpointUrl": "https://management.core.windows.net:8443/",
  "galleryEndpointUrl": "https://gallery.azure.com/",
  "managementEndpointUrl": "https://management.core.windows.net/"
}
```

Save this output as `azapi-credentials.json`.

### Create a Kubernetes secret with the AzAPI credentials JSON file
Use `kubectl create secret -n upbound-system` to generate the Kubernetes secret object inside the Kubernetes cluster.

`kubectl create secret generic azapi-secret -n upbound-system --from-file=creds=./azure-credentials.json`

View the secret with `kubectl describe secret`
```shell
$ kubectl describe secret azure-secret -n upbound-system
Name:         azapi-secret
Namespace:    upbound-system
Labels:       <none>
Annotations:  <none>

Type:  Opaque

Data
====
creds:  629 bytes
```
## Create a ProviderConfig
Create a `ProviderConfig` Kubernetes configuration file to attach the AzAPI credentials to the installed official `provider-azapi`.

```yaml
apiVersion: azapi.upbound.io/v1beta1
metadata:
  name: default
kind: ProviderConfig
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: upbound-system
      name: azapi-secret
      key: creds
```

Apply this configuration with `kubectl apply -f`.

**Note:** the `Providerconfig` value `spec.secretRef.name` must match the `name` of the secret in `kubectl get secrets -n upbound-system` and `spec.secretRef.key` must match the value in the `Data` section of the secret.

Verify the `ProviderConfig` with `kubectl describe providerconfigs`. 

```yaml
$ kubectl describe providerconfigs
Name:         default
Namespace:
API Version:  azapi.upbound.io/v1beta1
Kind:         ProviderConfig
# Output truncated
Spec:
  Credentials:
    Secret Ref:
      Key:        creds
      Name:       azapi-secret
      Namespace:  upbound-system
    Source:       Secret
```

## Create a managed resource
Create a managed resource to verify the `provider-azapi` is functioning.

This example creates an AzAPI resource.

**Note:** This example demonstrates the creation of an AzAPI resource. To create the resource below, you must provide a valid parent ID.
In this example, an existing Resource Group ID is used. Make sure to update the parent ID field according to your environment.

```yaml
apiVersion: resources.azapi.upbound.io/v1alpha1
kind: Resource
metadata:
  annotations:
    meta.upbound.io/example-id: resources/v1alpha1/resource
  labels:
    testing.upbound.io/example-name: example
  name: example
spec:
  forProvider:
    body: |-
      ${jsonencode({
          sku = {
            name = "Standard"
          }
          properties = {
            adminUserEnabled = true
          }
        })}
    location: West Europe
    name: registrytestupbound
    parentId: /subscriptions/<YOUR_SUBSCRIPTION_ID>/resourceGroups/<YOUR_RESOURCE_GROUP_NAME>
    responseExportValues:
    - properties.loginServer
    - properties.policies.quarantinePolicy.status
    tags:
      Key: Value
    type: Microsoft.ContainerRegistry/registries@2020-11-01-preview
```

**Note:** the `spec.providerConfigRef.name` must match the `ProviderConfig` `metadata.name` value.

Apply this configuration with `kubectl apply -f`.

Use `kubectl get managed` to verify resource group creation.

```shell
$ kubectl get managed
NAME                                          READY   SYNCED   EXTERNAL-NAME   AGE
resource.resources.azapi.upbound.io/example   True    True     <redacted>      3m24s
```

Provider created the resource when the values `READY` and `SYNCED` are `True`.

_Note:_ commands querying AzAPI resources may be slow to respond because of Azure API response times.

If the `READY` or `SYNCED` are blank or `False` use `kubectl describe` to understand why.

## Delete the managed resource
Remove the managed resource by using `kubectl delete -f` with the same `Resource` object file. It takes a up to 5 minutes for Kubernetes to delete the resource and complete the command.

Verify removal of the resource group with `kubectl get resourcegroup`

```shell
> kubectl get resource.resources.azapi.upbound.io/example
Error from server (NotFound): resources.resources.azapi.upbound.io "example" not found
```