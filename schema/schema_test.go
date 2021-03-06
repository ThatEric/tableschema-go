package schema

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/frictionlessdata/tableschema-go/table"
	"github.com/matryer/is"
)

func ExampleSchema_Decode() {
	// Lets assume we have a schema ...
	s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}

	// And a Table.
	t := table.FromSlices([]string{"Name", "Age"}, [][]string{
		{"Foo", "42"},
		{"Bar", "43"}})

	// And we would like to process them using Go types. First we need to create a struct to
	// hold the content of each row.
	// The tag tableheader maps the field to the schema. If no tag is set the name of the field
	// has to be the same like inside the schema.
	type person struct {
		MyName string `tableheader:"Name"`
		Age    int
	}

	// Now it is a matter of iterate over the table and Decode each row.
	iter, _ := t.Iter()
	for iter.Next() {
		var p person
		s.Decode(iter.Row(), &p)
		fmt.Printf("%+v\n", p)
	}
	// Output: {MyName:Foo Age:42}
	// {MyName:Bar Age:43}
}

func ExampleSchema_DecodeTable() {
	// Lets assume we have a schema ...
	s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}

	// And a Table.
	t := table.FromSlices([]string{"Name", "Age"}, [][]string{
		{"Foo", "42"},
		{"Bar", "43"}})

	// And we would like to process them using Go types. First we need to create a struct to
	// hold the content of each row.
	// The tag tableheader maps the field to the schema. If no tag is set the name of the field
	// has to be the same like inside the schema.
	type person struct {
		MyName string `tableheader:"Name"`
		Age    int
	}
	var people []person
	s.DecodeTable(t, &people)
	fmt.Print(people)
	// Output: [{Foo 42} {Bar 43}]
}

func ExampleSchema_Encode() {
	// Lets assume we have a schema.
	s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}

	// And would like to create a CSV out of this list. The tag tableheader maps
	// the field to the schema name. If no tag is set the name of the field
	// has to be the same like inside the schema.
	people := []struct {
		MyName string `tableheader:"Name"`
		Age    int
	}{{"Foo", 42}, {"Bar", 43}}

	// First create the writer and write the header.
	w := table.NewStringWriter()
	w.Write([]string{"Name", "Age"})

	// Then write the list
	for _, person := range people {
		row, _ := s.Encode(person)
		w.Write(row)
	}
	w.Flush()
	fmt.Print(w.String())
	// Output: Name,Age
	// Foo,42
	// Bar,43
}

func ExampleSchema_EncodeTable() {
	// Lets assume we have a schema.
	s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}

	// And would like to create a CSV out of this list. The tag tableheader maps
	// the field to the schema name. If no tag is set the name of the field
	// has to be the same like inside the schema.
	people := []struct {
		MyName string `tableheader:"Name"`
		Age    int
	}{{"Foo", 42}, {"Bar", 43}}

	// Then encode the people slice into a slice of rows.
	rows, _ := s.EncodeTable(people)

	// Now, simply write it down.
	w := table.NewStringWriter()
	w.Write([]string{"Name", "Age"})
	w.WriteAll(rows)
	w.Flush()
	fmt.Print(w.String())
	// Output: Name,Age
	// Foo,42
	// Bar,43
}

