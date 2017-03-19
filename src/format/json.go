package format

import (
  "os"
  "bufio"
  "encoding/json"
  "io/ioutil"
)


func LoadJSOND (localfile string) (map[string][]interface{}, error) {
  var items []interface{} = make([]interface{}, 0)
  
  file, err := os.Open(localfile)
  if err != nil {
    return nil, err
  }
  
  defer file.Close()
  
  scanner := bufio.NewScanner(file)
  
  for scanner.Scan() {
    bytes := scanner.Bytes();
    if (len(bytes) >= 2) {
      var obj interface{}
      json.Unmarshal(bytes, &obj)
      items = append(items, obj);
    }
  }
  if err := scanner.Err(); err != nil {
    return nil, err
  }

  heads := make(map[string]string, 0)
  for _, item := range items {
    for h, _ := range item.(map[string]interface{}) {
      heads[h] = h
    }
  }
  head := make([]string, 0)
  for h, _ := range heads {
    head = append(head, h)
  }
  
  output := make([][]interface{}, len(head))
  for index, _ := range head {
    output[index] = make([]interface{}, 0, len(items))
  }
  
  for _, record := range items {
    for index, h := range head {
      data, has := record.(map[string]interface{})[h]
      if has {
        output[index] = append(output[index], data)
      } else {
        output[index] = append(output[index], nil)
      }
    }
  }

  outdata := make(map[string][]interface{});
  for index, heading := range head {
    outdata[heading] = output[index]
  }
  
  return outdata, nil
}

func LoadJSON (localfile string) (map[string][]interface{}, error) {
  
  file, err := ioutil.ReadFile(localfile) ; if err != nil { return nil, err }
  
  items := make([]interface{}, 0)
  json.Unmarshal(file, &items)
  
  heads := make(map[string]string, 0)
  for _, item := range items {
    for h, _ := range item.(map[string]interface{}) {
      heads[h] = h
    }
  }
  head := make([]string, 0)
  for h, _ := range heads {
    head = append(head, h)
  }
  
  output := make([][]interface{}, len(head))
  for index, _ := range head {
    output[index] = make([]interface{}, 0, len(items))
  }
  
  for _, record := range items {
    for index, h := range head {
      data, has := record.(map[string]interface{})[h]
      if has {
        output[index] = append(output[index], data)
      } else {
        output[index] = append(output[index], nil)
      }
    }
  }

  outdata := make(map[string][]interface{});
  for index, heading := range head {
    outdata[heading] = output[index]
  }
  
  return outdata, nil
}
