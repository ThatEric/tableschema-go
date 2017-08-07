package schema

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func ExampleInfer() {
	headers := []string{"Person", "Height"}
	table := [][]string{
		[]string{"Foo", "5"},
		[]string{"Bar", "4"},
		[]string{"Bez", "5.5"},
	}
	s, _ := Infer(headers, table)
	fmt.Println("Fields:")
	for _, f := range s.Fields {
		fmt.Printf("{Name:%s Type:%s Format:%s}\n", f.Name, f.Type, f.Format)
	}
	// Output: Fields:
	// {Name:Person Type:string Format:default}
	// {Name:Height Type:integer Format:default}
}

func ExampleInferImplicitCasting() {
	headers := []string{"Person", "Height"}
	table := [][]string{
		[]string{"Foo", "5"},
		[]string{"Bar", "4"},
		[]string{"Bez", "5.5"},
	}
	s, _ := InferImplicitCasting(headers, table)
	fmt.Println("Fields:")
	for _, f := range s.Fields {
		fmt.Printf("{Name:%s Type:%s Format:%s}\n", f.Name, f.Type, f.Format)
	}
	// Output: Fields:
	// {Name:Person Type:string Format:default}
	// {Name:Height Type:number Format:default}
}

func TestInfer_Success(t *testing.T) {
	data := []struct {
		desc    string
		headers []string
		table   [][]string
		want    Schema
	}{
		{"1Cell_Date", []string{"Birthday"}, [][]string{[]string{"1983-10-15"}}, Schema{Fields: []Field{{Name: "Birthday", Type: DateType, Format: defaultFieldFormat}}}},
		{"1Cell_Integer", []string{"Age"}, [][]string{[]string{"10"}}, Schema{Fields: []Field{{Name: "Age", Type: IntegerType, Format: defaultFieldFormat}}}},
		{"1Cell_Number", []string{"Weight"}, [][]string{[]string{"20.2"}}, Schema{Fields: []Field{{Name: "Weight", Type: NumberType, Format: defaultFieldFormat}}}},
		{"1Cell_Boolean", []string{"Foo"}, [][]string{[]string{"0"}}, Schema{Fields: []Field{{Name: "Foo", Type: BooleanType, Format: defaultFieldFormat}}}},
		{"1Cell_Object", []string{"Foo"}, [][]string{[]string{`{"name":"foo"}`}}, Schema{Fields: []Field{{Name: "Foo", Type: ObjectType, Format: defaultFieldFormat}}}},
		{"1Cell_Array", []string{"Foo"}, [][]string{[]string{`["name"]`}}, Schema{Fields: []Field{{Name: "Foo", Type: ArrayType, Format: defaultFieldFormat}}}},
		{"1Cell_String", []string{"Foo"}, [][]string{[]string{"name"}}, Schema{Fields: []Field{{Name: "Foo", Type: StringType, Format: defaultFieldFormat}}}},
		{"1Cell_Time", []string{"Foo"}, [][]string{[]string{"10:15:50"}}, Schema{Fields: []Field{{Name: "Foo", Type: TimeType, Format: defaultFieldFormat}}}},
		{"1Cell_Year", []string{"Year"}, [][]string{[]string{"2017-08"}}, Schema{Fields: []Field{{Name: "Year", Type: YearMonthType, Format: defaultFieldFormat}}}},
		{"ManyCells",
			[]string{"Name", "Age", "Weight", "Bogus", "Boolean", "Boolean1"},
			[][]string{
				[]string{"Foo", "10", "20.2", "1", "1", "1"},
				[]string{"Foo", "10", "30", "1", "1", "1"},
				[]string{"Foo", "10", "30", "Daniel", "1", "2"},
			},
			Schema{Fields: []Field{
				{Name: "Name", Type: StringType, Format: defaultFieldFormat},
				{Name: "Age", Type: IntegerType, Format: defaultFieldFormat},
				{Name: "Weight", Type: IntegerType, Format: defaultFieldFormat},
				{Name: "Bogus", Type: BooleanType, Format: defaultFieldFormat},
				{Name: "Boolean", Type: BooleanType, Format: defaultFieldFormat},
				{Name: "Boolean1", Type: BooleanType, Format: defaultFieldFormat},
			}},
		},
	}
	for _, d := range data {
		t.Run(d.desc, func(t *testing.T) {
			s, err := Infer(d.headers, d.table)
			if err != nil {
				t.Fatalf("want:nil, got:%q", err)
			}
			sort.Sort(s.Fields)
			sort.Sort(d.want.Fields)
			if !reflect.DeepEqual(s, &d.want) {
				t.Errorf("want:%+v, got:%+v", d.want, s)
			}
		})
	}
}

func TestInfer_Error(t *testing.T) {
	data := []struct {
		desc    string
		headers []string
		table   [][]string
	}{
		{"NotATable", []string{}, [][]string{[]string{"1"}}},
	}
	for _, d := range data {
		t.Run(d.desc, func(t *testing.T) {
			_, err := Infer(d.headers, d.table)
			if err == nil {
				t.Fatalf("want:error, got:nil")
			}
		})
	}
}

