package main

import (
  "flag"
  "os"
  "./format"
  "./index"
  "./api"
)

func main() {
  address := ":8080"
  datafile := ""
  path := "/"
  flag.StringVar(&address, "address", address, "")
  flag.StringVar(&datafile, "datafile", datafile, "")
  flag.StringVar(&path, "path", path, "")
  flag.Parse()
  if len(datafile) > 0 {
    var search *index.Search = nil
    var schema *index.Schema = nil
    if data, err := format.Load(datafile); err == nil {
      schema, search = index.Index(data)
    }
    if search != nil && schema != nil {
      os.Exit(api.Serve(schema, search, address, path))
    } else {
      os.Exit(1)
    }
  }
  
  flag.PrintDefaults()
  os.Exit(2)
  
}
