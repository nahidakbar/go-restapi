package index

import (
  "fmt"
  "math"
  "reflect"
  "sort"
  "strings"
  "encoding/json"
)

const EMPTY string = ""

func DetermineFieldType(data * []interface{}) string {
  for _, value := range *data {
    if value != nil {
      switch reflect.TypeOf(value).Kind().String() {
        case "slice":
          return "array"
        case "map":
          return "object"
        case "float64":
          return"number"
        case "string":
          return "string"
        case "bool":
          return "boolean"
      }
    }
  }
  return "null"
}

func InitialiseBooleanField(field string, fieldData []interface {}, schema *Schema) {
  
  UniqueValues := []interface{}{false, true}
  ValueIndex := make([]int, len(fieldData))
  
  trueValues := 0.0
  falseValues := 0.0
  for index, value := range fieldData {
    if value == nil {
      ValueIndex[index] = 0
    } else {
      if value.(bool) {
        ValueIndex[index] = 0
      } else {
        ValueIndex[index] = 1
      }
    }
    if ValueIndex[index] == 1 {
      trueValues += 1.0
    } else {
      falseValues += 1.0
    }
  }
  Entropy := 0.0
  Total := float64(len(ValueIndex))
  
  Entropy -= trueValues / Total * math.Log2(trueValues / Total)
  Entropy -= falseValues / Total * math.Log2(falseValues / Total)
  
  schema.Properties[field] = SchemaField{Type: "boolean", Entropy: Entropy, UniqueValues: UniqueValues, ValueIndex: ValueIndex}
  schema.SummaryFields = append(schema.SummaryFields, field)
}

func InitialiseNumberField(field string, fieldData []interface {}, schema *Schema) {
  
  UniqueValues := make(Float64Set, 0, 2)
  for _, value := range fieldData {
    if value != nil {
      UniqueValues.AddToSet(value.(float64))
    }
  }
  sort.Sort(sort.Float64Slice(UniqueValues))
  
  minValue := UniqueValues[0]
  maxValue := UniqueValues[len(UniqueValues) - 1]
  
  ValueIndex := make([]int, len(fieldData))
  
  for index, value := range fieldData {
    if value == nil {
      ValueIndex[index] = -1
    } else {
      ValueIndex[index] = UniqueValues.IndexOf(value.(float64))
    }
  }
  
  schema.Properties[field] = SchemaField{Type: "number", UniqueValues: UniqueValues.Trim(), ValueIndex: ValueIndex, MinValue: minValue, MaxValue: maxValue}
}

func InitialiseStringField(field string, fieldData []interface {}, schema *Schema) {
  //output := make([]string, len(fieldData))

  totalValues := 0.0
  uniqueValues := make(map[string]float64)
  hasSpace := false
  allUnique := true

  for _, value := range fieldData {
    if value != nil && value.(string) != EMPTY {
      uniqueValues[value.(string)] = uniqueValues[EMPTY] + 1
      totalValues++
    }
  }
  
  fmt.Print(".")
  
  UniqueValues := make(StringSet, len(uniqueValues))
  
  Entropy := 0.0
  ValueCount := float64(len(uniqueValues))
  for value, count := range uniqueValues {
    UniqueValues.AddToSet(value)
    
    if !hasSpace && strings.Contains(value, " ") {
      hasSpace = true;
    }
    
    if ValueCount > 1 {
      allUnique = false
      Entropy -= count / totalValues * math.Log2(count / totalValues) / math.Log2(float64(ValueCount))
    } else {
      ValueCount = float64(len(fieldData))
      Entropy -= count / ValueCount * math.Log2(count / ValueCount)
      Entropy -= (ValueCount - count) / ValueCount * math.Log2((ValueCount - count) / ValueCount)
    }
  }
  
  fmt.Print(".")
  
  UniqueValues.Sort()
  
  ValueIndex := make([]int, len(fieldData))
  
  for index, value := range fieldData {
    if value == nil || value.(string) == EMPTY {
      ValueIndex[index] = -1
    } else {
      ValueIndex[index] = UniqueValues.IndexOf(value.(string))
    }
  }
  
  schema.Properties[field] = SchemaField{Type: "string", UniqueValues: UniqueValues.Trim(), ValueIndex: ValueIndex, Entropy: Entropy, HasSpace: hasSpace, AllUnique: allUnique}
  schema.SummaryFields = append(schema.SummaryFields, field)
  
  fmt.Print(".")
}

func ToJson(item interface{}) string {
  enc, err := json.Marshal(item)
  if err != nil {
    fmt.Println("JSON ENCODING ERR", err)
  }
  return string(enc)
}

func FromJson(item string) interface{} {
  var out interface{}
  err := json.Unmarshal([]byte(item) , out)
  if err != nil {
    fmt.Println("JSON DECODING ERR", err)
  }
  return out
}

func InitialiseMiscField(field string, fieldType string, fieldData []interface {}, schema *Schema) {
  
  uniqueValues := make(StringSet, 100)
  
  ValueIndex := make([]int, len(fieldData))
  
  for index, value := range fieldData {
    if value != nil {
      ValueIndex[index] = uniqueValues.AddToSet(ToJson(value))
    } else {
      ValueIndex[index] = -1
    }
  }
  
  UniqueValues := make([]interface{}, len(uniqueValues))
  for i, v := range uniqueValues.Trim() {
    UniqueValues[i] = FromJson(v.(string))
  }
  
  schema.Properties[field] = SchemaField{Type: fieldType, Entropy: 0, UniqueValues: UniqueValues, ValueIndex: ValueIndex}
}