func TestLoadRemote(t *testing.T) {
	is := is.New(t)
	h := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"fields": [{"name": "ID", "type": "integer"}]}`)
	}
	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()
	got, err := LoadRemote(ts.URL)
	is.NoErr(err)

	want := &Schema{Fields: []Field{asJSONField(Field{Name: "ID", Type: "integer"})}}
	is.Equal(got, want)

	t.Run("Error", func(t *testing.T) {
		is := is.New(t)
		_, err := LoadRemote("invalidURL")
		is.True(err != nil)
	})
}

func TestRead_Sucess(t *testing.T) {
	data := []struct {
		Desc   string
		JSON   string
		Schema Schema
	}{
		{
			"OneField",
			`{
                "fields":[{"name":"n","title":"ti","type":"integer","description":"desc","format":"f","trueValues":["ntrue"],"falseValues":["nfalse"]}]
            }`,
			Schema{
				Fields: []Field{{Name: "n", Title: "ti", Type: "integer", Description: "desc", Format: "f", TrueValues: []string{"ntrue"}, FalseValues: []string{"nfalse"},
					DecimalChar: defaultDecimalChar, GroupChar: defaultGroupChar, BareNumber: defaultBareNumber}},
			},
		},
		{
			"MultipleFields",
			`{
                "fields":[{"name":"n1","type":"t1","format":"f1","falseValues":[]}, {"name":"n2","type":"t2","format":"f2","trueValues":[]}]
            }`,
			Schema{
				Fields: []Field{
					{Name: "n1", Type: "t1", Format: "f1", TrueValues: defaultTrueValues, FalseValues: []string{}, DecimalChar: defaultDecimalChar, GroupChar: defaultGroupChar, BareNumber: defaultBareNumber},
					{Name: "n2", Type: "t2", Format: "f2", TrueValues: []string{}, FalseValues: defaultFalseValues, DecimalChar: defaultDecimalChar, GroupChar: defaultGroupChar, BareNumber: defaultBareNumber},
				},
			},
		},
		{
			"PKString",
			`{"fields":[{"name":"n1"}], "primaryKey":"n1"}`,
			Schema{Fields: []Field{asJSONField(Field{Name: "n1"})}, PrimaryKeys: []string{"n1"}},
		},
		{
			"PKSlice",
			`{"fields":[{"name":"n1"}], "primaryKey":["n1"]}`,
			Schema{Fields: []Field{asJSONField(Field{Name: "n1"})}, PrimaryKeys: []string{"n1"}},
		},
		{
			"FKFieldsString",
			`{"fields":[{"name":"n1"}], "foreignKeys":{"fields":"n1"}}`,
			Schema{Fields: []Field{asJSONField(Field{Name: "n1"})}, ForeignKeys: ForeignKeys{Fields: []string{"n1"}}},
		},
		{
			"FKFieldsSlice",
			`{"fields":[{"name":"n1"}], "foreignKeys":{"fields":["n1"]}}`,
			Schema{Fields: []Field{asJSONField(Field{Name: "n1"})}, ForeignKeys: ForeignKeys{Fields: []string{"n1"}}},
		},
		{
			"FKReferenceFieldsString",
			`{"fields":[{"name":"n1"}], "foreignKeys":{"reference":{"fields":"n1"}}}`,
			Schema{Fields: []Field{asJSONField(Field{Name: "n1"})}, ForeignKeys: ForeignKeys{Reference: ForeignKeyReference{Fields: []string{"n1"}}}},
		},
		{
			"FKReferenceFieldsSlice",
			`{"fields":[{"name":"n1"}], "foreignKeys":{"reference":{"fields":["n1"]}}}`,
			Schema{Fields: []Field{asJSONField(Field{Name: "n1"})}, ForeignKeys: ForeignKeys{Reference: ForeignKeyReference{Fields: []string{"n1"}}}},
		},
	}
	for _, d := range data {
		t.Run(d.Desc, func(t *testing.T) {
			is := is.New(t)
			s, err := Read(strings.NewReader(d.JSON))
			is.NoErr(err)
			is.Equal(s, &d.Schema)
		})
	}
	t.Run("MissingValues", func(t *testing.T) {
		is := is.New(t)
		reader := strings.NewReader(`{"fields":[{"name":"n","type":"integer"}],"missingValues":["na"]}`)
		s, err := Read(reader)
		is.NoErr(err)

		f := s.Fields[0]
		_, ok := f.MissingValues["na"]
		is.True(ok)
	})
}

func TestRead_Error(t *testing.T) {
	data := []struct {
		Desc string
		JSON string
	}{
		{"InvalidSchema", `{"fields":"f1"}`},
		{"EmptyDescriptor", ""},
		{"InvalidPKType", `{"fields":[{"name":"n1"}], "primaryKey":1}`},
		{"InvalidFKFieldsType", `{"fields":[{"name":"n1"}], "foreignKeys":{"fields":1}}`},
		{"InvalidFKReferenceFieldsType", `{"fields":[{"name":"n1"}], "foreignKeys":{"reference":{"fields":1}}}`},
	}
	for _, d := range data {
		t.Run(d.Desc, func(t *testing.T) {
			is := is.New(t)
			_, err := Read(strings.NewReader(d.JSON))
			is.True(err != nil)
		})
	}
}

func TestSchema_Decode(t *testing.T) {
	t.Run("NoImplicitCast", func(t *testing.T) {
		is := is.New(t)
		t1 := struct {
			Name string
			Age  int64
		}{}
		s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}
		is.NoErr(s.Decode([]string{"Foo", "42"}, &t1))
		is.Equal(t1.Name, "Foo")
		is.Equal(t1.Age, int64(42))
	})
	t.Run("StructWithTags", func(t *testing.T) {
		is := is.New(t)
		t1 := struct {
			MyName string `tableheader:"Name"`
			MyAge  int64  `tableheader:"Age"`
		}{}
		s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}
		is.NoErr(s.Decode([]string{"Foo", "42"}, &t1))
		is.Equal(t1.MyName, "Foo")
		is.Equal(t1.MyAge, int64(42))
	})
	t.Run("ImplicitCastToInt", func(t *testing.T) {
		is := is.New(t)
		t1 := struct{ Age int }{}
		s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}
		is.NoErr(s.Decode([]string{"Foo", "42"}, &t1))
		is.Equal(t1.Age, 42)
	})
	t.Run("Error_SchemaFieldAndStructFieldDifferentTypes", func(t *testing.T) {
		is := is.New(t)
		// Field is string and struct is int.
		t1 := struct{ Age int }{}
		s := Schema{Fields: []Field{{Name: "Age", Type: StringType}}}
		is.True(s.Decode([]string{"42"}, &t1) != nil)
	})
	t.Run("Error_NotAPointerToStruct", func(t *testing.T) {
		is := is.New(t)
		t1 := struct{ Age int }{}
		s := Schema{Fields: []Field{{Name: "Name", Type: StringType}}}
		is.True(s.Decode([]string{"Foo", "42"}, t1) != nil)
	})
	t.Run("Error_CellCanNotBeCast", func(t *testing.T) {
		is := is.New(t)
		// Field is string and struct is int.
		t1 := struct{ Age int }{}
		s := Schema{Fields: []Field{{Name: "Age", Type: IntegerType}}}
		is.True(s.Decode([]string{"foo"}, &t1) != nil)
	})
	t.Run("Error_CastToNil", func(t *testing.T) {
		is := is.New(t)
		t1 := &struct{ Age int }{}
		t1 = nil
		s := Schema{Fields: []Field{{Name: "Age", Type: IntegerType}}}
		is.True(s.Decode([]string{"foo"}, &t1) != nil)
	})
}

func TestValidate_SimpleValid(t *testing.T) {
	data := []struct {
		Desc   string
		Schema Schema
	}{
		{"PrimaryKey", Schema{Fields: []Field{{Name: "p"}, {Name: "i"}},
			PrimaryKeys: []string{"p"},
			ForeignKeys: ForeignKeys{
				Fields:    []string{"p"},
				Reference: ForeignKeyReference{Resource: "", Fields: []string{"i"}},
			}},
		},
	}
	for _, d := range data {
		t.Run(d.Desc, func(t *testing.T) {
			is := is.New(t)
			is.NoErr(d.Schema.Validate())
		})
	}
}

func TestValidate_Invalid(t *testing.T) {
	data := []struct {
		Desc   string
		Schema Schema
	}{
		{"MissingName", Schema{Fields: []Field{{Type: IntegerType}}}},
		{"PKNonexistingField", Schema{Fields: []Field{{Name: "n1"}}, PrimaryKeys: []string{"n2"}}},
		{"FKNonexistingField", Schema{Fields: []Field{{Name: "n1"}},
			ForeignKeys: ForeignKeys{Fields: []string{"n2"}},
		}},
		{"InvalidReferences", Schema{Fields: []Field{{Name: "n1"}},
			ForeignKeys: ForeignKeys{
				Fields:    []string{"n1"},
				Reference: ForeignKeyReference{Resource: "", Fields: []string{"n1", "n2"}},
			}},
		},
	}
	for _, d := range data {
		t.Run(d.Desc, func(t *testing.T) {
			is := is.New(t)
			is.True(d.Schema.Validate() != nil)
		})
	}
}

func TestWrite(t *testing.T) {
	is := is.New(t)
	s := Schema{
		Fields:      []Field{{Name: "Foo"}, {Name: "Bar"}},
		PrimaryKeys: []string{"Foo"},
		ForeignKeys: ForeignKeys{Reference: ForeignKeyReference{Fields: []string{"Foo"}}},
	}
	buf := bytes.NewBufferString("")
	is.NoErr(s.Write(buf))

	want := `{
    "fields": [
        {
            "name": "Foo",
            "Constraints": {}
        },
        {
            "name": "Bar",
            "Constraints": {}
        }
    ],
    "primaryKey": [
        "Foo"
    ],
    "foreignKeys": {
        "reference": {
            "fields": [
                "Foo"
            ]
        }
    }
}`

	is.Equal(buf.String(), want)
}

func TestGetField(t *testing.T) {
	t.Run("HasField", func(t *testing.T) {
		is := is.New(t)
		s := Schema{Fields: []Field{{Name: "Foo"}, {Name: "Bar"}}}
		field, pos := s.GetField("Foo")
		is.Equal(pos, 0)
		is.True(field != nil)

		field, pos = s.GetField("Bar")
		is.Equal(pos, 1)
		is.True(field != nil)
	})
	t.Run("DoesNotHaveField", func(t *testing.T) {
		is := is.New(t)
		s := Schema{Fields: []Field{{Name: "Bez"}}}
		field, pos := s.GetField("Foo")
		is.Equal(pos, InvalidPosition)
		is.True(field == nil)
	})
}

func TestHasField(t *testing.T) {
	t.Run("HasField", func(t *testing.T) {
		is := is.New(t)
		s := Schema{Fields: []Field{{Name: "Foo"}, {Name: "Bar"}}}
		is.True(s.HasField("Foo"))
		is.True(s.HasField("Bar"))
	})
	t.Run("DoesNotHaveField", func(t *testing.T) {
		is := is.New(t)
		s := Schema{Fields: []Field{{Name: "Bez"}}}
		is.True(!s.HasField("Bar"))
	})
}

func TestMissingValues(t *testing.T) {
	is := is.New(t)
	s := Schema{
		Fields:        []Field{{Name: "Foo"}},
		MissingValues: []string{"f"},
	}
	row := struct {
		Foo string
	}{}
	s.Decode([]string{"f"}, &row)
	is.Equal(row.Foo, "")
}

type csvRow struct {
	Name string
}

func TestDecodeTable(t *testing.T) {
	data := []struct {
		desc string
		got  []csvRow
	}{
		{"OutEmpty", []csvRow{}},
		{"OutNil", nil},
		{"OutInitialized", []csvRow{{"fooooo"}}},
	}
	for _, d := range data {
		t.Run(d.desc, func(t *testing.T) {
			is := is.New(t)
			tab := table.FromSlices(
				[]string{"Name"},
				[][]string{{"foo"}, {"bar"}})
			s := &Schema{Fields: []Field{{Name: "Name", Type: StringType}}}
			is.NoErr(s.DecodeTable(tab, &d.got))

			want := []csvRow{{"foo"}, {"bar"}}
			is.Equal(want, d.got)
		})
	}
	t.Run("MoarData", func(t *testing.T) {
		is := is.New(t)
		tab := table.FromSlices(
			[]string{"ID", "Age", "Name"},
			[][]string{{"1", "39", "Paul"}, {"2", "23", "Jimmy"}, {"3", "36", "Jane"}, {"4", "28", "Judy"}, {"5", "37", "Iñtërnâtiônàlizætiøn"}})

		type data struct {
			ID   int
			Age  int
			Name string
		}
		s := &Schema{Fields: []Field{{Name: "ID", Type: IntegerType}, {Name: "Age", Type: IntegerType}, {Name: "Name", Type: StringType}}}
		got := []data{}
		is.NoErr(s.DecodeTable(tab, &got))

		want := []data{{1, 39, "Paul"}, {2, 23, "Jimmy"}, {3, 36, "Jane"}, {4, 28, "Judy"}, {5, 37, "Iñtërnâtiônàlizætiøn"}}
		is.Equal(want, got)
	})
	t.Run("EmptyTable", func(t *testing.T) {
		is := is.New(t)
		tab := table.FromSlices([]string{}, [][]string{})
		s := &Schema{Fields: []Field{{Name: "name", Type: StringType}}}
		var got []csvRow
		is.NoErr(s.DecodeTable(tab, &got))
		is.Equal(len(got), 0)
	})
	t.Run("Error_OutNotAPointerToSlice", func(t *testing.T) {
		is := is.New(t)
		tab := table.FromSlices([]string{"name"}, [][]string{{""}})
		s := &Schema{Fields: []Field{{Name: "name", Type: StringType}}}
		is.True(s.DecodeTable(tab, []csvRow{}) != nil)
	})
}

func TestSchema_Encode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		is := is.New(t)
		type rowType struct {
			Name string
			Age  int
		}
		s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}
		got, err := s.Encode(rowType{Name: "Foo", Age: 42})
		is.NoErr(err)

		want := []string{"Foo", "42"}
		is.Equal(want, got)
	})
	t.Run("SuccessWithTags", func(t *testing.T) {
		is := is.New(t)
		type rowType struct {
			MyName string `tableheader:"Name"`
			MyAge  int    `tableheader:"Age"`
		}
		s := Schema{Fields: []Field{{Name: "Name", Type: StringType}, {Name: "Age", Type: IntegerType}}}
		got, err := s.Encode(rowType{MyName: "Foo", MyAge: 42})
		is.NoErr(err)

		want := []string{"Foo", "42"}
		is.Equal(want, got)
	})
	t.Run("Error_Encoding", func(t *testing.T) {
		is := is.New(t)
		type rowType struct {
			Age string
		}
		s := Schema{Fields: []Field{{Name: "Age", Type: IntegerType}}}
		_, err := s.Encode(rowType{Age: "10"})
		is.True(err != nil)
	})
	t.Run("Error_NotStruct", func(t *testing.T) {
		is := is.New(t)
		s := Schema{Fields: []Field{{Name: "name", Type: StringType}}}
		in := "string"
		_, err := s.Encode(in)
		is.True(err != nil)
	})
	t.Run("Error_StructIsNil", func(t *testing.T) {
		is := is.New(t)
		s := Schema{Fields: []Field{{Name: "name", Type: StringType}}}
		var in *csvRow
		_, err := s.Encode(in)
		is.True(err != nil)
	})
}

func TestEncodeTable(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		is := is.New(t)
		people := []struct {
			Name string
		}{{"Foo"}, {"Bar"}, {"Bez"}}
		s := Schema{Fields: []Field{{Name: "Name", Type: StringType}}}
		got, err := s.EncodeTable(people)
		is.NoErr(err)

		want := [][]string{{"Foo"}, {"Bar"}, {"Bez"}}
		is.Equal(want, got)
	})

	t.Run("Error_InputIsNotASlice", func(t *testing.T) {
		is := is.New(t)
		s := Schema{Fields: []Field{{Name: "Name", Type: StringType}}}
		_, err := s.EncodeTable(10)
		is.True(err != nil)
	})
	t.Run("Error_ErrorEncoding", func(t *testing.T) {
		is := is.New(t)
		people := []struct {
			Name string
		}{{"Foo"}}
		s := Schema{Fields: []Field{{Name: "Name", Type: IntegerType}}}
		_, err := s.EncodeTable(people)
		is.True(err != nil)
	})
}
