package openmock

// Pulled from gripmock
// https://github.com/tokopedia/gripmock

import (
	"io"
	"os"
	"strings"
	"text/template"
)

type generatorParam struct {
	Services []Service
	GrpcAddr string
	PbPath   string
}

type Options struct {
	writer   io.Writer
	grpcAddr string
	pbPath   string
}

func GenerateServer(services []Service, opt *Options) error {
	param := generatorParam{
		Services: services,
		GrpcAddr: opt.grpcAddr,
		PbPath:   opt.pbPath,
	}

	if opt == nil {
		opt = &Options{}
	}

	if opt.writer == nil {
		opt.writer = os.Stdout
	}

	tmpl := template.New("server.tmpl").Funcs(template.FuncMap{
		"Title": strings.Title,
	})
	tmpl, err := tmpl.Parse(SERVER_TEMPLATE)
	if err != nil {
		return err
	}

	return tmpl.Execute(opt.writer, param)
}

const SERVER_TEMPLATE = `// DO NOT EDIT. This file is autogenerated by OpenMock 
package openmock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	TCP_ADDRESS  = "{{.GrpcAddr}}"
)

{{ range .Services }}
{{ template "services" . }}
{{ end }}

func main() {
	lis, err := net.Listen("tcp", TCP_ADDRESS)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	{{ range .Services }}
	{{ template "register_services" . }}
	{{ end }}

	reflection.Register(s)
	fmt.Println("Serving gRPC on tcp://" + TCP_ADDRESS)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

{{ template "find_mock" }}

{{ define "services" }}
type {{.Name}} struct{}

{{ template "methods" .}}
{{ end }}

{{ define "methods" }}
{{ $serviceName := .Name }}
{{ range .Methods}}
{{ $methodName := .Name | Title }}
func (s *{{$serviceName}}) {{$methodName}}(ctx context.Context, in *{{.Input}}) (*{{.Output}},error){
	out := &{{.Output}}{}
	err := findStub("{{$serviceName}}", "{{$methodName}}", in, out)
	return out, err
}
{{end}}
{{end}}

{{ define "register_services" }}
	Register{{.Name}}Server(s, &{{.Name}}{})
{{ end }}

{{ define "find_mock" }}
type payload struct {
	Service string      ` + "`json:\"service\"`" + `
	Method  string      ` + "`json:\"method\"`" + `
	Data    interface{} ` + "`json:\"data\"`" + `
}

type response struct {
	Data  interface{} ` + "`json:\"data\"`" + `
	Error string      ` + "`json:\"error\"`" + `
}

func findMock(service, method string, in, out interface{}) error {
	fmt.Println("findMock Not Implemented. TODO: read om.repo.GRPCMocks")
}
{{ end }}`