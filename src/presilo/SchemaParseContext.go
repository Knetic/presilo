package presilo

/*
  Contains parsing context, such as the currently-defined schemas by ID, and schema-local definitions.
*/
type SchemaParseContext struct {

  SchemaDefinitions map[string]TypeSchema
}

func NewSchemaParseContext() *SchemaParseContext {

  var ret *SchemaParseContext

  ret = new(SchemaParseContext)
  ret.SchemaDefinitions = make(map[string]TypeSchema)
  return ret
}
