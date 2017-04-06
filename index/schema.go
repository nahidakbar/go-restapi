package index

import (
  "fmt"
  "runtime"
  "sort"
  "sync"
)

type Schema struct {
  Type string `json:"type,omitempty"`
  Properties map[string]SchemaField `json:"properties,omitempty"`
  TotalItems int `json:"-"`
  SummaryFields []string `json:"-"`
}

type SchemaField struct {
  Type string `json:"type"`
  
  // internal; use less memory
  Entropy float64 `json:"entropy"`
  UniqueValues []interface{} `json:"-"`
  ValueIndex []int `json:"-"`
  
  // number
  MinValue float64 `json:"minValue,omitempty"`
  MaxValue float64 `json:"maxValue,omitempty"`
  
  // string
  HasSpace bool `json:"-"`
  AllUnique bool `json:"-"`
  OutValues interface{} `json:"enum,omitempty"`
}

func (s Schema) Len() int {
  return len(s.SummaryFields)
}

func (s Schema) Swap(i, j int) {
  s.SummaryFields[i], s.SummaryFields[j] = s.SummaryFields[j], s.SummaryFields[i]
}

func (s Schema) Less(i, j int) bool {
  return s.Properties[s.SummaryFields[i]].Entropy < s.Properties[s.SummaryFields[j]].Entropy
}

func (schema *Schema) Initialise (data map[string][]interface{}) {
  schema.Type = "object";
  
  schema.Properties = make(map[string]SchemaField);
  
  var wait sync.WaitGroup
  
  for field, fieldData := range data {
    wait.Add(1)
    go func(field string, fieldData []interface{}) {
      defer wait.Done()
      fmt.Print("Bootstrapping schema... ", field);
      schema.TotalItems = len(fieldData)
      fieldType := DetermineFieldType(&fieldData)
      if fieldType != "null" {
        switch fieldType {
          case "boolean":
            InitialiseBooleanField(field, fieldData, schema);
            break;
          case "number":
            InitialiseNumberField(field, fieldData, schema);
            break;
          case "string":
            InitialiseStringField(field, fieldData, schema);
            break;
          default:
            InitialiseMiscField(field, fieldType, fieldData, schema);
        };
      }
      runtime.GC();
      fmt.Println(";");
    }(field, fieldData)
    data[field] = nil
  }
  
  wait.Wait()
  
  sort.Reverse(schema)
  fmt.Println("Bootstrapping schema... done.")
}

func (schema * Schema) GetSummary(index int, fields int) map[string]interface{} {
  if fields + 1 > len(schema.SummaryFields) {
    fields = len(schema.SummaryFields) - 1
  }
  output := make(map[string]interface{}, len(schema.Properties))
  for _, field := range schema.SummaryFields[:fields + 1] {
    if fieldValues, has := schema.Properties[field]; has {
      valueIndex := fieldValues.ValueIndex[index]
      if valueIndex != -1 {
        output[field] = fieldValues.UniqueValues[valueIndex]
      }
    }
  }
  return output
}

func (schema * Schema) GetItem(index int) map[string]interface{} {
  output := make(map[string]interface{}, len(schema.Properties))
  for field, fieldValues := range schema.Properties {
    valueIndex := fieldValues.ValueIndex[index]
    if valueIndex != -1 {
      output[field] = fieldValues.UniqueValues[valueIndex]
    }
  }
  return output
}
