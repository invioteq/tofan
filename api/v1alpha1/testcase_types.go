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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TestCaseSpec defines the desired state of TestCase
type TestCaseSpec struct {
	// Reference to a ObjectTemplate
	ObjectTemplateRef objectTemplateReference `json:"objectTemplateRef,omitempty"`
	// Action specifies the operation to perform with the ObjectTemplate (e.g., create, delete)
	Action string `json:"action,omitempty"`
	// Count specifies the number of instances to create/delete
	Count int `json:"count"`
	// Concurrency specifies how many operations can be performed concurrently
	Concurrency int `json:"concurrency"`
	// DynamicFields specifies how to dynamically set fields in the ObjectTemplate based on the test case.
	DynamicFields []DynamicField `json:"dynamicFields,omitempty"`
	// TargetMetrics defines the metrics that should be collected during the test
	TargetMetrics []MetricTarget `json:"targetMetrics,omitempty"`
}

// DynamicField defines a field to dynamically set based on TestCase parameters.
type DynamicField struct {
	// Path specifies the JSON path to the field within the ObjectTemplate that needs to be dynamically set.
	Path string `json:"path"`

	// Values are the values to apply to the dynamic field as simple strings.
	Values []string `json:"values"`
}

// objectTemplateReference
type objectTemplateReference struct {
	// Name of the ObjectTemplate.
	Name string `json:"name,omitempty"`
	// Kind specifies the kind of the referenced resource, which should be "ObjectTemplate".
	Kind string `json:"kind,omitempty"`
	// Group is the API group of the SpaceTemplate,  "tofan.io/v1alpha1".
	Group string `json:"group,omitempty"`
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
	// Conditions List of status conditions to indicate the status of Space
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",description="Ready"

// TestCase is the Schema for the testcases API
type TestCase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TestCaseSpec   `json:"spec,omitempty"`
	Status TestCaseStatus `json:"status,omitempty"`
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
