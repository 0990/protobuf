package netrpc

import (
	"bytes"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"html/template"
	"log"
)

func init(){
	generator.RegisterPlugin(new(netrpc))
}

type netrpc struct{
	gen *generator.Generator
}

func(p *netrpc)Name() string{
	return "netrpc"
}

func(p *netrpc)Init(g *generator.Generator){
	p.gen = g
}

func(p *netrpc)GenerateImports(file *generator.FileDescriptor){
	if len(file.Service)>0{

	}
}

func(p *netrpc)genImportCode(file *generator.FileDescriptor){
	p.gen.P(`import "net.rpc"`)
}

func(p *netrpc)Generate(file *generator.FileDescriptor){
	for _,v:=range file.Service{
		p.genServiceCode(v)
	}
}


const tmplService = `
{{$root := .}}
 
type {{.ServiceName}}Interface interface {
    {{- range $_, $m := .MethodList}}
    {{$m.MethodName}}(*{{$m.InputTypeName}}, *{{$m.OutputTypeName}}) error
    {{- end}}
}
 
func Register{{.ServiceName}}(
    srv *rpc.Server, x {{.ServiceName}}Interface,
) error {
    if err := srv.RegisterName("{{.ServiceName}}", x); err != nil {
        return err
    }
    return nil
}
 
type {{.ServiceName}}Client struct {
    *rpc.Client
}
 
var _ {{.ServiceName}}Interface = (*{{.ServiceName}}Client)(nil)
 
func Dial{{.ServiceName}}(network, address string) (
    *{{.ServiceName}}Client, error,
) {
    c, err := rpc.Dial(network, address)
    if err != nil {
        return nil, err
    }
    return &{{.ServiceName}}Client{Client: c}, nil
}
 
{{range $_, $m := .MethodList}}
func (p *{{$root.ServiceName}}Client) {{$m.MethodName}}(
    in *{{$m.InputTypeName}}, out *{{$m.OutputTypeName}},
) error {
    return p.Client.Call("{{$root.ServiceName}}.{{$m.MethodName}}", in, out)
}
{{end}}
`

func(p *netrpc)genServiceCode(svc *descriptor.ServiceDescriptorProto){
	spec:=p.buildServiceSpec(svc)
	var buf bytes.Buffer
	t:=template.Must(template.New("").Parse(tmplService))
	err:=t.Execute(&buf,spec)
	if err!=nil{
		log.Fatal(err)
	}
	p.gen.P(buf.String())
}

type ServiceSpec struct{
	ServiceName string
	MethodList []ServiceMethodSpec
}

type ServiceMethodSpec struct{
	MethodName string
	InputTypeName string
	OutputTypeName string
}

func(p *netrpc)buildServiceSpec(svc *descriptor.ServiceDescriptorProto)*ServiceSpec{
	spec:=&ServiceSpec{
		ServiceName: generator.CamelCase(svc.GetName()),
		MethodList:  nil,
	}

	for _,m:=range svc.Method{
		spec.MethodList = append(spec.MethodList,ServiceMethodSpec{
			MethodName: generator.CamelCase(m.GetName()),
			InputTypeName: p.gen.TypeName(p.gen.ObjectNamed(m.GetInputType())),
			OutputTypeName: p.gen.TypeName(p.gen.ObjectNamed(m.GetOutputType())),
		})
	}
	return spec
}
