package templates

var ServiceDotGo = `// Code generated by Micro. DO NOT EDIT.
package {{ .ServiceLowerName }}

import (
	_micro_handle_info "github.com/markity/micro/handleinfo"
	{{- range .ImportPaths }}
	{{ .Alias }} {{ .ImportPath }}
	{{- end }}
)

var serviceMethods = map[string]_micro_handle_info.HandleInfo{
	{{- range .AllMethods }}
	"{{ .RawName }}": {
		Name: "{{ .RawName }}",
		Type: _micro_handle_info.PROTO_TO_PROTO,
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
	_micro_errx "github.com/markity/micro/errx"
	_micro_server "github.com/markity/micro/server"
	_micro_options "github.com/markity/micro/server/options"
	{{- range .ImportPaths }}
	{{ .Alias }} {{ .ImportPath }}
	{{- end }}
)

type {{ .ServiceNameFirstCharUpper }} interface {
	{{- range .AllMethods }}
	{{ .RawName }}(ctx context.Context, req *{{ .ArgStructStr }}) (resp *{{.ResStructStr}}, err _micro_errx.ErrX)
	{{- end}}
}

func NewServer(serviceName string, addrPort string, implementedServer {{ .ServiceNameFirstCharUpper }}, opts ..._micro_options.Option) _micro_server.MicroServer {
	return _micro_server.NewServer(serviceName, addrPort, implementedServer, serviceMethods, opts...)
}

type UnimplementedService struct {}

{{- range .AllMethods }}
func (svc *UnimplementedService) {{ .RawName }}(ctx context.Context, req *{{ .ArgStructStr }}) (resp *{{.ResStructStr}}, err _micro_errx.ErrX) {
	return nil, &_micro_errx.BizError{
		Code: 404,
		Msg: "unimplemented",
		ExtraData: "",
	}
}
{{- end}}
`

var ClientDotGo = `
package echo

import (
	_micro_client "github.com/markity/micro/client"
	_micro_errx "github.com/markity/micro/errx"
	{{- range .ImportPaths }}
	{{ .Alias }} {{ .ImportPath }}
	{{- end }}
)

type {{ .ServiceNameFirstCharUpper }}Client interface {
	{{- range .AllMethods }}
	{{ .RawName }}(req *{{ .ArgStructStr }}) (*{{.ResStructStr}}, _micro_errx.ClientCallError)
	{{- end}}
}

type {{ .ServiceLowerName }}Client struct {
	serviceName string
	cliStub     _micro_client.MicroClient
}

{{- range .AllMethods }}
func (cli *{{ $.ServiceLowerName }}Client) {{ .RawName }}(req *{{ .ArgStructStr }}) (*{{.ResStructStr}}, _micro_errx.ClientCallError) {
	result1Iface, result2 := cli.cliStub.Call("{{ .RawName }}", req)
	if result1Iface != nil {
		return result1Iface.(*{{.ResStructStr}}), result2
	}
	return nil, result2
}
{{- end}}

func NewClient(serviceName string) {{ .ServiceNameFirstCharUpper }}Client {
	return &{{ .ServiceLowerName }}Client{
		serviceName: serviceName,
		cliStub:     _micro_client.NewClient(serviceName, serviceMethods),
	}
}
`
