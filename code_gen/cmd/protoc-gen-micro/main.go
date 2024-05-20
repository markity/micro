package main

import (
	"errors"
	"fmt"
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
				// 为每个service生成一个文件夹, 生成service.go, client.go, server.go
				lowerServiceName := strings.ToLower(service.GoName)
				genServiceDotGo := filepath.Join(dir, lowerServiceName, "service.go")
				genServerDotGo := filepath.Join(dir, lowerServiceName, "server.go")
				genClientDotGo := filepath.Join(dir, lowerServiceName, "client.go")

				debugInfo.WriteString(service.Methods[0].Input.GoIdent.GoImportPath.String() + "2\n")
				idx := 1
				protoIndexToImportPath := map[int]string{}
				importPathToAlias := map[string]string{}
				for _, v := range service.Methods {
					protoIndexToImportPath[v.Input.Desc.Index()] = v.Input.GoIdent.GoImportPath.String()
					protoIndexToImportPath[v.Output.Desc.Index()] = v.Output.GoIdent.GoImportPath.String()
					if _, ok := importPathToAlias[v.Input.GoIdent.GoImportPath.String()]; !ok {
						importPathToAlias[v.Input.GoIdent.GoImportPath.String()] = fmt.Sprintf("_proto_%v", idx)
						idx++
					}
					if _, ok := importPathToAlias[v.Output.GoIdent.GoImportPath.String()]; !ok {
						importPathToAlias[v.Output.GoIdent.GoImportPath.String()] = fmt.Sprintf("_proto_%v", idx)
						idx++
					}
				}

				importPaths := make([]struct {
					Alias      string
					ImportPath string
				}, 0)
				for k, v := range importPathToAlias {
					importPaths = append(importPaths, struct {
						Alias      string
						ImportPath string
					}{
						Alias:      v,
						ImportPath: k,
					})
				}

				serviceFileInput := struct {
					ServiceLowerName string
					ImportPaths      []struct {
						Alias      string
						ImportPath string
					}
					AllMethods []struct {
						RawName      string
						ArgStructStr string
						ResStructStr string
					}
				}{
					ServiceLowerName: lowerServiceName,
					ImportPaths:      importPaths,
					AllMethods: make([]struct {
						RawName      string
						ArgStructStr string
						ResStructStr string
					}, 0),
				}
				for _, method := range service.Methods {
					serviceFileInput.AllMethods = append(serviceFileInput.AllMethods, struct {
						RawName      string
						ArgStructStr string
						ResStructStr string
					}{
						RawName:      method.GoName,
						ArgStructStr: importPathToAlias[protoIndexToImportPath[method.Input.Desc.Index()]] + "." + method.Input.GoIdent.GoName,
						ResStructStr: importPathToAlias[protoIndexToImportPath[method.Output.Desc.Index()]] + "." + method.Output.GoIdent.GoName,
					})
				}
				serviceFile := plugin.NewGeneratedFile(genServiceDotGo, "")
				tmp, err := template.New("").Parse(templates.ServiceDotGo)
				if err != nil {
					panic(err)
				}
				tmp.Execute(serviceFile, serviceFileInput)

				serverFile := plugin.NewGeneratedFile(genServerDotGo, "")
				serverFileInput := struct {
					ServiceLowerName string
					ServiceName      string
					AllMethods       []struct {
						RawName      string
						ArgStructStr string
						ResStructStr string
					}
					ImportPaths []struct {
						Alias      string
						ImportPath string
					}
					ServiceNameFirstCharUpper string
				}{
					ServiceLowerName:          lowerServiceName,
					ServiceName:               service.GoName,
					AllMethods:                serviceFileInput.AllMethods,
					ImportPaths:               importPaths,
					ServiceNameFirstCharUpper: strings.ToUpper(service.GoName)[0:1] + service.GoName[1:],
				}
				tmpServer, err := template.New("").Parse(templates.ServerDotGo)
				if err != nil {
					panic(err)
				}
				tmpServer.Execute(serverFile, serverFileInput)
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
