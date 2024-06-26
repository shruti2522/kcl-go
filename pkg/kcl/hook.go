package kcl

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
	"kcl-lang.io/kcl-go/pkg/spec/gpyrpc"
)

var (
	DefaultHooks Hooks = []Hook{
		&typeAttributeHook{},
	}
)

type Hook interface {
	Do(o *Option, r *gpyrpc.ExecProgram_Result) error
}

type Hooks []Hook

type typeAttributeHook struct{}

func (t *typeAttributeHook) Do(o *Option, r *gpyrpc.ExecProgram_Result) error {
	// Deal the `_type` attribute
	if o != nil && r != nil && !o.fullTypePath && o.IncludeSchemaTypePath {
		return resultTypeAttributeHook(r)
	}
	return nil
}

func resultTypeAttributeHook(r *gpyrpc.ExecProgram_Result) error {
	// Modify the _type fields
	var data []map[string]interface{}
	var mapData map[string]interface{}
	// Unmarshal the JSON string into a Node
	if err := json.Unmarshal([]byte(r.JsonResult), &data); err == nil {
		modifyTypeList(data)
		marshal(r, data)
		return nil
	}

	if err := json.Unmarshal([]byte(r.JsonResult), &mapData); err == nil {
		modifyType(mapData)
		marshal(r, mapData)
		return nil
	}

	return nil
}

func marshal(r *gpyrpc.ExecProgram_Result, value interface{}) {
	// Marshal the modified Node back to YAML
	yamlOutput, _ := yaml.Marshal(value)
	// Marshal the modified Node back to JSON
	jsonOutput, _ := json.Marshal(&value)
	r.JsonResult = string(jsonOutput)
	r.YamlResult = string(yamlOutput)
}

func modifyTypeList(dataList []map[string]interface{}) {
	for _, data := range dataList {
		modifyType(data)
	}
}

func modifyType(data map[string]interface{}) {
	for key, value := range data {
		if key == "_type" {
			if v, ok := data[key].(string); ok {
				parts := strings.Split(v, ".")
				data[key] = parts[len(parts)-1]
			}
			continue
		}

		if nestedMap, ok := value.(map[string]interface{}); ok {
			modifyType(nestedMap)
		}
	}
}
