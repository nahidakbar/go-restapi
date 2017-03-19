package index

type Float64Set []float64

func (set Float64Set) IndexOf(value float64) int {
  for i, x := range set {
    if x == value {
      return i
    }
  }
  return -1
}

func (set * Float64Set) AddToSet(value float64) {
  if set.IndexOf(value) == -1 {
    *set = append(*set, value)
  }
}

func (set Float64Set) Trim() []interface{} {
  output := make([]interface{}, len(set))
  for i, x := range set {
    output[i] = x
  }
  return output
}
