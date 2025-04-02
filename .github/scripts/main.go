package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func joinWithComma(items []string) string {
	return strings.Join(items, ", ")
}

type TypeParam struct {
	Name     string
	NotFirst bool
}

type Field struct {
	Name string
	Type string
}

type FuncParam struct {
	Name     string
	Type     string
	NotFirst bool
}

type ResVar struct {
	Name string
	Type string
}

type SubscribeBlock struct {
	Future string
	Type   string
	ResVar string
}

type TupleVal struct {
	Name     string
	NotFirst bool
}

type TupleInfo struct {
	TypeParams      []TypeParam
	Fields          []Field
	FuncParams      []FuncParam
	ResVars         []ResVar
	SubscribeBlocks []SubscribeBlock
	TupleVals       []TupleVal
	TupleType       string
}

type TupleData struct {
	Tuples []TupleInfo
}

func main() {
	count := flag.Int("count", 16, "the count of tuple types to generate")
	outputDir := flag.String("outputDir", "../../", "the dir of output file")
	flag.Parse()

	// Prepare template data
	data := TupleData{}
	data.Tuples = make([]TupleInfo, *count-1)

	for i := 2; i <= *count; i++ {
		tuple := &data.Tuples[i-2]

		// Generate type parameters and tuple type
		tuple.TypeParams = make([]TypeParam, i)
		typeNames := make([]string, i)
		for j := 0; j < i; j++ {
			typeName := fmt.Sprintf("T%d", j)
			typeNames[j] = typeName
			tuple.TypeParams[j] = TypeParam{
				Name:     typeName,
				NotFirst: j > 0,
			}
		}
		tuple.TupleType = fmt.Sprintf("Tuple%d[%s]", i, joinWithComma(typeNames))

		// Generate fields
		tuple.Fields = make([]Field, i)
		for j := 0; j < i; j++ {
			tuple.Fields[j] = Field{
				Name: fmt.Sprintf("Val%d", j),
				Type: fmt.Sprintf("T%d", j),
			}
		}

		// Generate function parameters
		tuple.FuncParams = make([]FuncParam, i)
		for j := 0; j < i; j++ {
			tuple.FuncParams[j] = FuncParam{
				Name:     fmt.Sprintf("t%d", j),
				Type:     fmt.Sprintf("T%d", j),
				NotFirst: j > 0,
			}
		}

		// Generate result variables
		tuple.ResVars = make([]ResVar, i)
		for j := 0; j < i; j++ {
			tuple.ResVars[j] = ResVar{
				Name: fmt.Sprintf("res%d", j),
				Type: fmt.Sprintf("T%d", j),
			}
		}

		// Generate subscribe blocks
		tuple.SubscribeBlocks = make([]SubscribeBlock, i)
		for j := 0; j < i; j++ {
			tuple.SubscribeBlocks[j] = SubscribeBlock{
				Future: fmt.Sprintf("t%d", j),
				Type:   fmt.Sprintf("T%d", j),
				ResVar: fmt.Sprintf("res%d", j),
			}
		}

		// Generate tuple values
		tuple.TupleVals = make([]TupleVal, i)
		for j := 0; j < i; j++ {
			tuple.TupleVals[j] = TupleVal{
				Name:     fmt.Sprintf("res%d", j),
				NotFirst: j > 0,
			}
		}
	}

	// Generate tuple.go
	tupleTmpl, err := template.ParseFiles("tuple.tmpl")
	if err != nil {
		fmt.Printf("Error parsing tuple template: %v\n", err)
		os.Exit(1)
	}

	tuplePath := filepath.Join(*outputDir, "tuple.go")
	tuplefile, err := os.Create(tuplePath)
	if err != nil {
		fmt.Printf("Error creating tuple.go: %v\n", err)
		os.Exit(1)
	}
	defer tuplefile.Close()

	if err := tupleTmpl.Execute(tuplefile, data); err != nil {
		fmt.Printf("Error executing tuple template: %v\n", err)
		os.Exit(1)
	}

	// Generate of.go
	ofTmpl, err := template.ParseFiles("of.tmpl")
	if err != nil {
		fmt.Printf("Error parsing of template: %v\n", err)
		os.Exit(1)
	}

	ofPath := filepath.Join(*outputDir, "of.go")
	offile, err := os.Create(ofPath)
	if err != nil {
		fmt.Printf("Error creating of.go: %v\n", err)
		os.Exit(1)
	}
	defer offile.Close()

	if err := ofTmpl.Execute(offile, data); err != nil {
		fmt.Printf("Error executing of template: %v\n", err)
		os.Exit(1)
	}
}
