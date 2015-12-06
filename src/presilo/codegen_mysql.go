package presilo

import (
	"fmt"
	"regexp"
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
func GenerateMySQL(schema *ObjectSchema, module string, tabstyle string) string {

	var buffer *BufferedFormatString

	buffer = NewBufferedFormatString(tabstyle)
	generateMysqlCreate(schema, module, buffer)

	return buffer.String()
}

func ValidateMySQLModule(module string) bool {

	// ref: http://dev.mysql.com/doc/refman/5.7/en/identifiers.html
	pattern := "^[0-9a-zA-Z$_]+$"
	matched, err := regexp.MatchString(pattern, module)
	return err == nil && matched
}

func generateMysqlCreate(schema *ObjectSchema, module string, buffer *BufferedFormatString) {

  var required, firstProperty bool

  buffer.Printf("USE %s;\n", module)

  // create table
  buffer.Printf("CREATE TABLE %s\n(", schema.GetTitle())
  buffer.AddIndentation(1)
	generateMySQLPrimaryKey(schema, buffer)
	buffer.Printf("\n\n")

	firstProperty = true

  for propertyName, subschema := range schema.Properties {

    // determine nullability
    required = false
    for _, requiredProperty := range schema.RequiredProperties {

      if(requiredProperty == propertyName) {
        required = true
        break
      }
    }

		if(firstProperty) {
			firstProperty = false
		} else {
			buffer.Print(",\n")
		}

    switch subschema.GetSchemaType() {
		case SCHEMATYPE_BOOLEAN:
			generateMySQLBoolColumn(propertyName, required, subschema.(*BooleanSchema), buffer)
		case SCHEMATYPE_STRING:
			generateMySQLStringColumn(propertyName, required, subschema.(*StringSchema), buffer)
		case SCHEMATYPE_INTEGER:
			generateMySQLIntegerColumn(propertyName, required, subschema.(*IntegerSchema), buffer)
		case SCHEMATYPE_NUMBER:
			generateMySQLNumberColumn(propertyName, required, subschema.(*NumberSchema), buffer)
		case SCHEMATYPE_OBJECT:
			generateMySQLReferenceColumn(propertyName, required, subschema.(*ObjectSchema), buffer)
		case SCHEMATYPE_ARRAY:
			generateMySQLArrayColumn(propertyName, required, subschema.(*ArraySchema), buffer)
		}
  }

	buffer.AddIndentation(-1)
  buffer.Print("\n);")

  // execution.
  buffer.Print("\n\n")
}

func generateMySQLBoolColumn(name string, required bool, schema *BooleanSchema, buffer *BufferedFormatString) {

  buffer.Printf("\t%s bit", name)
  buffer.AddIndentation(1)

  if(required) {
    generateMySQLRequiredConstraint(buffer)
  }

	buffer.Printf("\nCHECK(%s = 0 OR %s = 1)", name, name)
	buffer.AddIndentation(-1)
}

func generateMySQLStringColumn(name string, required bool, schema *StringSchema, buffer *BufferedFormatString) {

  buffer.Printf("%s nvarchar(128)", name)
  buffer.AddIndentation(1)

  if(required) {
    generateMySQLRequiredConstraint(buffer)
  }

	if schema.MinLength != nil {
		generateMySQLRangeCheck(*schema.MinLength, name, "%d", false, "<", "", buffer)
	}

	if schema.MaxLength != nil {
		generateMySQLRangeCheck(*schema.MaxLength, name, "%d", false, ">", "", buffer)
	}

	if(schema.Enum != nil) {
		generateMySQLEnumCheck(schema, schema.GetEnum(), "'", "'", buffer)
	}

	buffer.AddIndentation(-1)
}

func generateMySQLIntegerColumn(name string, required bool, schema *IntegerSchema, buffer *BufferedFormatString) {

  buffer.Printf("%s int", name)
  buffer.AddIndentation(1)

  if(required) {
    generateMySQLRequiredConstraint(buffer)
  }

	generateMySQLNumericConstraints(name, schema, buffer)
  buffer.AddIndentation(-1)
}

func generateMySQLNumberColumn(name string, required bool, schema *NumberSchema, buffer *BufferedFormatString) {

  buffer.Printf("%s float", name)
  buffer.AddIndentation(1)

  if(required) {
    generateMySQLRequiredConstraint(buffer)
  }

	generateMySQLNumericConstraints(name, schema, buffer)
	buffer.AddIndentation(-1)
}

func generateMySQLReferenceColumn(name string, required bool, schema *ObjectSchema, buffer *BufferedFormatString) {

  buffer.Printf("%s__id int(4)", name)
  buffer.AddIndentation(1)

  if(required) {
    generateMySQLRequiredConstraint(buffer)
  }

	// add foreign key constraint.
	buffer.Printf(",\nFOREIGN KEY(%s__id)", name)
	buffer.AddIndentation(1)

	buffer.Printf("\nREFERENCES %s(__id)", schema.GetTitle())
	buffer.Printf("\nON DELETE CASCADE")

	buffer.AddIndentation(-1)
	buffer.AddIndentation(-1)
}

func generateMySQLPrimaryKey(schema *ObjectSchema, buffer *BufferedFormatString) {
	buffer.Printf("\n__id int NOT NULL,")
	buffer.AddIndentation(1)
	buffer.Printf("\nPRIMARY KEY(__id),")
	buffer.AddIndentation(-1)
}

func generateMySQLArrayColumn(name string, required bool, schema *ArraySchema, buffer *BufferedFormatString) {

  fmt.Println("Schema contains an array, which has no definite analogue in MySQL.")
}

func generateMySQLRequiredConstraint(buffer *BufferedFormatString) {

  buffer.Print("\nNOT NULL")
}

/*
	Generates code which throws an error if the given [parameter]'s value is not contained in the given [validValues].
*/
func generateMySQLEnumCheck(schema interface{}, enumValues []interface{}, prefix string, postfix string, buffer *BufferedFormatString) {

	var schemaName string
	var length int

	schemaName = ToJavaCase((schema.(TypeSchema)).GetTitle())
	length = len(enumValues)

	if length <= 0 {
		return
	}

	// write array of valid values
	buffer.Printf(",\nCONSTRAINT %sValuesCheck CHECK(%s in (%s%v%s", schemaName, schemaName, prefix, enumValues[0], postfix)

	for _, enumValue := range enumValues[1:length] {
		buffer.Printf(",%s%v%s", prefix, enumValue, postfix)
	}

	buffer.Print("))")
}

func generateMySQLNumericConstraints(name string, schema NumericSchemaType, buffer *BufferedFormatString) {

	if schema.HasMinimum() {
		generateMySQLRangeCheck(schema.GetMinimum(), "value", schema.GetConstraintFormat(), schema.IsExclusiveMinimum(), "<=", "<", buffer)
	}

	if schema.HasMaximum() {
		generateMySQLRangeCheck(schema.GetMaximum(), "value", schema.GetConstraintFormat(), schema.IsExclusiveMaximum(), ">=", ">", buffer)
	}

	if schema.HasEnum() {
		generateMySQLEnumCheck(schema, schema.GetEnum(), "", "", buffer)
	}

	if schema.HasMultiple() {
		buffer.Printf("\nCHECK(mod(%s, %v) = 0)", name, schema.GetMultiple())
	}
}

func generateMySQLRangeCheck(value interface{}, reference string, format string, exclusive bool, comparator, exclusiveComparator string, buffer *BufferedFormatString) {

	var compareString string

	if exclusive {
		compareString = exclusiveComparator
	} else {
		compareString = comparator
	}

	buffer.Printf("\nCHECK(%s %s %v)", reference, compareString, value)
}
