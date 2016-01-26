package presilo

/*
	Represents a schema which could not be resolved when it was encountered.
	Mainly serves as a placeholder so that when all actual schemas are defined,
	a later process can use this reference to properly link references.
	This should NEVER be used during codegen.
*/
type UnresolvedSchema struct {
	Reference string
}

func NewUnresolvedSchema(ref string) *UnresolvedSchema {

	ret := new(UnresolvedSchema)
	ret.Reference = ref
	return ret
}

func (this *UnresolvedSchema) GetSchemaType() SchemaType {
	return SCHEMATYPE_UNRESOLVED
}

func (this *UnresolvedSchema) GetID() string {
	return this.Reference
}

// Used to satisfy the TypeSchema contract, stub.
func (this *UnresolvedSchema) GetTitle() string {
	return ""
}

// Used to satisfy the TypeSchema contract, stub.
func (this *UnresolvedSchema) GetDescription() string {
	return ""
}

// Used to satisfy the TypeSchema contract, stub.
func (this *UnresolvedSchema) SetTitle(string) {
}

// Used to satisfy the TypeSchema contract, stub.
func (this *UnresolvedSchema) GetNullable() bool {
	return false
}

// Used to satisfy the TypeSchema contract, stub.
func (this *UnresolvedSchema) SetNullable(bool) {
}

// Used to satisfy the TypeSchema contract, stub.
func (this *UnresolvedSchema) HasConstraints() bool {
	return false
}
