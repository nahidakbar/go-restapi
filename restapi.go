package main

import (
  "flag"
  "os"
  "github.com/nahidakbar/go-restapi/input"
  "github.com/nahidakbar/go-restapi/index"
  "github.com/nahidakbar/go-restapi/api"
)

var address = flag.String("address", ":8080", "")
var datafile = flag.String("datafile", "", "")
var path = flag.String("path", "/", "")

func main() {
  flag.Parse()
  if len(*datafile) > 0 {
    if data, err := input.Load(*datafile); err == nil {
      os.Exit(api.Serve(index.Index(data), *address, *path))
    } else {
      os.Exit(1)
    }
  }
  flag.PrintDefaults()
  os.Exit(2)
}
