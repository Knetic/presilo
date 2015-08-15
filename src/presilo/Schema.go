package presilo

type TypeSchema interface {
	GetSchemaType() SchemaType
	GetTitle() string
	GetDescription() string
	SetTitle(string)
	GetID() string
	HasConstraints() bool
}

/*
  Represents the schema of one field in a json document.
*/
type Schema struct {
	Title       string `json:"title"`
	ID          string `json:"id"`
	Description string `json:"description"`
	typeCode    SchemaType
}

func (this *Schema) GetSchemaType() SchemaType {
	return this.typeCode
}

func (this *Schema) GetTitle() string {
	return this.Title
}

func (this *Schema) GetDescription() string {
	return this.Description
}

func (this *Schema) SetTitle(title string) {
	this.Title = title
}

func (this *Schema) GetID() string {
	return this.ID
}
