package repository

import (
	"fmt"
	"reflect"

	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func ParseCypherQueryResult(record neo4j.Record, alias string, target interface{}) error {
	elem := reflect.ValueOf(target).Elem()

	for i := 0; i < elem.Type().NumField(); i++ {
		structField := elem.Type().Field(i)

		tag := structField.Tag.Get("json") // Reading from the "json" tag
		fieldType := structField.Type
		fieldName := structField.Name

		if val, ok := record.Get(fmt.Sprintf("%s.%s", alias, tag)); ok && val != nil {
			field := elem.FieldByName(fieldName)
			if field.IsValid() && field.CanSet() {
				switch fieldType.Kind() {
				case reflect.String:
					if strVal, ok := val.(string); ok {
						field.SetString(strVal)
					}
				case reflect.Int, reflect.Int64:
					if intVal, ok := val.(int64); ok {
						field.SetInt(intVal)
					}
				case reflect.Bool:
					if boolVal, ok := val.(bool); ok {
						field.SetBool(boolVal)
					}
				case reflect.Ptr:
					switch fieldType.Elem().Kind() {
					case reflect.String:
						if strVal, ok := val.(*string); ok {
							field.Set(reflect.ValueOf(strVal))
						}
					case reflect.Int, reflect.Int64:
						if intVal, ok := val.(*int64); ok {
							field.Set(reflect.ValueOf(intVal))
						}
					case reflect.Bool:
						if boolVal, ok := val.(*bool); ok {
							field.Set(reflect.ValueOf(boolVal))
						}
					}
				case reflect.Slice:
					// Handle slice types
				default:
					return fmt.Errorf("unsupported type: %s", fieldType.Kind().String())
				}
			}
		}
	}

	return nil
}

// BoolPtr returns pointer to a boolean
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns pointer to a string
func StringPtr(str string) *string {
	return &str
}

// IntPtr returns pointer to an Int
func IntPtr(i int) *int {
	return &i
}

// PtrOrPtrEmptyString returns pointer to an Int
func PtrOrPtrEmptyString(ptr *string) *string {
	if ptr == nil {
		return StringPtr("")
	}
	return ptr
}

// Float64Ptr returns a pointer to the float64 value passed in.
func Float64Ptr(v float64) *float64 {
	return &v
}

func GetMetadataValue(meta map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if value, exists := meta[key]; exists {
		return value
	}
	return defaultValue
}
