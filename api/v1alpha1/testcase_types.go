/*
Copyright 2024 invioteq llc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TestCaseSpec defines the desired state of TestCase
type TestCaseSpec struct {
	// ObjectTemplateRef is a reference to an ObjectTemplate resource
	ObjectTemplateRef objectTemplateReference `json:"objectTemplateRef,omitempty"`
	// Action specifies the operation to perform on the objectTemplate, e.g., create
	Action string `json:"action,omitempty"`
	// Count specifies the number of instances to create
	Count int `json:"count"`
	// Concurrency controls how many instances can be created concurrently
	Concurrency int `json:"concurrency"`
	// DynamicFields allow specifying dynamic values for certain fields in the object template
	DynamicFields []DynamicField `json:"dynamicFields,omitempty"`
	// ReadinessCriteria specifies the conditions that must be met for the resources to be considered ready
	ReadinessCriteria []ResourceReadinessCriteria `json:"readinessCriteria"`
	// TargetMetrics defines the metrics that should be collected during the test
	TargetMetrics []MetricTarget `json:"targetMetrics,omitempty"`
}

// DynamicField defines a field to dynamically set based on TestCase parameters.
type DynamicField struct {
	// Path is the JSON path to the field within the object template that needs a dynamic value.
	Path string `json:"path"`
	// Values are the dynamic values to apply to the specified path.
	Values map[string]extv1.JSON `json:"values"`
}

// objectTemplateReference
type objectTemplateReference struct {
	// Name is the name of the ObjectTemplate resource
	Name string `json:"name,omitempty"`
	// Kind specifies the kind of the referenced resource, which should be "ObjectTemplate".
	Kind string `json:"kind,omitempty"`
	// Group is the API group of the SpaceTemplate,  "tofan.io/v1alpha1".
	Group string `json:"group,omitempty"`
}

// ResourceReadinessCriteria defines readiness criteria for a specific resource type
type ResourceReadinessCriteria struct {
	// ResourceType is the kind of resource to check for readiness, e.g., Deployment, CustomResource
	ResourceType string `json:"resourceType"`
	// Conditions are the specific conditions to be met for the resource to be considered ready
	Conditions []ResourceCondition `json:"conditions"`
}

// ResourceCondition defines a single condition to check on a resource for readiness
type ResourceCondition struct {
	// Type is the type of condition, e.g., Available for Deployments, Ready for custom resources
	Type string `json:"type"`
	// Status is the status value the condition must have, e.g., "True", "False"
	Status string `json:"status"`
	// JSONPath (optional) is a JSON path expression to locate the condition value within a custom resource status
	JSONPath string `json:"jsonPath,omitempty"`
}

// MetricTarget defines a target metric for collection by the testCase
type MetricTarget struct {
	// Name is the name of the metric
	Name string `json:"name"`

	// Expr is the expression used to calculate or define the metric
	Expr string `json:"expr"`
}

// TestCaseStatus defines the observed state of TestCase
type TestCaseStatus struct {
	// Phase indicates the testcase exec phase
	Phase string `json:"phase,omitempty"`
	// Conditions List of status conditions to indicate the status of Space
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",description="Ready"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// TestCase is the Schema for the testcases API
type TestCase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TestCaseSpec   `json:"spec,omitempty"`
	Status TestCaseStatus `json:"status,omitempty"`
}

func (in *TestCase) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

func (in *TestCase) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// TestCaseList contains a list of TestCase
type TestCaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestCase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TestCase{}, &TestCaseList{})
}
