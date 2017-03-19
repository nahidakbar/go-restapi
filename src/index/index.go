package index

func Index (data map[string][]interface{}) (*Schema, *Search) {
  schema := new(Schema)

  search := new(Search)
  
  go func(){
    schema.Initialise(data);
    search.Initialise(schema);
  }()
  
  return schema, search
}
