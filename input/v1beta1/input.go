// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=dynamicrequired.fn.dev.devoba.de
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Input is used to configure the function
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Input struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec Spec `json:"spec"`
}

// Spec holds the actual configuration for the function.
type Spec struct {
	// RequiredResources holds the information how to fill the required resources
	RequiredResources []RequiredResource `json:"requiredResources"`
}

// RequiredResource describes the dynamic resource requirement.
type RequiredResource struct {
	// RequirementName is the unique name to identify this required resource in the Required Resources map in the function request
	RequirementName string `json:"requirementName"`
	// APIVersion of the required resource
	// +optional
	APIVersion *ValueReference `json:"apiVersion,omitempty"`
	// Kind of the required resource
	// +optional
	Kind *ValueReference `json:"kind,omitempty"`
	// MatchLabels specifies the set of labels to match for finding the required resource. When specified, Name is ignored
	// +optional
	MatchLabels *[]MatchLabelConfig `json:"matchLabels,omitempty"`
	// Name of the required resource
	// +optional
	Name *ValueReference `json:"name,omitempty"`
	// Namespace of the required resource if it is namespaced
	// +optional
	Namespace *ValueReference `json:"namespace,omitempty"`
}

// MatchLabelConfig defines a label key and value for resources to match.
type MatchLabelConfig struct {
	// Key is the label key
	Key ValueReference `json:"key"`
	// Value is the label value
	Value ValueReference `json:"value"`
}

// ReferenceType specifies how a reference is resolved.
type ReferenceType string

const (
	// ReferenceTypeValue by a static value.
	ReferenceTypeValue ReferenceType = "Value"
	// ReferenceTypeFieldPath by a JSONPath selector selecting the value from the composite resource or managed resource object.
	ReferenceTypeFieldPath ReferenceType = "FieldPath"
	// ReferenceTypeEnvironment by an environment key.
	ReferenceTypeEnvironment ReferenceType = "Environment"
)

// ValueReference is used to determine where the value for the field is gathered from.
type ValueReference struct {
	// Type sets the type of this value reference
	// +optional
	// +kubebuilder:validation:Enum=Environment;FieldPath;Value
	// +kubebuilder:default=Value
	Type *ReferenceType `json:"type"`
	// Environment is the environment key the value is fetched from
	Environment *string `json:"environment,omitempty"`
	// FieldPath is the JSONPath selector selecting the value from the composite resource or managed resource object
	FieldPath *string `json:"fieldPath,omitempty"`
	// Value is a static value string
	Value *string `json:"value,omitempty"`
}
