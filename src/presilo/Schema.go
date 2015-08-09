package presilo

type TypeSchema interface {

  GetSchemaType() SchemaType
}

/*
  Represents the schema of one field in a json document.
*/
type Schema struct {

  Title string `json:"title"`
  ID string `json:"id"`
  typeCode SchemaType
}

func (this Schema) GetSchemaType() SchemaType {
  return this.typeCode
}
