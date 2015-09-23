package presilo

import (
	"bytes"
	"strings"
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
	var columns []string
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

		columns = append(columns, toWrite)
  }

	ret.WriteString(strings.Join(columns, ",\n"))
  ret.WriteString("\n);")

  // execution.
  ret.WriteString("\nGO;\n\n")
  return ret.String()
}

func generateMySQLBoolColumn(name string, required bool, schema *BooleanSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\t%s bit", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }

	toWrite = fmt.Sprintf("\n\t\tCHECK(%s = 0 OR %s = 1)", name, name)
	ret.WriteString(toWrite)

  return ret.String()
}

func generateMySQLStringColumn(name string, required bool, schema *StringSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\t%s nvarchar(128)", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }

	if schema.MinLength != nil {
		ret.WriteString(generateMySQLRangeCheck(*schema.MinLength, name, "%d", false, "<", ""))
	}

	if schema.MaxLength != nil {
		ret.WriteString(generateMySQLRangeCheck(*schema.MaxLength, name, "%d", false, ">", ""))
	}

	if(schema.Enum != nil) {
		ret.WriteString(generateMySQLEnumCheck(schema, schema.GetEnum(), "'", "'"))
	}

  return ret.String()
}

func generateMySQLIntegerColumn(name string, required bool, schema *IntegerSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\t%s int", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }

	ret.WriteString(generateMySQLNumericConstraints(name, schema))
  return ret.String()
}

func generateMySQLNumberColumn(name string, required bool, schema *NumberSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\t%s float", name)
  ret.WriteString(toWrite)

  if(required) {
    ret.WriteString(generateMySQLRequiredConstraint())
  }

	ret.WriteString(generateMySQLNumericConstraints(name, schema))
  return ret.String()
}

func generateMySQLReferenceColumn(name string, required bool, schema *ObjectSchema) string {

  var ret bytes.Buffer
  var toWrite string

  toWrite = fmt.Sprintf("\t%s int(1)", name)
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
  return "\n\t\tNOT NULL"
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateMySQLEnumCheck(schema interface{}, enumValues []interface{}, prefix string, postfix string) string {

	var ret bytes.Buffer
	var constraint string
	var length int

	length = len(enumValues)

	if length <= 0 {
		return ""
	}

	// write array of valid values
	constraint = fmt.Sprintf("\n\t\tENUM(%s%v%s", prefix, enumValues[0], postfix)
	ret.WriteString(constraint)

	for _, enumValue := range enumValues[1:length] {

		constraint = fmt.Sprintf(",%s%v%s", prefix, enumValue, postfix)
		ret.WriteString(constraint)
	}
	ret.WriteString(")")

	return ret.String()
}

func generateMySQLNumericConstraints(name string, schema NumericSchemaType) string {

  var ret bytes.Buffer
  var toWrite string

	if schema.HasMinimum() {
		ret.WriteString(generateMySQLRangeCheck(schema.GetMinimum(), "value", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<"))
	}

	if schema.HasMaximum() {
		ret.WriteString(generateMySQLRangeCheck(schema.GetMaximum(), "value", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">"))
	}

	if schema.HasEnum() {
		ret.WriteString(generateMySQLEnumCheck(schema, schema.GetEnum(), "", ""))
	}

	if schema.HasMultiple() {

		toWrite = fmt.Sprintf("\n\t\tCHECK(mod(%s, %v) = 0)", name, schema.GetMultiple())
		ret.WriteString(toWrite)
	}

	return ret.String()
}

func generateMySQLRangeCheck(value interface{}, reference string, format string, exclusive bool, comparator, exclusiveComparator string) string {

	var compareString string

	if exclusive {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	return fmt.Sprintf("\n\t\tCHECK(%s %s %v)", reference, compareString, value)
}
