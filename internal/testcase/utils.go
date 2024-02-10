package testcase

import (
	"context"
	"encoding/json"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	"github.com/invioteq/tofan/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func (r *Reconciler) ProcessTestCase(ctx context.Context, objectTemplate *tofaniov1alpha1.ObjectTemplate, testCase *tofaniov1alpha1.TestCase) error {
	for _, field := range testCase.Spec.DynamicFields {
		// `field.Values` is now a map, so iterate through the map
		for key, jsonValue := range field.Values {
			// Convert extv1.JSON to raw JSON bytes
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

			// Here, you would typically create or update the resource based on the modified template
			// This involves converting the JSON back into a Kubernetes object and applying it
			err = r.ApplyObjectToCluster(ctx, modifiedTemplate)
			if err != nil {
				// Handle the error appropriately
				r.Log.Error(err, "Failed to apply object to cluster", "ModifiedTemplate", modifiedTemplate)
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

func (r *Reconciler) ApplyObjectToCluster(ctx context.Context, objJSON []byte) error {
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
	if unstrObj.GetName() == "" {
		// Generate a unique name if not provided, for example:
		unstrObj.SetName("testcase-" + utils.GenerateRandomString(5))
	}

	// Use the controller-runtime client to apply the object to the cluster
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
