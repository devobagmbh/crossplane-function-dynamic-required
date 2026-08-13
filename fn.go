package main

import (
	"context"
	"fmt"

	"github.com/crossplane/crossplane-runtime/v2/pkg/fieldpath"
	"github.com/devobagmbh/function-dynamic-required/input/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	fncontext "github.com/crossplane/function-sdk-go/context"
	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	if req.GetRequiredResources() != nil {
		if err := f.writeToContext(req, rsp); err != nil {
			response.Fatal(rsp, err)
			return rsp, nil
		}
	}

	f.log.Debug("Checking input")
	in := &v1beta1.Input{}
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from request %v", req))
		return rsp, nil
	}

	if in.Spec.RequiredResources == nil {
		response.Fatal(rsp, errors.Errorf("no input found in request %v", req))
		return rsp, nil
	}

	f.log.Debug("Fetching environment")
	env := &unstructured.Unstructured{}
	if v, ok := request.GetContextKey(req, fncontext.KeyEnvironment); ok {
		if err := resource.AsObject(v.GetStructValue(), env); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot get composition environment context key %q in request %v", fncontext.KeyEnvironment, req))
			return rsp, nil
		}
		f.log.Debug("Loaded Composition environment from Function context", "context-key", fncontext.KeyEnvironment)
	}

	f.log.Debug("Fetching composite resource")
	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get observed composite resource"))
		return rsp, nil
	}

	extraResources := make(map[string]*fnv1.ResourceSelector, len(in.Spec.RequiredResources))

	f.log.Debug("Generating required resources")
	for _, requiredResource := range in.Spec.RequiredResources {
		if (requiredResource.APIVersion != nil && requiredResource.Kind == nil) ||
			(requiredResource.APIVersion == nil && requiredResource.Kind != nil) {
			response.Fatal(rsp, errors.New("APIVersion requires Kind and vice versa"))
			return rsp, nil
		}

		selector, err := f.handleSelector(requiredResource, oxr, env)
		if err != nil {
			response.Fatal(rsp, err)
			return rsp, nil
		}
		extraResources[requiredResource.RequirementName] = selector
	}

	rsp.Requirements = &fnv1.Requirements{Resources: extraResources}

	f.log.Info("Function ran successful")

	return rsp, nil
}

// handleSelector handles a specific resource filter and returns a resource selector for it.
func (f *Function) handleSelector(requiredResource v1beta1.RequiredResource, oxr *resource.Composite, env *unstructured.Unstructured) (*fnv1.ResourceSelector, error) {
	var err error

	selector := fnv1.ResourceSelector{}

	if requiredResource.Kind != nil {
		f.log.Debug("Adding kind")
		if selector.Kind, err = f.resolve(oxr, *env, *requiredResource.Kind); err != nil {
			return nil, err
		}
	}

	if requiredResource.MatchLabels != nil {
		f.log.Debug("Adding match labels")
		if len(*requiredResource.MatchLabels) == 0 {
			f.log.Info("No labels specified. Using an empty label selector")
			selector.Match = &fnv1.ResourceSelector_MatchLabels{}
		} else {
			labels := make(map[string]string)
			for _, matchLabel := range *requiredResource.MatchLabels {
				var key string
				var value string
				if key, err = f.resolve(oxr, *env, matchLabel.Key); err != nil {
					return nil, err
				}
				if value, err = f.resolve(oxr, *env, matchLabel.Value); err != nil {
					return nil, err
				}
				f.log.Debug(fmt.Sprintf("Adding match for label %s:%s", key, value))
				labels[key] = value
			}
			selector.Match = &fnv1.ResourceSelector_MatchLabels{MatchLabels: &fnv1.MatchLabels{Labels: labels}}
		}
	} else if requiredResource.Name != nil {
		f.log.Debug("Adding name match")
		var name string
		if name, err = f.resolve(oxr, *env, *requiredResource.Name); err != nil {
			return nil, err
		}
		selector.Match = &fnv1.ResourceSelector_MatchName{MatchName: name}
	}

	if requiredResource.APIVersion != nil {
		f.log.Debug("Adding apiversion")

		if selector.ApiVersion, err = f.resolve(oxr, *env, *requiredResource.APIVersion); err != nil {
			return nil, err
		}
	}

	if requiredResource.Kind != nil {
		f.log.Debug("Adding kind")

		if selector.Kind, err = f.resolve(oxr, *env, *requiredResource.Kind); err != nil {
			return nil, err
		}
	}
	if requiredResource.Namespace != nil {
		f.log.Debug("Adding namespace")

		var namespace string
		if namespace, err = f.resolve(oxr, *env, *requiredResource.Namespace); err != nil {
			return nil, err
		}
		selector.Namespace = &namespace
	}
	return &selector, nil
}

func (f *Function) writeToContext(req *fnv1.RunFunctionRequest, rsp *fnv1.RunFunctionResponse) error {
	f.log.Info("Found required resources. Adding them to the context as apiextensions.crossplane.io/required-resources")
	requiredResources, err := request.GetRequiredResources(req)
	if err != nil {
		return errors.Wrapf(err, "can not get required resources")
	}
	requiredResourcesMap := make(map[string]any)
	for key, resources := range requiredResources {
		objects := make([]any, 0, len(resources))
		for _, r := range resources {
			objects = append(objects, r.Resource.Object)
		}
		requiredResourcesMap[key] = objects
	}
	f.log.Info("Required resources have already been set, set them as apiextensions.crossplane.io/extra-resources")
	s, err := structpb.NewStruct(requiredResourcesMap)
	if err != nil {
		return errors.Wrapf(err, "cannot create new Struct from required resources")
	}
	response.SetContextKey(rsp, "apiextensions.crossplane.io/required-resources", structpb.NewStructValue(s))
	return nil
}

// resolve takes the request and a prepared environment and resolves a reference.
func (f *Function) resolve(oxr *resource.Composite, env unstructured.Unstructured, reference v1beta1.ValueReference) (string, error) {
	var referenceType v1beta1.ReferenceType
	if reference.Type == nil {
		referenceType = v1beta1.ReferenceTypeValue
	} else {
		referenceType = *reference.Type
	}

	var from map[string]any
	var key string
	var err error

	switch referenceType {
	case v1beta1.ReferenceTypeValue:
		if reference.Value == nil {
			return "", errors.Errorf("reference has no value field: %v", reference)
		}
		f.log.Debug(fmt.Sprintf("Using static value %s", *reference.Value))
		return *reference.Value, nil
	case v1beta1.ReferenceTypeEnvironment:
		f.log.Debug("Using an environment reference")

		if from, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&env); err != nil {
			return "", errors.Wrap(err, "can not convert environment")
		}
		key = *reference.Environment
	case v1beta1.ReferenceTypeFieldPath:
		f.log.Debug("Using a fieldpath reference")

		if from, err = runtime.DefaultUnstructuredConverter.ToUnstructured(&oxr.Resource); err != nil {
			return "", errors.Wrap(err, "can not convert object")
		}
		key = *reference.FieldPath
	}

	result, err := fieldpath.Pave(from).GetString(key)
	if err != nil {
		return "", errors.Wrapf(err, "Can not find path %s in %T", key, from)
	}
	f.log.Debug(fmt.Sprintf("Found %s for key %s", result, key))
	return result, nil
}
