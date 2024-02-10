package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// NavigateAndApplyValue navigates the templateMap based on the path and applies the value.
func NavigateAndApplyValue(templateMap *map[string]interface{}, path string, value interface{}) error {
	currentMap := *templateMap
	pathParts := strings.Split(path, ".")

	for i, part := range pathParts {
		if i == len(pathParts)-1 {
			// Apply the value at the target path using reflection to handle various types
			switch v := value.(type) {
			case int, int32, int64, float32, float64, string, bool:
				currentMap[part] = v
			case []interface{}:
				// Handle slice of interfaces directly
				currentMap[part] = v
			case map[string]interface{}:
				// Handle map directly
				currentMap[part] = v
			default:
				// For types not explicitly handled above, use reflection
				rv := reflect.ValueOf(v)
				switch rv.Kind() {
				case reflect.Slice, reflect.Array:
					var slice []interface{}
					for i := 0; i < rv.Len(); i++ {
						slice = append(slice, rv.Index(i).Interface())
					}
					currentMap[part] = slice
				case reflect.Map:
					// Ensure map keys are strings, as required by JSON and Kubernetes objects
					mapValue := make(map[string]interface{})
					for _, key := range rv.MapKeys() {
						strKey, ok := key.Interface().(string)
						if !ok {
							return fmt.Errorf("map key is not a string: %v", key)
						}
						mapValue[strKey] = rv.MapIndex(key).Interface()
					}
					currentMap[part] = mapValue
				default:
					// Attempt to handle as a generic interface, which might not be directly marshallable
					jsonVal, err := json.Marshal(v)
					if err != nil {
						return fmt.Errorf("failed to marshal unsupported type for path '%s': %v", path, err)
					}
					var genericVal interface{}
					if err := json.Unmarshal(jsonVal, &genericVal); err != nil {
						return fmt.Errorf("failed to unmarshal unsupported type for path '%s': %v", path, err)
					}
					currentMap[part] = genericVal
				}
			}
			return nil
		} else {
			// Navigate or create the next part of the path
			if nextMap, ok := currentMap[part].(map[string]interface{}); ok {
				currentMap = nextMap
			} else {
				newMap := make(map[string]interface{})
				currentMap[part] = newMap
				currentMap = newMap
			}
		}
	}

	return fmt.Errorf("path not found: %s", path)
}

// GenerateRandomString creates a random string of length n using rand.Source
func GenerateRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz0123456789"
	source := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	r := rand.New(source) // Create a new rand.Rand with the given source
	for i := range b {
		b[i] = letterBytes[r.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
