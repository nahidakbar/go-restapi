package api

import (
  "fmt"
  "net/http"
  "strconv"
  "strings"
)

func HandleFunc(collection Collection, pathPrefix string) (func(http.ResponseWriter, *http.Request)) {
  return func(w http.ResponseWriter, r *http.Request) {
    relativePath := r.URL.Path[len(pathPrefix) - 1:]
    if r.Method == "GET" {
      if relativePath == "/schema.json" { // schema
        SendJSONResponse(w, collection.Schema())
        return
      } else if relativePath == "/searchMeta.json" { // search schema
        SendJSONResponse(w, collection.SearchMeta())
        return
      } else if relativePath == "/search.json" { // search
        SendJSONResponse(w, collection.Search(r.URL.Query()))
        return
      } else if strings.HasSuffix(relativePath, ".json") && len(relativePath) > 6 { // read operation
        if index, err := strconv.Atoi(relativePath[1 : len(relativePath) - 5]); err == nil && index < collection.TotalItems() && index >= 0 {
          SendJSONResponse(w, collection.GetItem(index))
          return
        }
      }
    }
    SendJSONResponse(w, nil)
    return
  }
}

func Serve(collection Collection, address string, pathPrefix string) int {
  http.HandleFunc(pathPrefix, HandleFunc(collection, pathPrefix))

  fmt.Println("Listening to", address)
  if http.ListenAndServe(address, nil) == nil {
    return 0
  }
  fmt.Println("FAILED")
  return 1
}
