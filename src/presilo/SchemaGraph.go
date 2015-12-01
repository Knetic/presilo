package presilo

/*
  Represents a dependency graph that can order schemas.
*/
type SchemaGraph struct {

  nodes []*SchemaGraphNode
}

type SchemaGraphNode struct {

  schema *ObjectSchema
  neighbors []*SchemaGraphNode
}

func NewSchemaGraph(schemas []*ObjectSchema) *SchemaGraph {

  var ret *SchemaGraph
  var node *SchemaGraphNode

  ret = new(SchemaGraph)

  for _, schema := range schemas {

    if(schema.GetSchemaType() != SCHEMATYPE_OBJECT) {
      continue
    }

    node = new(SchemaGraphNode)
    node.schema = schema
    ret.nodes = append(ret.nodes, node)
  }

  for _, node = range ret.nodes {
    node.discoverNeighbors(ret)
  }

  return ret
}

/*
  Returns a slice of schemas which represents an ordering where dependent schemas are given first.
  If there is a circular dependency in this graph, an error is returned instead.
*/
func (this *SchemaGraph) GetOrderedSchemas() ([]*ObjectSchema, error) {

  var ret []*ObjectSchema
  //ret := make([]*ObjectSchema, len(this.nodes))

  for _, node := range this.nodes {
    ret = resolveDependency(node, ret)
  }

  return ret, nil
}

func resolveDependency(node *SchemaGraphNode, resolution []*ObjectSchema) ([]*ObjectSchema) {

  // otherwise, descend deeper into the object.
  for _, neighbor := range node.neighbors {
      resolution = resolveDependency(neighbor, resolution)
  }

  if(!elementExistsInSlice(node.schema, resolution)) {
    resolution = append(resolution, node.schema)
  }

  return resolution
}

/*
  Adds a dependency between the given [source] node and the node which contains the [target] schema.
*/
func (this *SchemaGraph) addDependency(source *SchemaGraphNode, target *ObjectSchema) {

  for _, node := range this.nodes {

    if(node.schema == target) {

      source.addNeighbor(node)
      break
    }
  }
}

func (this *SchemaGraphNode) addNeighbor(neighbor *SchemaGraphNode) {
  this.neighbors = append(this.neighbors, neighbor)
}

func (this *SchemaGraphNode) discoverNeighbors(graph *SchemaGraph) {

  var schema *ObjectSchema
  var subschema TypeSchema

  // only object schemas can possibly have dependencies
  if(this.schema.GetSchemaType() == SCHEMATYPE_OBJECT) {

    schema = this.schema

    for _, subschema = range schema.Properties {

      if(subschema.GetSchemaType() == SCHEMATYPE_OBJECT) {
        graph.addDependency(this, subschema.(*ObjectSchema))
      }
    }
  }
}

/*
  Returns true if the given [element] exists within the given [slice].
*/
func elementExistsInSlice(element *ObjectSchema, slice []*ObjectSchema) bool {

  for _, e := range slice {
    if(e == element) {
      return true
    }
  }
  return false
}
