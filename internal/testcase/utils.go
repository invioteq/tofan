package testcase

import (
	"context"
	"encoding/json"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"github.com/invioteq/tofan/pkg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

// isResourceReady  check the specific readiness conditions relevant to testcase resources.
func isResourceReady(resource *unstructured.Unstructured) bool {

	status, found, _ := unstructured.NestedFieldNoCopy(resource.Object, "status", "conditions")
	if !found {
		return false
	}

	conditions, ok := status.([]interface{})
	if !ok {
		return false
	}

	for _, cond := range conditions {
		condition, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		if condition["status"] == "True" {
			return true
		}
	}

	return false
}
