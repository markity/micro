package templates

var ServiceDotGo = `// Code generated by Micro. DO NOT EDIT.
package {{ .ServiceLowerName }}

import (
	_micro_service_info "github.com/markity/micro/serviceinfo"
	{{- range .ImportPaths }}
	{{ .Alias }} {{ .ImportPath }}
	{{- end }}
)

var serviceMethods = map[string]_micro_service_info.HandleInfo{
	{{- range .AllMethods }}
	"{{ .RawName }}": {
		Name: "{{ .RawName }}",
		Type: _micro_service_info.PROTO_TO_PROTO,
		Request: &{{ .ArgStructStr }}{},
		Response: &{{ .ResStructStr }}{},
	},
	{{- end}}
}
`

var ServerDotGo = `
// Code generated by Micro. DO NOT EDIT.
package {{ .ServiceLowerName }}

import (
	"context"
	_micro_errcode "github.com/markity/micro/serviceinfo/errcode"
	{{- range .ImportPaths }}
	{{ .Alias }} {{ .ImportPath }}
	{{- end }}
)

type UnimplementedService struct {
	
}

{{- range .AllMethods }}
func (svc *UnimplementedService) {{ .RawName }}(ctx context.Context, req *{{ .ArgStructStr }}) (resp *{{.ResStructStr}}, errcode *_micro_errcode.ErrCode) {
	return nil
}
{{- end}}
`
