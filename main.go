package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ardielle/ardielle-go/rdl"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	pOutdir := flag.String("o", ".", "Output directory")
	flag.String("s", "", "RDL source file")
	flag.Parse()
	data, err := ioutil.ReadAll(os.Stdin)
	if err == nil {
		var schema rdl.Schema
		err = json.Unmarshal(data, &schema)
		if err == nil {
			ExportToEll(&schema, *pOutdir)
			os.Exit(0)
		}
	}
	fmt.Fprintf(os.Stderr, "*** %v\n", err)
	os.Exit(1)
}

func capitalize(text string) string {
	return strings.ToUpper(text[0:1]) + text[1:]
}

func uncapitalize(text string) string {
	return strings.ToLower(text[0:1]) + text[1:]
}

func stringOfChar(count int, b byte) string {
	buf := make([]byte, 0, count)
	for i := 0; i < count; i++ {
		buf = append(buf, b)
	}
	return string(buf)
}

func outputWriter(outdir string, name string, ext string) (*bufio.Writer, *os.File, string, error) {
	sname := "anonymous"
	if strings.HasSuffix(outdir, ext) {
		name = filepath.Base(outdir)
		sname = name[:len(name)-len(ext)]
		outdir = filepath.Dir(outdir)
	}
	if name != "" {
		sname = name
	}
	if outdir == "" {
		return bufio.NewWriter(os.Stdout), nil, sname, nil
	}
	outfile := sname
	if !strings.HasSuffix(outfile, ext) {
		outfile += ext
	}
	path := filepath.Join(outdir, outfile)
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, "", err
	}
	writer := bufio.NewWriter(f)
	return writer, f, sname, nil
}

func ellName(name string) string {
	return strings.ToLower(name)
}

func ExportToEll(schema *rdl.Schema, outdir string) error {
	out, file, _, err := outputWriter(outdir, string(schema.Name), ".ell")
	if err != nil {
		return err
	}
	if file != nil {
		defer file.Close()
	}
	//registry := rdl.NewTypeRegistry(schema)

	fmt.Fprintf(out, ";;\n;; Generated by rdl (https://github.com/ardielle/ardielle-tools/)\n;;\n(use http-util)\n\n")
	handlers := make([]string, 0)

	emitTypes := false
	if emitTypes {
		registry := rdl.NewTypeRegistry(schema)
		if len(schema.Types) > 0 {
			for _, t := range schema.Types {
				tName, _, _ := rdl.TypeInfo(t)
				bt := registry.BaseType(t)
				//(defstruct point x: <number> y: <number>)
				switch bt {
				case rdl.BaseTypeStruct:
					fmt.Fprintf(out, "; (defstruct %s", ellTypeName(tName))
					f := flattenedFields(registry, t)
					for _, f := range f {
						if !f.Optional {
							fmt.Fprintf(out, "\n;    %s: %s", f.Name, ellTypeRef(f.Type))
						} else {
							fmt.Fprintf(out, "\n;    ; %s: (optional %s)", f.Name, ellTypeRef(f.Type))
						}
					}
					fmt.Fprintf(out, ")\n;\n")
				default:
					fmt.Fprintf(out, ";; type: %s\n", ellTypeName(tName))
				}
			}
			fmt.Fprintf(out, "\n")
		}
	}

	for _, resource := range schema.Resources {
		path := make([]string, 0)
		for _, el := range strings.Split(resource.Path, "/") {
			if el != "" {
				if el[0] == '{' && el[len(el)-1] == '}' {
					sym := el[1 : len(el)-1]
					path = append(path, sym+":")
				} else {
					path = append(path, fmt.Sprintf("%q", el))
				}
			}
		}
		spath := strings.Join(path, " ")
		shandler := string(resource.Name)
		if shandler == "" {
			shandler = strings.ToLower(resource.Method) + "-" + strings.ToLower(string(resource.Type))
		} else {
			shandler = strings.Replace(shandler, "_", "-", -1)
		}
		args := make([]string, 0)
		for _, in := range resource.Inputs {
			args = append(args, string(in.Name))
		}
		sargs := strings.Join(args, " ")
		fmt.Fprintf(out, "(defn %s (%s)\n   (http-fail 501 \"Not Implemented\"))\n\n", shandler, sargs)
		handler := fmt.Sprintf("(handler %s %q %s)", shandler, resource.Method, spath)
		handlers = append(handlers, handler)
	}

	fmt.Fprintf(out, "(http-serve 8080")
	for _, handler := range handlers {
		fmt.Fprintf(out, "\n            %s", handler)
	}
	fmt.Fprintf(out, ")\n")

	out.Flush()
	return nil
}

func ellTypeName(tName rdl.TypeName) string {
	return strings.ToLower(string(tName))
}

func ellTypeRef(tName rdl.TypeRef) string {
	return "<" + strings.ToLower(string(tName)) + ">"
}

func addFields(reg rdl.TypeRegistry, dst []*rdl.StructFieldDef, t *rdl.Type) []*rdl.StructFieldDef {
	switch t.Variant {
	case rdl.TypeVariantStructTypeDef:
		st := t.StructTypeDef
		if st.Type != "Struct" {
			dst = addFields(reg, dst, reg.FindType(st.Type))
		}
		for _, f := range st.Fields {
			dst = append(dst, f)
		}
	}
	return dst
}

func flattenedFields(reg rdl.TypeRegistry, t *rdl.Type) []*rdl.StructFieldDef {
	return addFields(reg, make([]*rdl.StructFieldDef, 0), t)
}
