package index

func Index (data map[string][]interface{}) (Collection) {
  schema := new(Schema)

  search := new(Search)
  
  go func(){
    schema.Initialise(data);
    search.Initialise(schema);
  }()
  
  return Collection{schema: schema, search: search}
}
