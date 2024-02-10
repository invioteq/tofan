package testcase

import (
	"context"
	"encoding/json"
	"fmt"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"github.com/invioteq/tofan/pkg/constants"
	"github.com/invioteq/tofan/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"
	"strings"
)

func (r *Reconciler) ProcessTestCase(ctx context.Context, objectTemplate *tofaniov1alpha1.ObjectTemplate, testCase *tofaniov1alpha1.TestCase) error {
	for _, field := range testCase.Spec.DynamicFields {
		// `field.Values` is now a map, so iterate through the map
		for key, jsonValue := range field.Values {
			rawJSONValue, err := jsonValue.MarshalJSON()
			if err != nil {
				r.Log.Error(err, "Failed to marshal dynamic field value to JSON", "Path", field.Path, "Key", key)
				continue // Move to the next value or field on error
			}

			// Apply each value to the template independently
			modifiedTemplate, err := r.ApplyValueToTemplate(objectTemplate, field.Path, rawJSONValue)
			if err != nil {
				r.Log.Error(err, "Failed to apply dynamic field value to template", "Path", field.Path, "Value", rawJSONValue)
				continue // Move to the next value or field on error
			}

			r.Log.Info("Successfully applied value to template", "Path", field.Path, "Key", key, "ModifiedTemplate", string(modifiedTemplate))

			// create or update the resource based on the modified template
			// This involves converting the JSON back into a Kubernetes object and applying it
			err = r.ApplyObjectToCluster(ctx, modifiedTemplate, testCase.GetName())
			if err != nil {
				r.Log.Error(err, "Failed to apply object to cluster", "ModifiedTemplate", modifiedTemplate)
				return err
			}
		}
	}
	return nil
}

// ApplyValueToTemplate applies dynamic field value to the specified ObjectTemplate
func (r *Reconciler) ApplyValueToTemplate(objectTemplate *tofaniov1alpha1.ObjectTemplate, path string, value []byte) ([]byte, error) {
	var templateMap map[string]interface{}
	if err := json.Unmarshal(objectTemplate.Spec.Template.Raw, &templateMap); err != nil {
		r.Log.Error(err, "Failed to unmarshal ObjectTemplate into map")
		return nil, err
	}

	// Deserialize the Raw content of runtime.RawExtension (now referred to as value) to the expected type
	var actualValue interface{}
	if err := json.Unmarshal(value, &actualValue); err != nil {
		r.Log.Error(err, "Failed to unmarshal value", "Path", path)
		return nil, err
	}

	// Navigate and apply the deserialized value to the specified path
	if err := utils.NavigateAndApplyValue(&templateMap, path, actualValue); err != nil {
		r.Log.Error(err, "Failed to apply value to path", "Path", path, "Value", actualValue)
		return nil, err
	}

	// Generate a unique name for the object if it has a metadata.name field
	if metadata, ok := templateMap["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			uniqueName := name + "-" + utils.GenerateRandomString(5) // Ensure unique naming
			metadata["name"] = uniqueName
			r.Log.Info("Set unique name for object", "Name", uniqueName)
		}
	}

	// Re-serialize the modified map back to JSON
	modifiedTemplate, err := json.Marshal(templateMap)
	if err != nil {
		r.Log.Error(err, "Failed to marshal modified template back to JSON")
		return nil, err
	}

	return modifiedTemplate, nil
}

func (r *Reconciler) ApplyObjectToCluster(ctx context.Context, objJSON []byte, testCaseName string) error {
	// First, convert JSON to YAML because some Kubernetes APIs expect YAML
	objJSON, err := yaml.YAMLToJSON(objJSON)
	if err != nil {
		r.Log.Error(err, "Failed to convert object YAML to JSON")
		return err
	}

	// Decode the JSON into an unstructured.Unstructured object
	var unstrObj unstructured.Unstructured
	if err := json.Unmarshal(objJSON, &unstrObj); err != nil {
		r.Log.Error(err, "Failed to unmarshal JSON into Unstructured object")
		return err
	}

	// Set GVK from the unstructured object itself
	gvk := unstrObj.GroupVersionKind()

	// Prepare the object for the Create or Update operation
	unstrObj.SetGroupVersionKind(gvk)
	if unstrObj.GetNamespace() == "" {
		unstrObj.SetNamespace("default")
	}
	// Prepare the resource name
	if unstrObj.GetName() == "" {
		// Generate a unique name if not provided, for example:
		unstrObj.SetName("testcase-" + utils.GenerateRandomString(5))
	}

	// Set tofan.io/testcase-name: testcase.metadata.name
	labels := unstrObj.GetLabels()
	if labels == nil {
		labels = make(map[string]string) // Initialize if nil
	}
	// Set or update the label with the object's name.
	labels[constants.TofanTestCaseNameLabel] = testCaseName
	unstrObj.SetLabels(labels)

	// Check if the object already exists
	namespacedName := client.ObjectKey{Namespace: unstrObj.GetNamespace(), Name: unstrObj.GetName()}
	var existing unstructured.Unstructured
	existing.SetGroupVersionKind(gvk)
	err = r.Client.Get(ctx, namespacedName, &existing)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Resource does not exist, so create it
			if err := r.Client.Create(ctx, &unstrObj); err != nil {
				r.Log.Error(err, "Failed to create new resource")
				return err
			}
			r.Log.Info("Successfully created new resource", "GVK", gvk, "Name", unstrObj.GetName())
			return nil
		} else {
			// An actual error occurred other than Not Found
			r.Log.Error(err, "Failed to get existing resource")
			return err
		}
	} else {
		// Resource exists, update it
		unstrObj.SetResourceVersion(existing.GetResourceVersion())
		if err := r.Client.Update(ctx, &unstrObj); err != nil {
			r.Log.Error(err, "Failed to update existing resource")
			return err
		}
		r.Log.Info("Successfully updated existing resource", "GVK", gvk, "Name", unstrObj.GetName())
		return nil
	}
}

// TeardownResourcesForTestCase deletes all resources associated with a given TestCase, using objTpl to identify resource types.
func (r *Reconciler) TeardownResourcesForTestCase(ctx context.Context, testCase *tofaniov1alpha1.TestCase, objTpl *tofaniov1alpha1.ObjectTemplate) error {
	cfg, err := config.GetConfig()
	if err != nil {
		r.Log.Error(err, "Failed to get cluster config")
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		r.Log.Error(err, "Failed to create dynamic client")
		return err
	}
	// Construct the GroupVersionResource from ObjectTemplate status information
	gvr := schema.GroupVersionResource{
		Group:    objTpl.Status.Group,
		Version:  objTpl.Status.Version,
		Resource: fmt.Sprintf("%ss", strings.ToLower(objTpl.Status.Kind)), // Assuming simple pluralization
	}

	// Matching labels indicating they belong to the testCase
	labelSelector := fmt.Sprintf("%s=%s", constants.TofanTestCaseNameLabel, testCase.Name)
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := dynamicClient.Resource(gvr).Namespace(testCase.Namespace).DeleteCollection(ctx, deleteOptions, metav1.ListOptions{LabelSelector: labelSelector}); err != nil {
		r.Log.Error(err, "Failed to delete resources for testCase", "TestCase", testCase.Name, "GVR", gvr)
		return err
	}

	r.Log.Info("Successfully deleted resources for testCase", "TestCase", testCase.Name, "GVR", gvr)
	return nil
}
