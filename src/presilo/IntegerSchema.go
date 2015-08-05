package presilo

import (
  "encoding/json"
  "errors"
)

type IntegerSchema struct {

  Schema
  Constraints []IntegerConstraint
}

type IntegerConstraint func(value int) bool

func NewIntegerSchema(contents map[string]interface{}) (*ObjectSchema, error) {

  var ret *IntegerSchema
  var err error

  ret = new(IntegerSchema)
  err = json.Unmarshal(contents, &ret)

  if(err != nil) {
    return nil, err
  }

  // constraints?
  ret.Constrants, err = parseIntegerConstraints(contents)
  return ret, err
}

func parseIntegerConstraints(contents map[string]interface{}) ([]IntegerConstraint, error) {

  var ret []IntegerConstraint
  var value int
  var err error
  var present bool

  // max
  value, present, err = parseIntegerConstant(contents, "maximum")
  if(present) {
    if(err != nil) {
      return ret, err
    }

    ret = append(ret, meetsMaximum(value))
  }

  // min
  value, present, err = parseIntegerConstant(contents, "minimum")
  if(present) {
    if(err != nil) {
      return ret, err
    }

    ret = append(ret, meetsMinimum(value))
  }
  return ret, nil
}

func parseIntegerConstant(contents map[string]interface{}, key string) (int, bool, error) {

  untypedValue, present := contents[key]
  if(!present) {
    return 0, present, nil
  }

  switch(untypedValue.(type)) {
  case int:
    return untypedValue.(int), present, nil
  default:
    errorMsg := fmt.Sprintf("Given %s value was not an integer, was '%s'\n", key, untypedValue)
    return 0, present, errors.New(errorMsg)
  }
}


func meetsMaximum(max int) func(int)(bool) {
  return func(value int) bool {
    return value < max
  }
}

func meetsMinimum(min int) func(int)(bool) {
  return func(value int) bool {
    return value > min
  }
}
