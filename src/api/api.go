package api

import (
  "fmt"
  "net/http"

  "../index"
  "strconv"
  "strings"
)

func HandleFunc(schema *index.Schema, search *index.Search, pathPrefix string) (func(http.ResponseWriter, *http.Request)) {
  return func(w http.ResponseWriter, r *http.Request) {
    relativePath := r.URL.Path[len(pathPrefix) - 1:]
    if relativePath == "/schema.json" { // schema
      SendJSONResponse(w, schema)
      return
    } else if relativePath == "/searchMeta.json" { // search schema
      SendJSONResponse(w, search)
      return
    } else if relativePath == "/search.json" { // search
      SendJSONResponse(w, search.Search(r.URL.Query(), schema))
      return
    } else if strings.HasSuffix(relativePath, ".json") && len(relativePath) > 6 { // read operation
      if index, err := strconv.Atoi(relativePath[1 : len(relativePath) - 5]); err == nil && index < schema.TotalItems && index >= 0 {
        SendJSONResponse(w, schema.GetItem(index))
        return
      }
    } else {
      fmt.Println("UGLY REQUEST", r.URL.Path)
    }
    SendJSONResponse(w, nil)
    return
  }
}

func Serve(schema *index.Schema, search *index.Search, address string, pathPrefix string) int {
  http.HandleFunc(pathPrefix, HandleFunc(schema, search, pathPrefix))

  fmt.Println("Listening to", address)
  if http.ListenAndServe(address, nil) == nil {
    return 0
  }
  fmt.Println("FAILED")
  return 1
}
