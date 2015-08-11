package presilo

import (
  "encoding/json"
  "fmt"
  "errors"
)

/*
  A schema which describes an integer.
*/
type ObjectSchema struct {

  Schema
  Properties map[string]TypeSchema
  RequiredProperties []string `json:"required"`

  // TODO: MaxProperties *int `json:"maxProperties"`
  // TODO: MinProperties *int `json:"minProperties"`
  // TODO: AdditionalProperties *bool `json:"additionalProperties"`
  // NOT SUPPORTED: patternProperties
  RawProperties map[string]*json.RawMessage `json:"properties"`

  ConstrainedProperties []string
  UnconstrainedProperties []string
}

/*
  Creates a new object schema from a byte slice that can be interpreted as json.
  Object schemas may contain multiple schemas.
*/
func NewObjectSchema(contents []byte) (*ObjectSchema, error) {

  var ret *ObjectSchema
  var sub TypeSchema
  var subschemaBytes []byte
  var err error
  var constrained bool

  ret = new(ObjectSchema)
  ret.typeCode = SCHEMATYPE_OBJECT

  err = json.Unmarshal(contents, &ret)
  if(err != nil) {
    return ret, err
  }

  ret.Properties = make(map[string]TypeSchema, len(ret.RawProperties))

  err = ret.checkRequiredProperties()
  if(err != nil) {
    return ret, err
  }

  // parse individual sub-schemas
  for propertyName, propertyContents := range ret.RawProperties {

    subschemaBytes, err = propertyContents.MarshalJSON()
    if(err != nil) {
      return ret, err
    }

    sub, err = ParseSchema(subschemaBytes, propertyName)
    if(err != nil) {
      return ret, err
    }

    ret.Properties[propertyName] = sub
  }

  // for convenience, populate "ConstrainedProperties" to all required properties,
  // along with any other properties which have constraints
  for propertyName, subschema := range ret.Properties {

    constrained = false

    switch subschema.GetSchemaType() {
      case SCHEMATYPE_INTEGER: constrained = subschema.(*IntegerSchema).HasConstraints()
      case SCHEMATYPE_NUMBER: constrained = subschema.(*NumberSchema).HasConstraints()
      case SCHEMATYPE_STRING: constrained = subschema.(*StringSchema).HasConstraints()
      case SCHEMATYPE_ARRAY: constrained = subschema.(*ArraySchema).HasConstraints()
    }

    if(constrained) {
      ret.ConstrainedProperties = append(ret.ConstrainedProperties, propertyName)
    } else {
      ret.UnconstrainedProperties = append(ret.UnconstrainedProperties, propertyName)
    }
  }
  return ret, nil
}

func (this *ObjectSchema) checkRequiredProperties() error {

  var propertyName string
  var found bool

  // make sure all required properties are defined
  for _, propertyName = range this.RequiredProperties {

    _, found = this.RawProperties[propertyName]
    if(!found) {
      errorMsg := fmt.Sprintf("Property '%s' was listed as required, but was not defined\n", propertyName)
      return errors.New(errorMsg)
    }
  }

  return nil
}
