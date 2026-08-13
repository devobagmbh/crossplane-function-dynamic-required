package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
)

const testKey = "test"

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1.RunFunctionResponse
		err error
	}

	defaultMeta := &fnv1.ResponseMeta{Ttl: &durationpb.Duration{Seconds: 60}}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NoInputSpecified": {
			reason: "The Function should return a fatal result if no input was specified",
			args: args{
				req: &fnv1.RunFunctionRequest{},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: defaultMeta,
					Results: []*fnv1.Result{
						{
							Severity: fnv1.Severity_SEVERITY_FATAL,
							Message:  "no input found in request ",
							Target:   fnv1.Target_TARGET_COMPOSITE.Enum(),
						},
					},
				},
			},
		},
		"RequireApiVersionAndKind": {
			reason: "The response does not include the apiversion and kind",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Input: resource.MustStructJSON(`{
						"apiVersion": "dynamicrequired.fn.dev.devoba.de/v1beta1",
						"kind": "Input",
						"spec": {
							"requiredResources": [{
								"requirementName": "test",
								"apiVersion": {
									"type": "Value",
									"value": "test"
								},
								"kind": {
									"type": "Value",
									"value": "test"
								}
							}]
						}
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: defaultMeta,
					Requirements: &fnv1.Requirements{
						Resources: map[string]*fnv1.ResourceSelector{
							testKey: {ApiVersion: testKey, Kind: testKey},
						},
					},
				},
			},
		},
		"RequireNamespace": {
			reason: "The response does not include the namespace",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Input: resource.MustStructJSON(`{
						"apiVersion": "dynamicrequired.fn.dev.devoba.de/v1beta1",
						"kind": "Input",
						"spec": {
							"requiredResources": [{
								"requirementName": "test",
								"namespace": {
									"type": "Value",
									"value": "test"
								}
							}]
						}
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: defaultMeta,
					Requirements: &fnv1.Requirements{
						Resources: map[string]*fnv1.ResourceSelector{
							testKey: {Namespace: new(testKey)},
						},
					},
				},
			},
		},
		"RequireName": {
			reason: "The response does not include the name",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Input: resource.MustStructJSON(`{
						"apiVersion": "dynamicrequired.fn.dev.devoba.de/v1beta1",
						"kind": "Input",
						"spec": {
							"requiredResources": [{
								"requirementName": "test",
								"name": {
									"type": "Value",
									"value": "test"
								}
							}]
						}
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: defaultMeta,
					Requirements: &fnv1.Requirements{
						Resources: map[string]*fnv1.ResourceSelector{
							testKey: {Match: &fnv1.ResourceSelector_MatchName{MatchName: testKey}},
						},
					},
				},
			},
		},
		"RequireMatchLabels": {
			reason: "The response does not include the match labels",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Input: resource.MustStructJSON(`{
						"apiVersion": "dynamicrequired.fn.dev.devoba.de/v1beta1",
						"kind": "Input",
						"spec": {
							"requiredResources": [{
								"requirementName": "test",
								"matchLabels": [{
									"key": {
										"type": "Value",
										"value": "key"
									},
									"value": {
										"type": "Value",
										"value": "value"
									}
								}]
							}]
						}
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: defaultMeta,
					Requirements: &fnv1.Requirements{
						Resources: map[string]*fnv1.ResourceSelector{
							testKey: {Match: &fnv1.ResourceSelector_MatchLabels{MatchLabels: &fnv1.MatchLabels{Labels: map[string]string{
								"key": "value",
							}}}},
						},
					},
				},
			},
		},
		"RequireWithEnvironment": {
			reason: "The response does not include the apiversion",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Input: resource.MustStructJSON(`{
						"apiVersion": "dynamicrequired.fn.dev.devoba.de/v1beta1",
						"kind": "Input",
						"spec": {
							"requiredResources": [{
								"requirementName": "test",
								"namespace": {
									"type": "Environment",
									"environment": "testkey"
								}
							}]
						}
					}`),
					Context: resource.MustStructJSON(`{
						"apiextensions.crossplane.io/environment": { "testkey": "testvalue" }
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: defaultMeta,
					Context: resource.MustStructJSON(`{
						"apiextensions.crossplane.io/environment": { "testkey": "testvalue" }
					}`),
					Requirements: &fnv1.Requirements{
						Resources: map[string]*fnv1.ResourceSelector{
							testKey: {Namespace: new("testvalue")},
						},
					},
				},
			},
		},
		"RequireWithFieldPath": {
			reason: "The response does not include the apiversion",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Input: resource.MustStructJSON(`{
						"apiVersion": "dynamicrequired.fn.dev.devoba.de/v1beta1",
						"kind": "Input",
						"spec": {
							"requiredResources": [{
								"requirementName": "test",
								"namespace": {
									"type": "FieldPath",
									"fieldPath": "metadata.name"
								}
							}]
						}
					}`),
					Observed: &fnv1.State{Composite: &fnv1.Resource{Resource: resource.MustStructJSON(`{
						"metadata": {
							"name": "test"
						}
					}`)}},
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: defaultMeta,
					Requirements: &fnv1.Requirements{
						Resources: map[string]*fnv1.ResourceSelector{
							testKey: {Namespace: new(testKey)},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}
