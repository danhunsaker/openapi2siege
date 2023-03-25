package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func createFakePayload(schemaProxy *base.SchemaProxy) (interface{}, error) {
	var output interface{}

	for i := 0; i < 5; i++ {
		schemaOption, err := schemaProxy.BuildSchema()
		if err != nil {
			return nil, err
		}

		output, err = createFakePayloadFromSchema(schemaOption)
		if err == nil {
			break
		}

		if err.Error() == "try-rebuild" {
			log.Printf("Retry %d failed; trying again", i)
			continue
		}

		return nil, err
	}

	return output, nil
}

func createFakePayloadFromSchema(schema *base.Schema) (interface{}, error) {
	// schemaJson, err := json.Marshal(schema)
	// if err != nil {
	// 	return nil, fmt.Errorf("Couldn't JSON-encode schema %v\n\t%s", schema, err)
	// }
	// log.Printf("Processing schema: %s", schemaJson)

	if schema.Default != nil {
		return schema.Default, nil
	}

	if schema.Example != nil {
		return schema.Example, nil
	}

	for _, example := range schema.Examples {
		return example, nil
	}

	if len(schema.OneOf) > 0 {
		return createFakePayload(schema.OneOf[0])
	}

	if len(schema.AnyOf) > 0 {
		return createFakePayload(schema.AnyOf[0])
	}

	payload := make(map[string]interface{})
	for _, schemaProxy := range schema.AllOf {
		partialPayload, err := createFakePayload(schemaProxy)
		if err != nil {
			return nil, err
		}

		partialMap, isType := partialPayload.(map[string]interface{})
		if !isType {
			return nil, fmt.Errorf("Couldn't generate payload; expected `%T`, but got `%T`", payload, partialPayload)
		}

		for key, val := range partialMap {
			payload[key] = val
		}
	}

	if len(payload) > 0 {
		return payload, nil
	}

	for _, schemaType := range schema.Type {
		switch schemaType {
		case "object":
			return iterateObject(schema)
		case "array":
			return iterateArray(schema)
		case "null":
			return nil, nil
		case "boolean":
			return true, nil
		case "number":
			switch schema.Format {
			case "float":
				return float32(0), nil
			case "double":
				return float64(0), nil
			default:
				return float64(0), nil
			}
		case "integer":
			switch schema.Format {
			case "int32":
				return int32(0), nil
			case "int64":
				return int64(0), nil
			default:
				return int(0), nil
			}
		case "string":
			return "test", nil
		default:
			return nil, fmt.Errorf("Unknown type %s; can't generate a fake value for something we don't understand\n\tPlease let us know you'd like it added!", schemaType)
		}
	}

	switch {
	case len(schema.Properties) > 0:
		return iterateObject(schema)
	case schema.Items != nil:
		return iterateArray(schema)
	default:
		return nil, fmt.Errorf("try-rebuild")
	}
}

func getPayloadFromType(mediatype string, data interface{}) (string, error) {
	switch mediatype {
	case "application/json":
		// don't serialize a string that's already serialized
		if stringData, isType := data.(string); isType && (strings.Contains(stringData, "{") || strings.Contains(stringData, "[")) {
			return stringData, nil
		}

		bytes, err := json.Marshal(data)
		// log.Printf("marshalled %T value to %s\n", data, bytes)
		return string(bytes), err
	default:
		return "", fmt.Errorf("Unsupported mediatype %s\n\tAsk us to add it!", mediatype)
	}
}

func iterateObject(schema *base.Schema) (interface{}, error) {
	output := make(map[string]interface{})

	for k, v := range schema.Properties {
		payload, err := createFakePayload(v)
		if err != nil {
			return nil, err
		}

		output[k] = payload
	}

	return output, nil
}

func iterateArray(schema *base.Schema) (interface{}, error) {
	if schema.Items.IsB() {
		if schema.Items.B {
			return []interface{}{}, nil
		}

		return nil, fmt.Errorf("Can't determine how to generate a value for %s (whose definition in the spec is `false`)\n", schema.Title)
	}

	payload, err := createFakePayload(schema.Items.A)
	if err != nil {
		return nil, err
	}

	return []interface{}{payload}, nil
}
