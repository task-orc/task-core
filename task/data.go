package task

type DataTypeBasic string

const (
	DataTypeString DataTypeBasic = "string"
	DataTypeInt    DataTypeBasic = "int"
	DataTypeFloat  DataTypeBasic = "float"
	DataTypeBool   DataTypeBasic = "bool"
	DataTypeObject DataTypeBasic = "object"
	DataTypeArray  DataTypeBasic = "array"
)

func (d DataTypeBasic) IsPrimitive() bool {
	return d == DataTypeString || d == DataTypeInt || d == DataTypeFloat || d == DataTypeBool
}

func (d DataTypeBasic) IsObject() bool {
	return d == DataTypeObject
}

func (d DataTypeBasic) IsArray() bool {
	return d == DataTypeArray
}

// DataObjectDef is a struct that represents the Data object definition
// Eg: { "name": "string", "age": "int", "address": { "street": "string", "city": "string" }, "hobbies": ["string"] }
// Def would be:
//
//	DataObjectDef{
//		Fields: []DataField{
//			{
//				Field: "name",
//				Type: DataType{
//					Type: "string",
//				},
//			},
//			{
//				Field: "age",
//				Type: DataType{
//					Type: "int",
//				},
//			},
//			{
//				Field: "address",
//				Type: DataType{
//					Type: "object",
//					ObjectDef: &DataObjectDef{
//						Fields: []DataField{
//							{
//								Field: "street",
//								Type: DataType{
//									Type: "string",
//								},
//							},
//							{
//								Field: "city",
//								Type: DataType{
//									Type: "string",
//								},
//							},
//						},
//					},
//				},
//			},
//			{
//				Field: "hobbies",
//				Type: DataType{
//					Type: "array",
//					ArrayDef: &DataArrayDef{
//						TypeDef: DataType{
//							Type: "string",
//						},
//					},
//				},
//			},
//		},
//	}
type DataObjectDef struct {
	Fields []DataField `json:"fields"`
}

func (d DataObjectDef) CreateDataValue() *DataValue {
	result := DataValue{
		DataType: DataType{
			Type:      DataTypeObject,
			ObjectDef: d.Copy(),
		},
		Value: nil,
	}
	return &result
}

func (d DataObjectDef) Copy() *DataObjectDef {
	result := &DataObjectDef{}
	for _, field := range d.Fields {
		result.Fields = append(result.Fields, field.Copy())
	}
	return result
}

type DataField struct {
	Field      string   `json:"field"`
	IsRequired bool     `json:"isRequired"`
	Type       DataType `json:"type"`
}

func (d DataField) Copy() DataField {
	return DataField{
		Field: d.Field,
		Type:  d.Type.Copy(),
	}
}

type DataArrayDef struct {
	TypeDef DataType `json:"typeDef"`
}

func (d DataArrayDef) Copy() *DataArrayDef {
	return &DataArrayDef{
		TypeDef: d.TypeDef.Copy(),
	}
}

type DataType struct {
	Type      DataTypeBasic  `json:"type"`
	ArrayDef  *DataArrayDef  `json:"arrayDef"`
	ObjectDef *DataObjectDef `json:"objectDef"`
}

func (d DataType) Copy() DataType {
	result := DataType{
		Type: d.Type,
	}
	if d.ArrayDef != nil {
		result.ArrayDef = d.ArrayDef.Copy()
	}
	if d.ObjectDef != nil {
		result.ObjectDef = d.ObjectDef.Copy()
	}
	return result
}

type DataValue struct {
	DataType
	Value interface{}
}

func (d DataValue) IsNull() bool {
	if d.Value == nil {
		return true
	}

	ma, ok := d.Value.(map[string]interface{})
	if ok {
		return len(ma) == 0
	}
	str, ok := d.Value.(string)
	if ok {
		return len(str) == 0
	}
	//TODO: Check for empty array
	arr, ok := d.Value.([]interface{})
	if ok {
		return len(arr) == 0
	}
	return false
}
