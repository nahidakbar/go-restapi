package api

// read only collection

type Collection interface {
  Schema() interface{}
  SearchMeta() interface{}
  GetItem(int) interface{}
  TotalItems() int
  Search(map[string][]string) interface{}
}
