package main

import (
  "flag"
  "os"
  "log"
  "runtime"
  "runtime/pprof"
  "./format"
  "./index"
  "./api"
)

var address = flag.String("address", ":8080", "")
var datafile = flag.String("datafile", "", "")
var path = flag.String("path", "/", "")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
  flag.Parse()
  if len(*datafile) > 0 {
    var search *index.Search = nil
    var schema *index.Schema = nil
    if data, err := format.Load(*datafile); err == nil {
      schema, search = index.Index(data)
    }
    if search != nil && schema != nil {
      os.Exit(api.Serve(schema, search, *address, *path))
    } else {
      os.Exit(1)
    }
  }
  
  flag.PrintDefaults()
  os.Exit(2)
  
  if *cpuprofile != "" {
      f, err := os.Create(*cpuprofile)
      if err != nil {
          log.Fatal("could not create CPU profile: ", err)
      }
      if err := pprof.StartCPUProfile(f); err != nil {
          log.Fatal("could not start CPU profile: ", err)
      }
      defer pprof.StopCPUProfile()
  }

  if *memprofile != "" {
      f, err := os.Create(*memprofile)
      if err != nil {
          log.Fatal("could not create memory profile: ", err)
      }
      runtime.GC() // get up-to-date statistics
      if err := pprof.WriteHeapProfile(f); err != nil {
          log.Fatal("could not write memory profile: ", err)
      }
      f.Close()
  }
  
}
