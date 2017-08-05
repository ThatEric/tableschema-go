package schema

import (
	"encoding/json"
	"fmt"
)

// Default for schema fields.
const (
	defaultFieldType   = "string"
	defaultFieldFormat = "default"
)

// Default schema variables.
var (
	defaultTrueValues  = []string{"yes", "y", "true", "t", "1"}
	defaultFalseValues = []string{"no", "n", "false", "f", "0"}
)

// Field types.
const (
	IntegerType  = "integer"
	StringType   = "string"
	BooleanType  = "boolean"
	NumberType   = "number"
	DateType     = "date"
	ObjectType   = "object"
	ArrayType    = "array"
	DateTimeType = "datetime"
	TimeType     = "time"
)

// Formats.
const (
	AnyDateFormat = "any"
)

// JSON object that describes a single field.
// More: https://specs.frictionlessdata.io/table-schema/#field-descriptors
type Field struct {
	// Name of the field. It is mandatory and shuold correspond to the name of field/column in the data file (if it has a name).
	Name   string `json:"name"`
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
	// A human readable label or title for the field.
	Title string `json:"title,omitempty"`
	// A description for this field e.g. "The recipient of the funds"
	Description string `json:"description,omitempty"`

	// Boolean properties. Define set of the values that represent true and false, respectively.
	// https://specs.frictionlessdata.io/table-schema/#boolean
	TrueValues  []string `json:"trueValues,omitempty"`
	FalseValues []string `json:"falseValues,omitempty"`
}

// UnmarshalJSON sets *f to a copy of data. It will respect the default values
// described at: https://specs.frictionlessdata.io/table-schema/
func (f *Field) UnmarshalJSON(data []byte) error {
	// This is neded so it does not call UnmarshalJSON from recursively.
	type fieldAlias Field
	u := &fieldAlias{
		Type:        defaultFieldType,
		Format:      defaultFieldFormat,
		TrueValues:  defaultTrueValues,
		FalseValues: defaultFalseValues,
	}
	if err := json.Unmarshal(data, u); err != nil {
		return err
	}
	*f = Field(*u)
	return nil
}

// CastValue casts a value against field. Returns an error if the value can
// not be cast or any field constraint can no be satisfied.
func (f *Field) CastValue(value string) (interface{}, error) {
	switch f.Type {
	case IntegerType:
		return castInt(value)
	case StringType:
		return castString(f.Format, value)
	case BooleanType:
		return castBoolean(value, f.TrueValues, f.FalseValues)
	case NumberType:
		return castNumber(value)
	case DateType:
		return castDate(f.Format, value)
	case ObjectType:
		return castObject(value)
	case ArrayType:
		return castArray(value)
	case TimeType:
		return castTime(f.Format, value)
	}
	return nil, fmt.Errorf("invalid field type: %s", f.Type)
}

// TestValue checks whether the value can be casted against the field.
func (f *Field) TestValue(value string) bool {
	_, err := f.CastValue(value)
	return err == nil
}

// asReadField returns the field passed-in as parameter like it's been read as JSON.
// That include setting default values.
// Created for being used in tests.
// IMPORTANT: Not ready for being used in production due to possibly bad performance.
func asJSONField(f Field) Field {
	var out Field
	data, _ := json.Marshal(&f)
	json.Unmarshal(data, &out)
	return out
}
