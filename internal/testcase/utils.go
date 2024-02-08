package testcase

import (
	"encoding/json"
	"fmt"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"strings"
)

func (r *Reconciler) ProcessTestCase(objectTemplate *tofaniov1alpha1.ObjectTemplate, testCase *tofaniov1alpha1.TestCase) error {
	for _, field := range testCase.Spec.DynamicFields {
		for _, value := range field.Values {
			// Apply each value to the template independently
			modifiedTemplate, err := r.ApplyValueToTemplate(objectTemplate, field.Path, value)
			if err != nil {
				r.Log.Error(err, "Failed to apply dynamic field value to template", "Path", field.Path, "Value", value)
				continue // Move to the next value or field on error
			}

			r.Log.Info("Successfully applied value to template", "Path", field.Path, "Value", value, "ModifiedTemplate", string(modifiedTemplate))

			// Here, you would typically create or update the resource based on the modified template
			// This involves converting the JSON back into a Kubernetes object and applying it
			// The specific implementation will depend on your use case and Kubernetes client library
		}
	}
	return nil
}

// ApplyValueToTemplate applies dynamic field value to the specified ObjectTemplate
func (r *Reconciler) ApplyValueToTemplate(objectTemplate *tofaniov1alpha1.ObjectTemplate, path string, value string) ([]byte, error) {
	var templateMap map[string]interface{}
	if err := json.Unmarshal(objectTemplate.Spec.Template.Raw, &templateMap); err != nil {
		r.Log.Error(err, "Failed to unmarshal ObjectTemplate into map")
		return nil, err
	}

	// Navigate and apply the value to the specified path
	if err := navigateAndApplyValue(&templateMap, path, value); err != nil {
		r.Log.Error(err, "Failed to apply value to path", "Path", path, "Value", value)
		return nil, err
	}

	// Re-serialize the modified map back to JSON
	modifiedTemplate, err := json.Marshal(templateMap)
	if err != nil {
		r.Log.Error(err, "Failed to marshal modified template back to JSON")
		return nil, err
	}

	return modifiedTemplate, nil
}

// navigateAndApplyValue navigates the templateMap based on the path and applies the value.
func navigateAndApplyValue(templateMap *map[string]interface{}, path string, value string) error {
	currentMap := *templateMap
	pathParts := strings.Split(path, ".")

	for i, part := range pathParts {
		if i == len(pathParts)-1 {
			currentMap[part] = value // Apply the value at the target path
			return nil
		} else {
			if nextMap, ok := currentMap[part].(map[string]interface{}); ok {
				currentMap = nextMap
			} else {
				// The next part of the path does not exist or is not a map, so we need to create it
				newMap := make(map[string]interface{})
				currentMap[part] = newMap
				currentMap = newMap
			}
		}
	}

	return fmt.Errorf("path not found: %s", path)
}
