# Infra
<p align="center">
  <br/>
  <br/>
  <img src="https://user-images.githubusercontent.com/3325447/109728544-68423100-7b84-11eb-8fc0-759df7c3b974.png" height="128" />
  <br/>
  <br/>
</p>

* Website: https://infrahq.com
* Docs: https://infrahq.com/docs
* Slack: https://infra-slack.slack.com

## Introduction

Identity and access management for Kubernetes. Instead of creating separate credentials and writing scripts to map permissions to Kubernetes, developers & IT teams can integrate existing identity providers (Okta, Google accounts, GitHub auth, Azure active directory) to securely provide developers with access to Kubernetes.

Use cases:
* Fine-grained permissions
* Mapping existing users & groups (in Okta, Azure AD, Google, etc) into Kubernetes groups
* On-boarding and off-boarding users (automatically sync users against identity providers)
* No more out of sync Kubeconfig
* Cloud vendor-agnostic
* Coming soon: Audit logs (who did what, when)


## Documentation
* [API Documentation](https://github.com/infrahq/infra/blob/main/docs/api.md)
* [Developing Infra](https://github.com/infrahq/infra/blob/main/docs/development.md)

## Architecture

<p align="center">
  <br/>
  <br/>
  <img src="https://user-images.githubusercontent.com/251292/113448649-395cec00-93ca-11eb-9c70-ea4c5c9f82da.png" />
  <br/>
  <br/>
</p>

## Install on Kubernetes

Deploy via `kubectl`:

```
kubectl apply -f https://raw.githubusercontent.com/infrahq/infra/master/deploy/infra.yaml
```

then check the load balancer exposed by Infra Engine:

```
NAME            TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)        AGE
infra-service   LoadBalancer   10.12.11.116   31.58.101.169   80:32326/TCP   1m
```

optionally map a DNS name e.g. `infra.acme.com` to `31.58.101.169`.


## Using Infra CLI

### Installing

```
# macOS
brew cask install infra

# Windows
winget install --id infra.infra

# Linux
curl -L "https://github.com/infrahq/infra/releases/download/latest/infra-linux-$(uname -m)" -o /usr/local/bin/infra
```

### Logging in

```
$ infra login 31.58.101.169
```

### Adding a user

```
$ infra users add michael@acme.com
User michael@acme.com added with the following permissions:
USER                    PROVIDER             ROLES            NAMESPACE
michael@acme.com        local                view             default 

Please share the following login with michael@acme.com:

infra login 31.58.101.169 --token VDBkRmVWbFRNV3BhVkZKc1dtcFazlIVFhkT2FsRjNaMmRGYVUxQk1kd18jdj10
```

Note: users can also be added via `infra.yaml` (see below) for scriptability and integration into existing infrastructure as code tools such as Terraform, Ansible, Pulumi, and more.


### Listing users

List users that have been added to Infra:

```
$ infra users
USER                 PROVIDER             ROLES            NAMESPACE
admin                local                admin            default
michael@acme.com     local                view             default 
```

### Listing groups

To view groups that have been synchronized to Infra, use `infra groups`:

```
$ infra groups
NAME                  PROVIDER        USERS          ROLES
local                 local           1              view
developers@acme.com   google          1              admin
```

### Listing roles

To view all roles in the cluster, use `infra roles`:

```
$ infra roles
NAME        NAMESPACE           GRANTED GROUPS      GRANTED USERS        DESCRIPTION 
admin       default             1                   1                    Admin access
view        default             1                   1                    Read-only access
```


### Listing access 

List the user's access permissions

```
$ infra permissions -u admin

NAME                                                          LIST  CREATE  UPDATE  DELETE
daemonsets.apps                                               ✔     ✔       ✔       ✔
daemonsets.extensions                                         ✔     ✔       ✔       ✔
deletebackuprequests.velero.io                                ✔     ✔       ✔       ✔
deployments.apps                                              ✔     ✔       ✔       ✔
deployments.extensions                                        ✔     ✔       ✔       ✔
downloadrequests.velero.io                                    ✔     ✔       ✔       ✔
endpoints                                                     ✔     ✔       ✔       ✔
events                                                        ✔     ✔       ✔       ✔
events.events.k8s.io                                          ✔     ✔       ✔       ✔
felixconfigurations.crd.projectcalico.org                     ✔     ✔       ✔       ✔
pods                                                          ✔     ✔       ✔       ✔
pods.metrics.k8s.io                                           ✔                     
podsecuritypolicies.extensions                                ✔     ✔       ✔       ✔
podsecuritypolicies.policy                                    ✔     ✔       ✔       ✔
podtemplates                                                  ✔     ✔       ✔       ✔
podvolumebackups.velero.io                                    ✔     ✔       ✔       ✔
podvolumerestores.velero.io                                   ✔     ✔       ✔       ✔
priorityclasses.scheduling.k8s.io                             ✔     ✔       ✔       ✔
prometheuses.monitoring.coreos.com                            ✔     ✔       ✔       ✔
prometheusrules.monitoring.coreos.com                         ✔     ✔       ✔       ✔
replicasets.apps                                              ✔     ✔       ✔       ✔
replicasets.extensions                                        ✔     ✔       ✔       ✔
replicationcontrollers                                        ✔     ✔       ✔       ✔
resourcequotas                                                ✔     ✔       ✔       ✔
resticrepositories.velero.io                                  ✔     ✔       ✔       ✔
restores.velero.io                                            ✔     ✔       ✔       ✔
rolebindings.rbac.authorization.k8s.io                        ✔     ✔       ✔       ✔
roles.rbac.authorization.k8s.io                               ✔     ✔       ✔       ✔
runtimeclasses.node.k8s.io                                    ✔     ✔       ✔       ✔
schedules.velero.io                                           ✔     ✔       ✔       ✔
secrets                                                       ✔     ✔       ✔       ✔
selfsubjectaccessreviews.authorization.k8s.io                       ✔               
selfsubjectrulesreviews.authorization.k8s.io                        ✔               
serverstatusrequests.velero.io                                ✔     ✔       ✔       ✔
serviceaccounts                                               ✔     ✔       ✔       ✔
services                                                      ✔     ✔       ✔       ✔
statefulsets.apps                                             ✔     ✔       ✔       ✔
storageclasses.storage.k8s.io                                 ✔     ✔       ✔       ✔
studyjobs.kubeflow.org                                        ✔     ✔       ✔       ✔
subjectaccessreviews.authorization.k8s.io                           ✔               
tfjobs.kubeflow.org                                           ✔     ✔       ✔       ✔
tokenreviews.authentication.k8s.io                                  ✔               
validatingwebhookconfigurations.admissionregistration.k8s.io  ✔     ✔       ✔       ✔
volumeattachments.storage.k8s.io                              ✔     ✔       ✔       ✔
volumesnapshotlocations.velero.io                             ✔     ✔       ✔       ✔
No namespace given, this implies cluster scope (try -n if this is not intended)
```

## CLI Reference

```
$ infra
Infra: manage Kubernetes access

Usage:
  infra [command]
  infra [flags]

Available Commands:
  help          Help about any command
  users         List all users across all groups
  groups        List available groups
  roles         List available roles
  permissions   List configured permissions
  login         Log in to an Infra engine
  logout        Log out of an Infra engine

Flags:
  -h, --help   help for infra

Use "infra [command] --help" for more information about a command.
```

## Configuration

For scriptability, Infra Engine can be configured using a yaml file

```yaml
providers:
  - name: google
    kind: oidc
    config: 
      client-id: acme-12345678.apps.googleusercontent.com
      client-secret: /etc/infra/client-secret
      issuer-url: https://accounts.google.com
      redirect-url: https://infra.acme.com:3090/v1/oidc/callback
      scope: ['https://www.googleapis.com/auth/admin.directory.group.readonly', 'openid', 'email']
    groups:
      - developers@acme.com

permissions:
  - provider: google
    group: developers@acme.com
    role: admin
    namespace: default            # optional namespace
```

## Security
We take security very seriously. If you have found a security vulnerability please disclose it privately to us by email via [security@infrahq.com](mailto:security@infrahq.com)
