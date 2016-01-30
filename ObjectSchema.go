package presilo

import (
	"encoding/json"
	"errors"
	"fmt"
)

/*
  A schema which describes an integer.
*/
type ObjectSchema struct {
	Schema
	Properties         map[string]TypeSchema
	RequiredProperties []string `json:"required"`

	// TODO: MaxProperties *int `json:"maxProperties"`
	// TODO: MinProperties *int `json:"minProperties"`
	// TODO: AdditionalProperties *bool `json:"additionalProperties"`
	// NOT SUPPORTED: patternProperties
	RawProperties map[string]*json.RawMessage `json:"properties"`

	ConstrainedProperties   SortableStringArray
	UnconstrainedProperties SortableStringArray
}

/*
  Creates a new object schema from a byte slice that can be interpreted as json.
  Object schemas may contain multiple schemas.
*/
func NewObjectSchema(contents []byte, context *SchemaParseContext) (*ObjectSchema, error) {

	var ret *ObjectSchema
	var sub TypeSchema
	var subschemaBytes []byte
	var err error

	ret = new(ObjectSchema)
	ret.typeCode = SCHEMATYPE_OBJECT

	err = json.Unmarshal(contents, &ret)
	if err != nil {
		return ret, err
	}

	ret.Properties = make(map[string]TypeSchema, len(ret.RawProperties))

	err = ret.checkRequiredProperties()
	if err != nil {
		return ret, err
	}

	// parse individual sub-schemas
	for propertyName, propertyContents := range ret.RawProperties {

		subschemaBytes, err = propertyContents.MarshalJSON()
		if err != nil {
			return ret, err
		}

		sub, err = ParseSchema(subschemaBytes, propertyName, context)
		if err != nil {
			return ret, err
		}

		ret.Properties[propertyName] = sub
	}

	// for convenience, populate "ConstrainedProperties" to all required properties,
	// along with any other properties which have constraints
	for propertyName, subschema := range ret.Properties {

		if subschema.HasConstraints() {
			ret.ConstrainedProperties = append(ret.ConstrainedProperties, propertyName)
		} else {
			ret.UnconstrainedProperties = append(ret.UnconstrainedProperties, propertyName)
		}
	}

	ret.ConstrainedProperties.Sort()
	ret.UnconstrainedProperties.Sort()
	return ret, nil
}

func (this *ObjectSchema) AddProperty(name string, schema TypeSchema) {

	this.Properties[name] = schema

	if(schema.HasConstraints()) {
		this.ConstrainedProperties = append(this.ConstrainedProperties, name)
		this.ConstrainedProperties.Sort()
	} else {
		this.UnconstrainedProperties = append(this.UnconstrainedProperties, name)
		this.UnconstrainedProperties.Sort()
	}
}

/*
	Returns an ordered array of property names, guaranteed to be the same for the same schema input over multiple runs of the program.
*/
func (this *ObjectSchema) GetOrderedPropertyNames() []string {

	var ret SortableStringArray

	// fill
	for key, _ := range this.Properties {
		ret = append(ret, key)
	}

	ret.Sort()
	return ret
}

func (this *ObjectSchema) checkRequiredProperties() error {

	var propertyName string
	var found bool

	// make sure all required properties are defined
	for _, propertyName = range this.RequiredProperties {

		_, found = this.RawProperties[propertyName]
		if !found {
			errorMsg := fmt.Sprintf("Property '%s' was listed as required, but was not defined\n", propertyName)
			return errors.New(errorMsg)
		}
	}

	return nil
}

func (this *ObjectSchema) HasConstraints() bool {
	return false
}
