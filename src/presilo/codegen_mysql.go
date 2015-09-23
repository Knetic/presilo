package presilo

import (
	"bytes"
	"fmt"
)

/*
  Generates valid mysql table creation/update query.

  SQL query generation is a bit different than other languages, it:
  - Generates one query file to represent all data structures
  - Uses "module" not as a harmless namespace, but as the DB name tables should be contained
  - Does not attempt to represent arrays
  - Assumes internationalized "nvarchar" for all strings
  - Provides a best guess for string column length based on constraints for min/max length
  - Does not provide a primary key!
  - Does not support regex constraints.
  - Uses 'bit' to represent booleans, with 0 = true, 1 = false.
*/
func GenerateMySQL(schema *ObjectSchema, module string) string {

	var ret bytes.Buffer

	ret.WriteString(generateMysqlCreate(schema, module))

	return ret.String()
}

func generateMysqlCreate(schema *ObjectSchema, module string) string {

  var ret bytes.Buffer
  var toWrite string
  var required bool

  toWrite = fmt.Sprintf("USE %s;\nGO;\n", module)
  ret.WriteString(toWrite)

  // create table
  toWrite = fmt.Sprintf("CREATE TABLE dbo.%s\n(\n", schema.GetTitle())
  ret.WriteString(toWrite)

  for propertyName, subschema := range schema.Properties {

    // determine nullability
    required = false
    for _, requiredProperty := range schema.RequiredProperties {

      if(requiredProperty == propertyName) {
        required = true
        break
      }
    }

    switch subschema.GetSchemaType() {
		case SCHEMATYPE_BOOLEAN:
			toWrite = generateMySQLBoolColumn(propertyName, required, subschema.(*BooleanSchema))
		case SCHEMATYPE_STRING:
			toWrite = generateMySQLStringColumn(propertyName, required, subschema.(*StringSchema))
		case SCHEMATYPE_INTEGER:
			toWrite = generateMySQLIntegerColumn(propertyName, required, subschema.(*IntegerSchema))
		case SCHEMATYPE_NUMBER:
			toWrite = generateMySQLNumberColumn(propertyName, required, subschema.(*NumberSchema))
		case SCHEMATYPE_OBJECT:
			toWrite = generateMySQLReferenceColumn(propertyName, required, subschema.(*ObjectSchema))
		case SCHEMATYPE_ARRAY:
			toWrite = generateMySQLArrayColumn(propertyName, required, subschema.(*ArraySchema))
		}

    ret.WriteString(toWrite)
  }

  ret.WriteString("\n)")

  // execution.
  ret.WriteString("\nGO;\n\n")
  return ret.String()
}

func generateMySQLBoolColumn(name string, required bool, schema *BooleanSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\n\t%s bit", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }
  return ret.String()
}

func generateMySQLStringColumn(name string, required bool, schema *StringSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\n\t%s nvarchar(128)", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }
  return ret.String()
}

func generateMySQLIntegerColumn(name string, required bool, schema *IntegerSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\n\t%s int(32)", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }
  return ret.String()
}

func generateMySQLNumberColumn(name string, required bool, schema *NumberSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\n\t%s float(32)", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }
  return ret.String()
}

func generateMySQLReferenceColumn(name string, required bool, schema *ObjectSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\n\t%s int(1)", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }
  return ret.String()
}

func generateMySQLArrayColumn(name string, required bool, schema *ArraySchema) string {

  var ret bytes.Buffer

  fmt.Println("Schema contains an array, which has no definite analogue in MySQL.")
  return ret.String()
}

func generateMySQLRequiredConstraint() string {
  return " NOT NULL"
}