func TestInferImplicitCasting_Success(t *testing.T) {
	data := []struct {
		desc    string
		headers []string
		table   [][]string
		want    Schema
	}{
		{"1Cell_Date", []string{"Birthday"}, [][]string{[]string{"1983-10-15"}}, Schema{Fields: []Field{{Name: "Birthday", Type: DateType, Format: defaultFieldFormat}}}},
		{"1Cell_Integer", []string{"Age"}, [][]string{[]string{"10"}}, Schema{Fields: []Field{{Name: "Age", Type: IntegerType, Format: defaultFieldFormat}}}},
		{"1Cell_Number", []string{"Weight"}, [][]string{[]string{"20.2"}}, Schema{Fields: []Field{{Name: "Weight", Type: NumberType, Format: defaultFieldFormat}}}},
		{"1Cell_Boolean", []string{"Foo"}, [][]string{[]string{"0"}}, Schema{Fields: []Field{{Name: "Foo", Type: BooleanType, Format: defaultFieldFormat}}}},
		{"1Cell_Object", []string{"Foo"}, [][]string{[]string{`{"name":"foo"}`}}, Schema{Fields: []Field{{Name: "Foo", Type: ObjectType, Format: defaultFieldFormat}}}},
		{"1Cell_Array", []string{"Foo"}, [][]string{[]string{`["name"]`}}, Schema{Fields: []Field{{Name: "Foo", Type: ArrayType, Format: defaultFieldFormat}}}},
		{"1Cell_String", []string{"Foo"}, [][]string{[]string{"name"}}, Schema{Fields: []Field{{Name: "Foo", Type: StringType, Format: defaultFieldFormat}}}},
		{"1Cell_Time", []string{"Foo"}, [][]string{[]string{"10:15:50"}}, Schema{Fields: []Field{{Name: "Foo", Type: TimeType, Format: defaultFieldFormat}}}},
		{"ManyCells",
			[]string{"Name", "Age", "Weight", "Bogus", "Boolean", "Int"},
			[][]string{
				[]string{"Foo", "10", "20.2", "1", "1", "1"},
				[]string{"Foo", "10", "30", "1", "1", "1"},
				[]string{"Foo", "10", "30", "Daniel", "1", "2"},
			},
			Schema{Fields: []Field{
				{Name: "Name", Type: StringType, Format: defaultFieldFormat},
				{Name: "Age", Type: IntegerType, Format: defaultFieldFormat},
				{Name: "Weight", Type: NumberType, Format: defaultFieldFormat},
				{Name: "Bogus", Type: StringType, Format: defaultFieldFormat},
				{Name: "Boolean", Type: BooleanType, Format: defaultFieldFormat},
				{Name: "Int", Type: IntegerType, Format: defaultFieldFormat},
			}},
		},
	}
	for _, d := range data {
		t.Run(d.desc, func(t *testing.T) {
			s, err := InferImplicitCasting(d.headers, d.table)
			if err != nil {
				t.Fatalf("want:nil, got:%q", err)
			}
			sort.Sort(s.Fields)
			sort.Sort(d.want.Fields)
			if !reflect.DeepEqual(s, &d.want) {
				t.Errorf("want:%+v, got:%+v", d.want, s)
			}
		})
	}
}

func TestInferImplicitCasting_Error(t *testing.T) {
	data := []struct {
		desc    string
		headers []string
		table   [][]string
	}{
		{"NotATable", []string{}, [][]string{[]string{"1"}}},
	}
	for _, d := range data {
		t.Run(d.desc, func(t *testing.T) {
			_, err := InferImplicitCasting(d.headers, d.table)
			if err == nil {
				t.Fatalf("want:error, got:nil")
			}
		})
	}
}

var (
	benchmarkHeaders = []string{"Name", "Birthday", "Weight", "Address", "Siblings"}
	benchmarkTable   = [][]string{
		[]string{"Foo", "2015-10-12", "20.2", `{"Street":"Foo", "Number":10, "City":"New York", "State":"NY"}`, `["Foo"]`},
		[]string{"Bar", "2015-10-12", "30", `{"Street":"Foo", "Number":10, "City":"New York", "State":"NY"}`, `["Foo"]`},
		[]string{"Bez", "2015-10-12", "30", `{"Street":"Foo", "Number":10, "City":"New York", "State":"NY"}`, `["Foo"]`},
	}
)

func benchmarkInfer(growthMultiplier int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		Infer(benchmarkHeaders, generateBenchmarkTable(growthMultiplier))
	}
}

func benchmarkInferImplicitCasting(growthMultiplier int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		InferImplicitCasting(benchmarkHeaders, generateBenchmarkTable(growthMultiplier))
	}
}

func generateBenchmarkTable(growthMultiplier int) [][]string {
	var t [][]string
	for i := 0; i < growthMultiplier; i++ {
		t = append(t, benchmarkTable...)
	}
	return t
}

func BenchmarkInferSmall(b *testing.B)                 { benchmarkInfer(1, b) }
func BenchmarkInferMedium(b *testing.B)                { benchmarkInfer(100, b) }
func BenchmarkInferBig(b *testing.B)                   { benchmarkInfer(1000, b) }
func BenchmarkInferImplicitCastingSmall(b *testing.B)  { benchmarkInferImplicitCasting(1, b) }
func BenchmarkInferImplicitCastingMedium(b *testing.B) { benchmarkInferImplicitCasting(100, b) }
func BenchmarkInferImplicitCastingBig(b *testing.B)    { benchmarkInferImplicitCasting(1000, b) }
