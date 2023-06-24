package task

import "fmt"

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
		Field:      d.Field,
		Type:       d.Type.Copy(),
		IsRequired: d.IsRequired,
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

func (d DataType) Validate(value interface{}) error {
	if value == nil {
		return nil
	}

	if d.Type.IsPrimitive() {
		return d.ValidatePrimitive(value)
	}

	if d.Type.IsObject() {
		return d.ValidateObject(value)
	}

	if d.Type.IsArray() {
		return d.ValidateArray(value)
	}

	return nil
}

func (d DataType) ValidatePrimitive(value interface{}) error {
	switch d.Type {
	case DataTypeString:
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %+v %w", value, ErrInvalidDataType)
		}
	case DataTypeInt:
		_, ok := value.(int)
		if !ok {
			return fmt.Errorf("expected integer, got %+v %w", value, ErrInvalidDataType)
		}
	case DataTypeFloat:
		_, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected float, got %+v %w", value, ErrInvalidDataType)
		}
	case DataTypeBool:
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected boolean, got %+v %w", value, ErrInvalidDataType)
		}
	}
	return nil
}

func (d DataType) ValidateObject(value interface{}) error {
	ma, ok := value.(map[string]interface{})
	if !ok {
		return ErrInvalidDataType
	}
	if d.ObjectDef == nil {
		return nil
	}
	for _, field := range d.ObjectDef.Fields {
		if field.IsRequired {
			if _, ok := ma[field.Field]; !ok {
				return fmt.Errorf("required field %s not found %w", field.Field, ErrInvalidDataType)
			}
		}
		if err := field.Type.Validate(ma[field.Field]); err != nil {
			return fmt.Errorf("error validating field %s %w", field.Field, err)
		}
	}
	return nil
}

func (d DataType) ValidateArray(value interface{}) error {
	arr, ok := value.([]interface{})
	if !ok {
		return ErrInvalidDataType
	}
	if d.ArrayDef == nil {
		return nil
	}
	for _, item := range arr {
		if err := d.ArrayDef.TypeDef.Validate(item); err != nil {
			return fmt.Errorf("error validating array item %w", err)
		}
	}
	return nil
}

type DataValue struct {
	DataType
	Value interface{}
}
