package index

import "sort"

type StringSet map[string]int

func (set StringSet) IndexOf(value string) int {
  idx, has := set[value]
  if  !has {
    idx = -1
  }
  return idx
}

func (set * StringSet) AddToSet(value string) int {
  ind := set.IndexOf(value)
  if ind == -1 {
    ind = len(*set)
    (*set)[value] = ind
  }
  return ind
}

func (set StringSet) Trim() []interface{} {
  output := make([]interface{}, len(set))
  for i, x := range set {
    output[x] = i
  }
  return output
}

func (set * StringSet) Sort() {
  output := make([]string, len(*set))
  for i, x := range *set {
    output[x] = i
  }
  *set = make(StringSet)
  sort.Sort(sort.StringSlice(output))
  for i, x := range output {
    (*set)[x] = i
  }
}
