# function-dynamic-required

Adding required resources with support for dynamic references.

## Introduction

The support for [required resources](https://docs.crossplane.io/v2.3/composition/compositions/#required-resources) 
in Crossplane is currently limited to static values. So compositions using required
resources lack support for configuration through composite resources based on their spec. Dynamic values require writing
a [composition function](https://docs.crossplane.io/v2.3/composition/compositions/#dynamic-resource-requests), which
might be too complex for most cases.

This function solves that limitation by configuring required resources with the support for referencing values from
the composite resource based on fieldpaths and from environment configs based on keys. After resolving those values,
the required resources are configured for the composition resource and loaded by Crossplane.

## Usage

Reference the function in a pipeline step like this:

```yaml
  - step: run-the-template
    functionRef:
      name: function-dynamic-required
    input:
      apiVersion: dynamicrequired.fn.dev.devoba.de/v1beta1
      kind: Input
      spec:
        requiredResources:
          - requirementName: "test"
            apiVersion:
              type: "FieldPath"
              fieldPath: "spec.apiVersion"
            kind:
              type: "FieldPath"
              fieldPath: "spec.kind"
```

In this case a required resource configuration named "test" will be added, gathering resources by APIVersion 
and Kind based on values from the composition resource set in the spec.

All of the required resource filters are available:

* apiVersion (requires kind)
* kind (requires apiVersion)
* matchLabels (a list of label filters as objects with the properties "key" and "value"; will ignore name filter if used)
* name (only if matchLabels is not used)
* namespace

Each of these filters (and the key/value properties for matchLabels respectively) are configured using an object with
these properties:

* type: type of the reference. Valid values are: Value, FieldPath, Environment
* value: a static value (type set to Value)
* fieldPath: the field path referencing the value from the composition resource (type set to FieldPath)
* environment: the environment key referencing the value

## Context support

Additionally, this function will write the gathered required resource into the context under the key
`apiextensions.crossplane.io/required-resources` to use in other functions which don't support accessign required
resources directly.

See the manifests in the `example` folder for details of using this with the go-template function.