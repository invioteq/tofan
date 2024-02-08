package testcase

import (
	"context"
	"encoding/json"
	"fmt"
	tofaniov1alpha1 "github.com/invioteq/tofan/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

func (r *Reconciler) ProcessTestCase(ctx context.Context, objectTemplate *tofaniov1alpha1.ObjectTemplate, testCase *tofaniov1alpha1.TestCase) error {
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
			err = r.ApplyObjectToCluster(ctx, modifiedTemplate)
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

	// Generate a unique name for the object if it has a metadata.name field
	if metadata, ok := templateMap["metadata"].(map[string]interface{}); ok {
		if _, ok := metadata["name"].(string); ok {
			uniqueName := metadata["name"].(string) + "-" + generateRandomString(5)
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
		unstrObj.SetName("example-name-" + generateRandomString(5))
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

// generateRandomString creates a random string of length n using rand.Source
func generateRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	source := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	r := rand.New(source) // Create a new rand.Rand with the given source
	for i := range b {
		b[i] = letterBytes[r.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
