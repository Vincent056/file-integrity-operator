# file-integrity-operator
The file-integrity-operator is a OpenShift Operator that continually runs file integrity checks on the cluster nodes. It deploys a DaemonSet that initializes and runs privileged AIDE ([Advanced Intrusion Detection Environment](https://aide.github.io)) containers on each node, providing a log of files that have been modified since the initial run of the DaemonSet pods.

### Deploying from source:
Default upstream images:
```
$ (clone repo)
$ oc create -f deploy/crds/
$ oc create -f deploy/
```

Images built from HEAD:
```
$ (clone repo)
$ oc login -u kubeadmin -p <pw>
$ make deploy-to-cluster
```

### Deploying from OLM:
```
$ (clone repo)
$ oc login -u kubeadmin -p <pw>
$ oc create namespace openshift-file-integrity
$ oc create -f deploy/olm-catalog/operator-source.yaml
```

### FileIntegrity API:

The operator works with `FileIntegrity` objects. Each of these objects represents a managed deployment of AIDE on one or more nodes.

```
apiVersion: fileintegrity.openshift.io/v1alpha1
kind: FileIntegrity
metadata:
  name: example-fileintegrity
  namespace: openshift-file-integrity
spec:
  nodeSelector:
    kubernetes.io/hostname: "ip-10-10-10-1"
  tolerations:
  - key: "myNode"
    operator: "Exists"
    effect: "NoSchedule"
  config:
    name: "myconfig"
    namespace: "openshift-file-integrity"
    key: "config"
    gracePeriod: 20
  debug: false
status:
  phase: Active
```
In the `spec`:
* **nodeSelector**: Selector for nodes to schedule the scan instances on.
* **tolerations**: Specify tolerations to schedule on nodes with custom taints. When not specified, a default toleration allowing running on master nodes is applied.
* **config**: Point to a ConfigMap containing an AIDE configuration to use instead of the CoreOS optimized default. See "Applying an AIDE config" below.
* **config.gracePeriod**: The number of seconds to pause in between AIDE integrity checks. Frequent AIDE checks on a node may be resource intensive, so it can be useful to specify a longer interval. Defaults to 900 (15 mins).

In the `status`:
* **phase**: The running status of the `FileIntegrity` instance. Can be `Initializing`, `Pending`, or `Active`. `Initializing` is displayed if the FileIntegrity is currently initializing or re-initializing the AIDE database, `Pending` if the FileIntegrity deployment is still being created, and `Active` if the scans are active and ongoing. For node scan results, see the `FileIntegrityNodeStatus` objects explained below.

### Usage:

After deploying the operator, you must create a `FileIntegrity` object. The following example will enable scanning on all nodes.
```
apiVersion: fileintegrity.openshift.io/v1alpha1
kind: FileIntegrity
metadata:
  name: example-fileintegrity
  namespace: openshift-file-integrity
spec:
  config: {}
```

Viewing the scan phase: An `Active` phase indicates that on each node, the AIDE database has been initialized and periodic scanning is enabled:
```
$ oc get fileintegrities -n openshift-file-integrity
NAME                    AGE
example-fileintegrity   11m

$ oc get fileintegrities/example-fileintegrity -n openshift-file-integrity -o jsonpath="{ .status.phase }"
Active
```

Each node will have a corresponding `FileIntegrityNodeStatus` object:
```
$ oc get fileintegritynodestatuses
NAME                                                               AGE
example-fileintegrity-ip-10-0-139-137.us-east-2.compute.internal   4h24m
example-fileintegrity-ip-10-0-140-35.us-east-2.compute.internal    4h24m
example-fileintegrity-ip-10-0-162-216.us-east-2.compute.internal   4h24m
example-fileintegrity-ip-10-0-172-188.us-east-2.compute.internal   4h24m
example-fileintegrity-ip-10-0-210-181.us-east-2.compute.internal   4h24m
example-fileintegrity-ip-10-0-210-89.us-east-2.compute.internal    4h24m
```

The `results` field can contain up to three entries. The most recent Successful scan, the most recent Failed scan (if any), and the most recent Errored scan (if any). When there are multiple entries, the newest `lastProbeTime` indicates the current status.

A Failed scan indicates that there were changes to the files that AIDE monitors, and displays a brief status. The `resultConfigMap` fields point to a ConfigMap containing a more detailed report.

Note: Currently the failure log is only exposed to the admin through this result ConfigMap. In order to provide some permanence of record, the result ConfigMaps are not owned by the FileIntegrity object, so manual cleanup is necessary. Additionally, deleting the FileIntegrity object leaves the AIDE database on the nodes, and the scan state will resume if the FileIntegrity is re-created.

```
$ oc get fileintegritynodestatus/example-fileintegrity-ip-10-0-139-137.us-east-2.compute.internal -o yaml
apiVersion: fileintegrity.openshift.io/v1alpha1
kind: FileIntegrityNodeStatus
...
nodeName: ip-10-0-139-137.us-east-2.compute.internal
results:
- condition: Succeeded
  lastProbeTime: "2020-06-18T01:17:14Z"
- condition: Failed
  filesAdded: 1
  filesChanged: 1
  lastProbeTime: "2020-06-18T01:28:57Z"
  resultConfigMapName: aide-ds-example-fileintegrity-ip-10-0-139-137.us-east-2.compute.internal-failed
  resultConfigMapNamespace: openshift-file-integrity

$ oc get cm/aide-ds-example-fileintegrity-ip-10-0-139-137.us-east-2.compute.internal-failed -n openshift-file-integrity -o jsonpath="{ .data.integritylog }"
AIDE 0.15.1 found differences between database and filesystem!!
Start timestamp: 2020-06-18 02:00:38

Summary:
  Total number of files:        29447
  Added files:                  1
  Removed files:                0
  Changed files:                1


---------------------------------------------------
Added files:
---------------------------------------------------

added: /hostroot/root/.bash_history

---------------------------------------------------
Changed files:
---------------------------------------------------

changed: /hostroot/etc/resolv.conf

---------------------------------------------------
Detailed information about changes:
---------------------------------------------------


File: /hostroot/etc/resolv.conf
 SHA512   : Xl2pzxjmRPtW8bl6Kj49SkKOSBVJgsCI , tebxD8QZd/5/SqsVkExCwVqVO22zxmcq
```
AIDE logs over 1MB are gzip compressed and base64 encoded, due to the configMap data size limit. In this case, you will want to pipe the output of the above command to `base64 -d | gunzip`.  Compressed logs are indicated by the presense of a `file-integrity.openshift.io/compressed` annotation key in the configMap.

### Events

Transitions in the status of the FileIntegrity and FileIntegrityNodeStatus objects are also logged by events. The creation time of the event reflects the latest transition (i.e., Initializing to Active), and not necessarily the latest scan result. However, the newest event will always reflect the most recent status.
```
$ oc get events --field-selector reason=FileIntegrityStatus
LAST SEEN   TYPE     REASON                OBJECT                                MESSAGE
97s         Normal   FileIntegrityStatus   fileintegrity/example-fileintegrity   Pending
67s         Normal   FileIntegrityStatus   fileintegrity/example-fileintegrity   Initializing
37s         Normal   FileIntegrityStatus   fileintegrity/example-fileintegrity   Active
```

When a node has a failed scan, an event is created with the add/changed/removed and configMap information.
```
$ oc get events --field-selector reason=NodeIntegrityStatus
LAST SEEN   TYPE      REASON                OBJECT                                MESSAGE
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-134-173.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-168-238.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-169-175.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-152-92.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-158-144.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-131-30.ec2.internal
87m         Warning   NodeIntegrityStatus   fileintegrity/example-fileintegrity   node ip-10-0-152-92.ec2.internal has changed! a:1,c:1,r:0 log:openshift-file-integrity/aide-ds-example-fileintegrity-ip-10-0-152-92.ec2.internal-failed
```

Changes to the number of added/changed/removed files will result in a new event, even if the status of the node has not transitioned.
```
$ oc get events --field-selector reason=NodeIntegrityStatus
LAST SEEN   TYPE      REASON                OBJECT                                MESSAGE
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-134-173.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-168-238.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-169-175.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-152-92.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-158-144.ec2.internal
114m        Normal    NodeIntegrityStatus   fileintegrity/example-fileintegrity   no changes to node ip-10-0-131-30.ec2.internal
87m         Warning   NodeIntegrityStatus   fileintegrity/example-fileintegrity   node ip-10-0-152-92.ec2.internal has changed! a:1,c:1,r:0 log:openshift-file-integrity/aide-ds-example-fileintegrity-ip-10-0-152-92.ec2.internal-failed
40m         Warning   NodeIntegrityStatus   fileintegrity/example-fileintegrity   node ip-10-0-152-92.ec2.internal has changed! a:3,c:1,r:0 log:openshift-file-integrity/aide-ds-example-fileintegrity-ip-10-0-152-92.ec2.internal-failed
```

### Local testing
```
$ make run
```

### Running the end-to-end suite
```
$ make e2e
```

## Overriding the AIDE configuration
By default the AIDE containers run with an aide.conf that is tailored to a default RHCOS node. If you need to add or exclude files on nodes that are not covered by the default config, you can override it with a modified config.

- Create a ConfigMap containing the aide.conf, e.g.,
```
$ oc project openshift-file-integrity
$ oc create configmap myconf --from-file=aide-conf=aide.conf.rhel8
```
- Post the `FileIntegrity` CR containing the name, namespace, and data key containing the aide.conf in the spec.
```
apiVersion: file-integrity.openshift.io/v1alpha1
kind: FileIntegrity
metadata:
  name: example-fileintegrity
  namespace: openshift-file-integrity
spec:
  config:
    name: myconf
    namespace: openshift-file-integrity
    key: aide-conf
```
* At this point the operator will update the active AIDE config and perform a re-initialization of the AIDE database, as well as a restart of the AIDE pods to begin scanning with the new configuration. A backup of the logs and database from the previously applied configurations are left available on the nodes under /etc/kubernetes.
* The operator automatically converts the `database`, `database_out`, `report_url`, `DBDIR`, and `LOGDIR` options in the configuration to accommodate running inside of a pod.
* Removing the config section from the FileIntegrity resource when active reverts the running config to the default and re-initializes the database.
* In the case of where small modifications are needed (such as excluding a file or directory), it's recommended to copy the [default config](https://github.com/openshift/file-integrity-operator/blob/master/pkg/controller/fileintegrity/config_defaults.go#L16) to a new ConfigMap and add to it as needed.
* Some AIDE configuration options may not be supported by the AIDE container. For example, the `mhash` digest types are not supported. For digest selection, it is recommended to use the default config's [CONTENT_EX group](https://github.com/openshift/file-integrity-operator/blob/master/pkg/controller/fileintegrity/config_defaults.go#L25).
* Manually re-initializing the AIDE database can be done by adding the annotation key `file-integrity.openshift.io/re-init` to the FileIntegrity object.
