# Example manifests

You can run your function locally and test it using `crossplane render`
with these example manifests.

```shell
# Run the function locally
$ go run . --insecure --debug
```

```shell
# Then, in another terminal, call it with these example manifests
$ crossplane render xr.yaml composition-value.yaml functions.yaml -e resources -r
---
apiVersion: example.crossplane.io/v1
kind: XR
metadata:
  name: example-xr
spec:
  crossplane:
    resourceRefs:
    - apiVersion: sample/v1
      kind: Example
      name: test
status:
  conditions:
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    reason: WatchCircuitClosed
    status: "True"
    type: Responsive
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    reason: ReconcileSuccess
    status: "True"
    type: Synced
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    message: 'Unready resources: example-xr'
    reason: Creating
    status: "False"
    type: Ready
---
apiVersion: sample/v1
kind: Example
metadata:
  annotations:
    crossplane.io/composition-resource-name: example-xr
  labels:
    crossplane.io/composite: example-xr
  name: test
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: 0749c14a-4f6a-5a82-a298-88e58db51cf6
---
apiVersion: render.crossplane.io/v1beta1
kind: Result
message: 'Successfully selected composition: function-dynamic-required'
reason: SelectComposition
severity: Normal
---
apiVersion: render.crossplane.io/v1beta1
kind: Result
message: Composed resource "example-xr" is not yet ready
reason: ComposeResources
severity: Normal

$ crossplane render xr.yaml composition-fieldpath.yaml functions.yaml -e resources -r
---
apiVersion: example.crossplane.io/v1
kind: XR
metadata:
  name: example-xr
spec:
  crossplane:
    resourceRefs:
    - apiVersion: sample/v1
      kind: Example
      name: test
status:
  conditions:
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    reason: WatchCircuitClosed
    status: "True"
    type: Responsive
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    reason: ReconcileSuccess
    status: "True"
    type: Synced
  - lastTransitionTime: "2024-01-01T00:00:00Z"
    message: 'Unready resources: example-xr'
    reason: Creating
    status: "False"
    type: Ready
---
apiVersion: sample/v1
kind: Example
metadata:
  annotations:
    crossplane.io/composition-resource-name: example-xr
  labels:
    crossplane.io/composite: example-xr
  name: test
  ownerReferences:
  - apiVersion: example.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: XR
    name: example-xr
    uid: 0749c14a-4f6a-5a82-a298-88e58db51cf6
---
apiVersion: render.crossplane.io/v1beta1
kind: Result
message: 'Successfully selected composition: function-dynamic-required'
reason: SelectComposition
severity: Normal
---
apiVersion: render.crossplane.io/v1beta1
kind: Result
message: Composed resource "example-xr" is not yet ready
reason: ComposeResources
severity: Normal
```
