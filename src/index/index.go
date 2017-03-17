package index

import (
  "github.com/reiver/go-porterstemmer"
  "fmt"
  "math"
  "net/url"
  "reflect"
  "regexp"
  "sort"
  "strconv"
  "strings"
)

const RESULTS_TO_RETURN int = 20

type Schema struct {
  Type string `json:"type,omitempty"`
  Properties map[string]SchemaField `json:"properties,omitempty"`
  TotalItems int `json:"-"`
  SummaryFields []string `json:"-"`
}

type SchemaField struct {
  Type string `json:"type"`
  
  // internal
  Entropy float64 `json:"entropy"`
  Values interface{} `json:"-"`
  
  // number
  MinValue float64 `json:"minValue,omitempty"`
  MaxValue float64 `json:"maxValue,omitempty"`
  
  // string
  OutValues interface{} `json:"enum,omitempty"`
  StringStemmed []string `json:"-"`
}

type Search struct {
  Fields map[string]SearchField `json:"fields"`
  Sort []string `json:"sort,omitempty"`
}

type SearchField struct {
  Filters []string `json:"filters,omitempty"`
  Values []string `json:"enum,omitempty"`
  MinValue float64 `json:"minValue,omitempty"`
  MaxValue float64 `json:"maxValue,omitempty"`
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

func Index (data map[string][]interface{}) (*Schema, *Search) {
  schema := new(Schema);
  schema.Type = "object";
  schema.Properties = make(map[string]SchemaField);
  
  search := new(Search);
  search.Fields = make(map[string]SearchField);
  
  go func(){
  
    fmt.Print("Preparing data ...")
  
    for field, fieldData := range data {
      
      fmt.Print(" ", field);
      
      schema.TotalItems = len(fieldData)
      fieldType := determineFieldType(&fieldData)
      if fieldType != "null" {
        switch fieldType {
          case "boolean":
            indexBooleanField(field, fieldData, schema, search);
            break;
          case "number":
            indexNumberField(field, fieldData, schema, search);
            break;
          case "string":
            indexStringField(field, fieldData, schema, search);
            break;
          case "array", "object":
            indexMiscField(field, fieldType, fieldData, schema, search);
            break;
          default:
            fmt.Print("SKIP [", field, fieldType, "]");
        };
      }
      fmt.Print("; ");
    }
    
    sort.Reverse(schema)
    
    fmt.Println(" and done.")
  
  }()
  
  return schema, search
}

func determineFieldType(data * []interface{}) string {
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

func indexBooleanField(field string, fieldData []interface {}, schema *Schema, search *Search) {
  
  output := make([]bool, len(fieldData))
  
  trueValues := 0.0
  falseValues := 0.0
  for index, value := range fieldData {
    if value == nil {
      output[index] = false
    } else {
      output[index] = value.(bool)
    }
    if output[index] {
      trueValues += 1.0
    } else {
      falseValues += 1.0
    }
  }
  Entropy := 0.0
  Total := float64(len(output))
  
  Entropy -= trueValues / Total * math.Log2(trueValues / Total)
  Entropy -= falseValues / Total * math.Log2(falseValues / Total)
  
  schema.Properties[field] = SchemaField{Type: "boolean", Entropy: Entropy, Values: output}
  
  search.Fields[field] = SearchField{Filters: []string{"equals"}, Values: []string{"true", "false"}}
  
  schema.SummaryFields = append(schema.SummaryFields, field)
}

func indexNumberField(field string, fieldData []interface {}, schema *Schema, search *Search) {
  output := make([]float64, len(fieldData))
  
  minValue := math.Inf(1)
  maxValue := math.Inf(-1)
  
  for index, value := range fieldData {
    if value == nil {
      output[index] = math.NaN()
    } else {
      output[index] = value.(float64)
      if output[index] > maxValue {
        maxValue = output[index]
      }
      if output[index] < minValue {
        minValue = output[index]
      }
    }
  }
  
  schema.Properties[field] = SchemaField{Type: "number", Values: output, MinValue: minValue, MaxValue: maxValue}
  search.Fields[field] = SearchField{Filters: []string{"equals", "lessThan", "greaterThan", "notEquals"}, MinValue: minValue, MaxValue: maxValue}
}


func indexStringField(field string, fieldData []interface {}, schema *Schema, search *Search) {
  empty := ""
  output := make([]string, len(fieldData))
  
  totalValues := 0.0
  uniqueValues := make(map[string]float64)
  hasSpace := false
  
  for index, value := range fieldData {
    if value == nil {
      output[index] = empty
    } else {
      output[index] = value.(string)
      totalValues++
      if !hasSpace && strings.Contains(output[index], " ") {
        hasSpace = true
      }
      existing, _ := uniqueValues[output[index]]
      uniqueValues[output[index]] = existing + 1
    }
  }
  
  Entropy := 0.0
  ValueCount := float64(len(uniqueValues))

  for _, count := range uniqueValues {
    if ValueCount > 1 {
      Entropy -= count / totalValues * math.Log2(count / totalValues) / math.Log2(float64(ValueCount))
    } else {
      ValueCount = float64(len(fieldData))
      Entropy -= count / ValueCount * math.Log2(count / ValueCount)
      Entropy -= (ValueCount - count) / ValueCount * math.Log2((ValueCount - count) / ValueCount)
    }
  }

  Filters := []string{};
  
  var OutValues []string = nil
  var StringStemmed []string = nil

  if ValueCount < 500 {
    OutValues = make([]string, 0, len(uniqueValues))
    for value, _ := range uniqueValues {
      OutValues = append(OutValues, value)
    }
    Filters = append(Filters, "within");
  } else if hasSpace {
    fmt.Print(" indexing");
    StringStemmed = make([]string, len(output))
    for index, item := range output {
      StringStemmed[index] = PadWithSpace(LexAndStem(item));
    }
    Filters = append(Filters, "search", "regex")
  } else {
    Filters = append(Filters, "regex")
  }
  
  schema.Properties[field] = SchemaField{Type: "string", Values: output, Entropy: Entropy, OutValues: OutValues, StringStemmed: StringStemmed}
  search.Fields[field] = SearchField{Filters: Filters, Values: OutValues}
  schema.SummaryFields = append(schema.SummaryFields, field)
}

func indexMiscField(field string, fieldType string, fieldData []interface {}, schema *Schema, search *Search) {
  schema.Properties[field] = SchemaField{Type: fieldType, Entropy: 0, Values: fieldData}
}

func PadWithSpace(parts []string) string {
  return " " + strings.Join(parts, " ") + " "
}

func CharClass(item rune) rune {
  if item >= 'a' && item <= 'z' {
    return 'a'
  } else if item >= '0' && item <= '9' {
    return '0'
  } else {
    return item
  }
}

func LexAndStem(str string) []string {
  tokens := make([]string, 0, len(str))
  // hack togeather a barely legal lexer
  last := '\x00'
  lastTokensIndex := -1
  for _, item := range strings.ToLower(str) {
    current := CharClass(item)
    if current != ' ' {
      if current == last {
        tokens[lastTokensIndex] = tokens[lastTokensIndex] + string(item)
      } else {
        tokens = append(tokens, string(item))
        lastTokensIndex++
      }
    }
    last = current
  }
  // stem
  for x, xx := range tokens {
    tokens[x] = porterstemmer.StemString(xx)
  }
  return tokens
}

func (schema * Schema) GetSummary(index int, fields int) map[string]interface{} {
  
  if fields + 1 > len(schema.SummaryFields) {
    fields = len(schema.SummaryFields) - 1
  }
  
  output := make(map[string]interface{}, len(schema.Properties))
  for _, field := range schema.SummaryFields[:fields + 1] {
    fieldValues := schema.Properties[field]
    switch fieldValues.Type {
      case "string":
        value := fieldValues.Values.([]string)[index];
        if len(value) > 0 {
          output[field] = value
        }
        break
      case "number":
        value := fieldValues.Values.([]float64)[index];
        if !math.IsNaN(value) {
          output[field] = value
        }
        break
      case "boolean":
        value := fieldValues.Values.([]bool)[index];
        output[field] = value
        break
      case "array", "object":
        value := fieldValues.Values.([]interface{})[index];
        if value != nil {
          output[field] = value
        }
        break
    }
  }
  return output
}

func (schema * Schema) GetItem(index int) map[string]interface{} {
  output := make(map[string]interface{}, len(schema.Properties))
  for field, fieldValues := range schema.Properties {
    switch fieldValues.Type {
      case "string":
        value := fieldValues.Values.([]string)[index];
        if len(value) > 0 {
          output[field] = value
        }
        break
      case "number":
        value := fieldValues.Values.([]float64)[index];
        if !math.IsNaN(value) {
          output[field] = value
        }
        break
      case "boolean":
        value := fieldValues.Values.([]bool)[index];
        output[field] = value
        break
      case "array", "object":
        value := fieldValues.Values.([]interface{})[index];
        if value != nil {
          output[field] = value
        }
        break
    }
  }
  return output
}

func (search * Search) Search(queries url.Values, schema * Schema) map[string]interface{} {
  
  results := SearchStart(schema)
  
  errors := make([]string, 0, 10)
  
  for _, query := range QueryByEntropy(queries, schema, search) {
    queryValues := queries[query]
    field := schema.Properties[query]
    searchField := search.Fields[query]
    switch field.Type {
      case "boolean":
        for _, value := range queryValues {
          if i := strings.Index(value, ":"); i != -1 {
            value = value[i:]
          }
          switch value {
            case "true":
              results = SearchBoolean(BooleanAccessor(field), results, true)
              break;
            case "false":
              results = SearchBoolean(BooleanAccessor(field), results, false)
              break;
            default:
              errors = append(errors, "field '" + query + "' value '" + value + "' is not supported")
          }
        }
        break
      case "number":
        for _, value := range queryValues {
          filter := searchField.Filters[0]
          if i := strings.Index(value, ":"); i != -1 {
            filter = value[:i]
            value = value[i+1:]
          }
          val, _ := strconv.ParseFloat(value, 64)
          switch filter {
            case "equals":
              results = SearchNumberEqualTo(NumberAccessor(field), results, val)
              break;
            case "notEqual":
              results = SearchNumberNotEqualTo(NumberAccessor(field), results, val)
              break;
            case "lessThan":
              results = SearchNumberLessThan(NumberAccessor(field), results, val)
              break;
            case "greaterThan":
              results = SearchNumberGreaterThan(NumberAccessor(field), results, val)
              break;
            default:
             errors = append(errors, "field '" + query + "' filter '" + filter + "' value '" + value + "' is not supported")
          }
        }
        break
      case "string":
        for _, value := range queryValues {
          filter := searchField.Filters[0]
          if i := strings.Index(value, ":"); i != -1 {
            filter = value[:i]
            value = value[i+1:]
          }
          switch filter {
            case "within":
              results = SearchStringWithin(StringAccessor(field), results, value)
              break;
            case "regex":
              results = SearchStringRegex(StringAccessor(field), results, value)
              break;
            case "search":
              results = SearchStringSearch(StringIndexAccessor(field), results, value)
              break;
            default:
              errors = append(errors, "field '" + query + "' filter '" + filter + "' value '" + value + "' is not supported")
          }
        }
        break
    }
  }
  
  results_ := make([]int, 0, len(results))
  for i, x := range results {
    if x {
      results_ = append(results_, i)
    }
  }

  output := map[string]interface{}{
    "errors": errors,
    "total": len(results_),
    "limit": RESULTS_TO_RETURN,
  }

  if offset, has := queries["offset"]; has {
    off, _ := strconv.ParseInt(offset[0], 10, 64)
    results_ = results_[off:]
    output["offset"] = off
  } else {
    output["offset"] = 0
  }
  
  if len(results_) > RESULTS_TO_RETURN {
    results_ = results_[:RESULTS_TO_RETURN]
  }
  
  resultObjects := make([]map[string]interface{}, len(results_))
  for index, id := range results_ {
    item := schema.GetSummary(id, 9)
    item["id"] = id
    resultObjects[index] = item;
  }
  
  output["items"] = resultObjects
  
  return output
}

func QueryByEntropy(queries url.Values, schema * Schema, search * Search) []string {
  qq := make([]string, 0, len(queries))
  for query, _ := range queries {
    if field, hasField := search.Fields[query]; hasField && field.Filters != nil {
      qq = append(qq, query)
    }
  }
  sorted := false
  for !sorted {
    sorted = true
    for i := 1; i < len(qq); i++ {
      a := schema.Properties[qq[i - 1]].Entropy
      b := schema.Properties[qq[i]].Entropy
      if a > b {
        t := qq[i]
        qq[i] = qq[i - 1]
        qq[i - 1] = t
        sorted = false
      }
    }
  }
  return qq
}

func BooleanAccessor (field SchemaField) func(int) bool {
  data := field.Values.([]bool)
  return func(x int) bool {
    return data[x]
  }
}

func NumberAccessor (field SchemaField) func(int) (float64, bool) {
  data := field.Values.([]float64)
  return func(x int) (float64, bool) {
    xx := data[x]
    return xx, math.IsNaN(xx)
  }
}


func StringAccessor (field SchemaField) func(int) (string, bool) {
  data := field.Values.([]string)
  return func(x int) (string, bool) {
    xx := data[x]
    return xx, len(xx) > 0
  }
}

func StringIndexAccessor (field SchemaField) func(int) (string, bool) {
  data := field.StringStemmed
  return func(x int) (string, bool) {
    xx := data[x]
    return xx, len(xx) > 2
  }
}

// search methods

func SearchStart (schema * Schema) []bool {
  result := make([]bool, schema.TotalItems)
  for i, _ := range result {
    result[i] = true
  }
  return result
}

// boolean search

func SearchBoolean (accessor func(int) bool, results []bool, value bool) []bool {
  for i, x := range results {
    if x {
      xValue := accessor(i)
      if xValue != value {
        results[i] = false
      }
    }
  }
  return results
}

// numbers search

func SearchNumberLessThan(accessor func(int) (float64, bool), results []bool, value float64) []bool {
  for i, x := range results {
    if x {
      xValue, hasValue := accessor(i)
      if (!hasValue || xValue >= value) {
        results[i] = false
      }
    }
  }
  return results
}

func SearchNumberGreaterThan(accessor func(int) (float64, bool), results[]bool, value float64) []bool {
  for i, x := range results {
    if x {
      xValue, hasValue := accessor(i)
      if (!hasValue || xValue <= value) {
        results[i] = false
      }
    }
  }
  return results
}

func SearchNumberEqualTo(accessor func(int) (float64, bool), results[]bool, value float64) []bool {
  for i, x := range results {
    if x {
      xValue, hasValue := accessor(i)
      if (!hasValue || xValue != value) {
        results[i] = false
      }
    }
  }
  return results
}

func SearchNumberNotEqualTo(accessor func(int) (float64, bool), results[]bool, value float64) []bool {
  for i, x := range results {
    if x {
      xValue, hasValue := accessor(i)
      if (!hasValue || xValue == value) {
        results[i] = false
      }
    }
  }
  return results
}

// string search

func SearchStringWithin(accessor func(int) (string, bool), results []bool, value string) []bool {
  return SearchStringRegex(accessor, results, "^" + value + "$");
}

func SearchStringRegex(accessor func(int) (string, bool), results []bool, value string) []bool {
  re := regexp.MustCompile(value)
  for i, x := range results {
    if x {
      xValue, hasValue := accessor(i)
      if !hasValue || !re.MatchString(xValue) {
        results[i] = false
      }
    }
  }
  return results
}

func SearchStringSearch(accessor func(int) (string, bool), results []bool, value string) []bool {
  
  tokens := LexAndStem(value)
  
  present := make([]string, 0, len(tokens))
  direct := make([]string, 0, len(tokens))
  absent := make([]string, 0, len(tokens))
  
  for i := 0; i < len(tokens); i++ {
    class := CharClass(rune(tokens[i][0]))
    if class == '"' {
      i++
      slice := " "
      for tokens[i] != "\"" {
        slice += tokens[i] + " "
        present = append(present, tokens[i])
        i++
      }
      direct = append(direct, slice)
    } else if class == '-' {
      i++
      if i < len(tokens) {
        absent = append(absent, tokens[i])
      }
    } else if class == 'a' || class == '0' {
      present = append(present, tokens[i])
    }
  }
  
  for i, x := range results {
    if x {
      xValue, hasValue := accessor(i)
      
      if (hasValue) {
        
        valid := true
        
        if valid {
          for _, p := range present {
            if !strings.Contains(xValue, p) {
              valid = false
              break
            }
          }
        }
        
        if valid {
          for _, p := range direct {
            if !strings.Contains(xValue, p) {
              valid = false
              break
            }
          }
        }
        
        if valid {
          for _, p := range absent {
            if strings.Contains(xValue, p) {
              valid = false
              break
            }
          }
        }
        
        if !valid {
          hasValue = false
        }
      }
      
      if !hasValue {
        results[i] = false
      }
    }
  }
  return results
}
