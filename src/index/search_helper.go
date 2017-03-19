package index

import (
  "github.com/reiver/go-porterstemmer"
  "strings"
)

func PadWithSpace(parts []string) string {
  return " " + strings.Join(parts, " ") + " "
}

func CharClass(item rune) rune {
  if item >= 'a' && item <= 'z' {
    return 'a'
  } else if item >= '0' && item <= '9' {
    return '0'
  } else {
    return item
  }
}

func LexAndStem(str string) []string {
  tokens := make([]string, 0, len(str))
  // hack togeather a barely legal lexer
  last := '\x00'
  lastTokensIndex := -1
  for _, item := range strings.ToLower(str) {
    current := CharClass(item)
    if current != ' ' {
      if current == last {
        tokens[lastTokensIndex] = tokens[lastTokensIndex] + string(item)
      } else {
        tokens = append(tokens, string(item))
        lastTokensIndex++
      }
    }
    last = current
  }
  // stem
  for x, xx := range tokens {
    tokens[x] = porterstemmer.StemString(xx)
  }
  return tokens
}


func Strip(input []string) []string {
  lookup := make(map[string]int)
  
  output := make([]string, len(input))
  
  for i, x := range input {
    xx, has := lookup[x]
    if !has {
      xx = i
      lookup[x] = i
    }
    output[i] = input[xx]
  }
  return output
}
