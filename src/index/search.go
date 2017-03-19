package index

import (
  "fmt"
  "net/url"
  "regexp"
  "runtime"
  "strconv"
  "strings"
)

const RESULTS_TO_RETURN int = 20

const ENUMERATE_THRESHOLD_FRACTION float64 = 0.01
const ENUMERATE_THRESHOLD_COUNT float64 = 100

type Search struct {
  Fields map[string]SearchField `json:"fields"`
  Sort []string `json:"sort,omitempty"`
}

type SearchField struct {
  Entropy float64 `json:"entropy"`
  Filters []string `json:"filters,omitempty"`
  OutValues interface{} `json:"enum,omitempty"`
  MinValue float64 `json:"minValue,omitempty"`
  MaxValue float64 `json:"maxValue,omitempty"`
}


func (search * Search) Initialise (schema * Schema) {
  search.Fields = make(map[string]SearchField);
  
  for _, property := range schema.SummaryFields[:5] {
    propertyData := schema.Properties[property]
    
    fmt.Print("Bootstrapping search... ", property, " " , propertyData.Type);
    
    switch propertyData.Type {
      case "boolean":
        IndexBooleanField(property, &propertyData, search);
        break;
      case "number":
        IndexNumberField(property, &propertyData, search);
        break;
      case "string":
        IndexStringField(property, &propertyData, search);
        break;
    };
    
    schema.Properties[property] = propertyData
    
    fmt.Println(";");
  }
  
  IndexFullText(schema, search);
  
  runtime.GC();
  
  fmt.Println("Bootstrapping search... done.")
}


func IndexBooleanField(field string, fieldData * SchemaField, search *Search) {
  search.Fields[field] = SearchField{Filters: []string{"equals"}, OutValues: []string{"true", "false"}, Entropy: fieldData.Entropy}
}

func IndexNumberField(field string, fieldData * SchemaField, search *Search) {
  search.Fields[field] = SearchField{Filters: []string{"equals", "lessThan", "greaterThan", "notEquals"}, MinValue: fieldData.MinValue, MaxValue: fieldData.MaxValue, Entropy: fieldData.Entropy}
}

func IndexStringField(field string, fieldData * SchemaField, search *Search) {
  Filters := []string{};
  
  var OutValues interface{}
  
  UniqueValuesCount := float64(len(fieldData.UniqueValues))
  UniqueValuesFraction := UniqueValuesCount / float64(len(fieldData.ValueIndex))
  
  if UniqueValuesFraction <= ENUMERATE_THRESHOLD_FRACTION && UniqueValuesCount <= ENUMERATE_THRESHOLD_COUNT {
    fieldData.OutValues = fieldData.UniqueValues
    OutValues = fieldData.UniqueValues
    Filters = append(Filters, "within")
  } else {
    Filters = append(Filters, "regex")
  }
  
  search.Fields[field] = SearchField{Filters: Filters, OutValues: OutValues, Entropy: fieldData.Entropy}
}

func IndexFullText(schema * Schema, search *Search) {
  
  fields := make(map[string]SchemaField, 5)
  
  for field, fieldData := range schema.Properties {
    if fieldData.Type == "string" && fieldData.HasSpace && !fieldData.AllUnique {
      UniqueValuesCount := float64(len(fieldData.UniqueValues))
      UniqueValuesFraction := UniqueValuesCount / float64(len(fieldData.ValueIndex))
      if UniqueValuesFraction > ENUMERATE_THRESHOLD_FRACTION && UniqueValuesCount > ENUMERATE_THRESHOLD_COUNT {
        fields[field] = fieldData;
      }
    }
  }
  
  if len(fields) > 0 {
    fmt.Print("Bootstrapping search... ", "search fulltext:");
    stemmed := make([]string, schema.TotalItems)
    for field, fieldData := range fields {
      getValue := StringAccessor(fieldData)
      fmt.Print(" ", field);
      for i := 0; i < schema.TotalItems; i++ {
        if value, has := getValue(i); has {
          stemmed[i] += PadWithSpace(LexAndStem(value))
        }
      }
      fmt.Print(",");
    }
    schema.TextIndex = Strip(stemmed)
    search.Fields["search"] = SearchField{Filters: []string{"search"}, Entropy: 0}
    fmt.Println(";")
  }
}

func (search * Search) Search(queries url.Values, schema * Schema) map[string]interface{} {
  
  results := SearchStart(schema)
  
  errors := make([]string, 0, 10)
  
  for _, query := range QueryByEntropy(queries, search) {
    queryValues := queries[query]
    field, has := schema.Properties[query]
    if !has && query == "search" {
      field = SchemaField{Type: "string"}
      has = true
    }
    if has {
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
                results = SearchStringSearch(StringIndexAccessor(schema), results, value)
                break;
              default:
                errors = append(errors, "field '" + query + "' filter '" + filter + "' value '" + value + "' is not supported")
            }
          }
          break
      }
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

func QueryByEntropy(queries url.Values, search * Search) []string {
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
      a := search.Fields[qq[i - 1]].Entropy
      b := search.Fields[qq[i]].Entropy
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
  uniqueValues := field.UniqueValues
  valueIndex := field.ValueIndex
  return func(x int) bool {
    return uniqueValues[valueIndex[x]].(bool)
  }
}

func NumberAccessor (field SchemaField) func(int) (float64, bool) {
  uniqueValues := field.UniqueValues
  valueIndex := field.ValueIndex
  return func(x int) (float64, bool) {
    xx := valueIndex[x]
    if xx != -1 {
      return uniqueValues[xx].(float64), true
    } else {
      return 0.0, false
    }
  }
}

func StringAccessor (field SchemaField) func(int) (string, bool) {
  uniqueValues := field.UniqueValues
  valueIndex := field.ValueIndex
  return func(x int) (string, bool) {
    xx := valueIndex[x]
    if xx != -1 {
      return uniqueValues[xx].(string), true
    } else {
      return EMPTY, false
    }
  }
}

func StringIndexAccessor (schema *Schema) func(int) (string, bool) {
  data := schema.TextIndex
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
