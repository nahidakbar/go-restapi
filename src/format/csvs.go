package format

import (
  "strings"
  "io/ioutil"
  "encoding/csv"
  "io"
  "fmt"
)

func LoadCSV1 (localfile string, delim rune) (map[string][]interface{}, error) {
  file, err := ioutil.ReadFile(localfile); if err != nil { return nil, err }
  
  r := csv.NewReader(strings.NewReader(string(file)))
  r.Comma = delim

  head, err := r.Read(); if err != nil { return nil, err }
  
  output := make([][]interface{}, len(head))
  for index, _ := range head {
    output[index] = make([]interface{}, 0)
  }
  
  for {
    record, err := r.Read()
    if err == io.EOF {
      break
    }
    if err != nil {
      fmt.Print(" ", err, "; line ignored! ")
    } else {
      if len(record) == len(head) {
        for index, _ := range head {
          if (len(record[index]) > 0) {
            output[index] = append(output[index], record[index])
          } else {
            output[index] = append(output[index], nil)
          }
        }
      }
    }
  }

  outdata := make(map[string][]interface{});
  for index, heading := range head {
    outdata[heading] = output[index]
  }
  
  return outdata, nil
}


func LoadCSV2 (localfile string, delim string) (map[string][]interface{}, error) {
  file, err := ioutil.ReadFile(localfile); if err != nil { return nil, err }

  data := strings.Split(string(file), "\n");

  head := strings.Split(data[0], delim)

  output := make([][]interface{}, len(head))

  for index, _ := range head {
    output[index] = make([]interface{}, 0, len(data) - 1)
  }

  for _, row := range data[1:] {
    record := strings.Split(row, delim)
    if len(record) == len(head) {
      for index, _ := range head {
        if (len(record[index]) > 0) {
          output[index] = append(output[index], record[index])
        } else {
          output[index] = append(output[index], nil)
        }
      }
    }
  }

  outdata := make(map[string][]interface{});
  for index, heading := range head {
    outdata[heading] = output[index]
  }
  return outdata, nil
}
