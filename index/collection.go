package index

// read only collection

type Collection struct {
  schema * Schema
  search * Search
}

func (collection Collection) Schema() interface{} {
  return collection.schema
}

func (collection Collection) SearchMeta() interface{} {
  return collection.search
}

func (collection Collection) GetItem(index int) interface{} {
  return collection.schema.GetItem(index)
}

func (collection Collection) TotalItems() int {
  return collection.schema.TotalItems
}

func (collection Collection) Search(query map[string][]string) interface{} {
  return collection.search.Search(query, collection.schema)
}
