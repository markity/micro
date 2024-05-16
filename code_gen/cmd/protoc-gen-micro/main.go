package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/markity/micro/code_gen/cmd/protoc-gen-micro/templates"
	"github.com/markity/micro/code_gen/util"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	// debug output
	debugInfo, err := os.Create("someinfo")
	if err != nil {
		panic(err)
	}

	// read input protobuf information
	input, err := util.ReadInput()
	if err != nil {
		// unreachable
		panic(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		// unreachable
		panic(err)
	}

	_, _, ok := util.SearchGoMod(cwd)
	if !ok {
		panic("no go mod")
	}

	// modify go package
	req := &pluginpb.CodeGeneratorRequest{}
	err = proto.Unmarshal(input, req)
	if err != nil {
		panic(err)
	}

	genPath := ""
	protoPath := ""
	sps := strings.Split(req.GetParameter(), ",")
	if len(sps) != 2 {
		panic("unexpected")
	}
	for _, v := range sps {
		kvs := strings.SplitN(v, "=", 2)
		if len(kvs) != 2 {
			panic("unexpected")
		}
		switch kvs[0] {
		case "gen_path":
			genPath = kvs[1]
		case "proto_path":
			protoPath = kvs[1]
		default:
			panic("unexpected")
		}
	}
	if genPath == "" || protoPath == "" {
		panic("unexpected")
	}
	debugInfo.WriteString(genPath + "\n")
	debugInfo.WriteString(protoPath + "\n")

	modName, modPath, ok := util.SearchGoMod(cwd)
	if !ok {
		panic("no go mod")
	}

	for _, f := range req.ProtoFile {
		rel, err := filepath.Rel(modPath, filepath.Join(genPath, *f.Name))
		if err != nil {
			panic(err)
		}
		newGoPackage := filepath.Dir((filepath.Join(modName, rel)))

		f.Options.GoPackage = &newGoPackage
		debugInfo.WriteString("\n" + newGoPackage + "\n")
	}

	options := protogen.Options{}
	plugin, err := options.New(req)
	if err != nil {
		panic(err)
	}

	idl := plugin.Request.FileToGenerate[0]
	for _, v := range plugin.Files {
		v.GeneratedFilenamePrefix = *v.Proto.Name
		debugInfo.WriteString("kk" + v.GeneratedFilenamePrefix + "\n")
		gengo.GenerateFile(plugin, v)
		if *v.Proto.Name == idl {
			if len(v.Services) == 0 {
				plugin.Error(errors.New("no service found"))
			}

			dir, _ := filepath.Split(v.GeneratedFilenamePrefix)
			for _, service := range v.Services {
				lowerServiceName := strings.ToLower(service.GoName)
				genServiceDotGo := filepath.Join(dir, lowerServiceName, "service.go")
				genServerDotGo := filepath.Join(dir, lowerServiceName, "server.go")
				genClientDotGo := filepath.Join(dir, lowerServiceName, "client.go")

				protoImport, err := filepath.Rel(modPath, filepath.Join(genPath, filepath.Dir(service.Location.SourceFile)))
				if err != nil {
					panic(err)
				}
				protoImport = filepath.Join(modName, protoImport)

				serviceFileInput := struct {
					ServiceLowerName string
					ProtoImportPath  string
					AllMethods       []struct {
						RawName       string
						ArgStructName string
						ResStructName string
					}
				}{
					ServiceLowerName: lowerServiceName,
					ProtoImportPath:  protoImport,
					AllMethods: make([]struct {
						RawName       string
						ArgStructName string
						ResStructName string
					}, 0),
				}
				for _, method := range service.Methods {
					serviceFileInput.AllMethods = append(serviceFileInput.AllMethods, struct {
						RawName       string
						ArgStructName string
						ResStructName string
					}{
						RawName:       method.GoName,
						ArgStructName: method.Input.GoIdent.GoName,
						ResStructName: method.Output.GoIdent.GoName,
					})
				}
				serviceFile := plugin.NewGeneratedFile(genServiceDotGo, "")
				tmp, err := template.New("").Parse(templates.ServiceDotGo)
				if err != nil {
					panic(err)
				}
				tmp.Execute(serviceFile, serviceFileInput)
				serverFile := plugin.NewGeneratedFile(genServerDotGo, "")
				serverFile.P("package " + lowerServiceName)
				clientFile := plugin.NewGeneratedFile(genClientDotGo, "")
				clientFile.P("package " + lowerServiceName)
			}
		}
	}

	resp := plugin.Response()
	out, err := proto.Marshal(resp)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stdout.Write(out); err != nil {
		panic(err)
	}
}
