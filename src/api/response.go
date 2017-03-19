package api

import (
  "fmt"
  "net/http"
  "encoding/json"
)

func SendJSONResponse(w http.ResponseWriter, o interface{}) {
  //js, err := json.Marshal(o)
  js, err := json.MarshalIndent(o, "", "  ")
  if (err != nil) {
    fmt.Println("JSON ENCODING ERROR", err)
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  w.Write(js)
}

func SendTextResponse(w http.ResponseWriter, o * []byte, mime string) (int, error) {
  // js, err := json.Marshal(o)
  w.Header().Set("Content-Type", mime)
  w.WriteHeader(http.StatusOK)
  w.Write(*o)
  return 0, nil
}
