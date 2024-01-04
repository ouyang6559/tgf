package util

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"
)

//***************************************************
//@Link  https://github.com/thkhxm/tgf
//@Link  https://gitee.com/timgame/tgf
//@QQ群 7400585
//author tim.huang<thkhxm@gmail.com>
//@Description
//2023/4/27
//***************************************************

var autoGenerateAPICodePath = ""
var autoGenerateAPICSCodePath = ""
var autoGenerateAPICSCodeNamespace = ""
var autoGenerateFileNameSuffix = "api"

func SetAutoGenerateAPICodePath(path string) {
	var err error
	autoGenerateAPICodePath, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	fmt.Printf("设置api代码自动生成路径为 %v", autoGenerateAPICodePath)
}
func SetGenerateFileNameSuffix(suffix string) {
	autoGenerateFileNameSuffix = suffix
	fmt.Printf("设置文件后缀为 %v", autoGenerateFileNameSuffix)
}
func SetAutoGenerateAPICSCode(path, namespace string) {
	var err error
	autoGenerateAPICSCodePath, err = filepath.Abs(path)
	autoGenerateAPICSCodeNamespace = namespace
	if err != nil {
		panic(err)
	}
	fmt.Printf("设置C# api代码自动生成路径为 %v", autoGenerateAPICSCodePath)
}

func GeneratorRPC[T any](moduleName, version, packageName, dir string) {
	var t T
	v := reflect.ValueOf(&t)
	ty := v.Type().Elem()
	s := make([]struct {
		Args       string
		Reply      string
		MethodName string
	}, 0)
	a := struct {
		PackageImports []string
		Apis           []struct {
			Args       string
			Reply      string
			MethodName string
		}
	}{}
	tt := make(map[string]bool)
	tt["github.com/thkhxm/tgf/rpc"] = true
	for i := 0; i < ty.NumMethod(); i++ {
		m := ty.Method(i)
		// 遍历方法的参数列表
		for j := 0; j < m.Type.NumIn(); j++ {
			// 获取参数类型对象
			argType := m.Type.In(j)
			pkg := argType.PkgPath()
			if argType.Kind() == reflect.Pointer {
				pkg = argType.Elem().PkgPath()
				tt[pkg] = true
			}
		}
		d := struct {
			Args       string
			Reply      string
			MethodName string
		}{Args: m.Type.In(1).String(), Reply: m.Type.In(2).String(), MethodName: m.Name}

		var r = regexp.MustCompile("[A-Za-z0-9_]+\\.[A-Za-z0-9_]+\\[(.*)\\]")
		match := r.FindStringSubmatch(d.Args)
		if len(match) > 1 {
			pointIndex := strings.LastIndex(match[1], ".")
			pk := match[1][1:pointIndex]
			l := strings.LastIndex(pk, "/")
			d.Args = "*" + pk[l+1:] + match[1][pointIndex:]
			tt[pk] = true
		}

		match = r.FindStringSubmatch(d.Reply)
		if len(match) > 1 {
			pointIndex := strings.LastIndex(match[1], ".")
			pk := match[1][1:pointIndex]
			l := strings.LastIndex(pk, "/")
			d.Reply = "*" + pk[l+1:] + match[1][pointIndex:]
			tt[pk] = true
		}
		s = append(s, d)
	}
	pi := make([]string, 0)
	for k, _ := range tt {
		pi = append(pi, k)
	}
	a.Apis = s
	a.PackageImports = pi

	tpl := fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

package %v

import (
{{range .PackageImports}}
"{{.}}"
{{end}}
)
var %vService = &rpc.Module{Name: "%v", Version: "%v"}

var (
	{{range .Apis}}
	{{.MethodName}} = rpc.ServiceAPI[{{.Args}}, {{.Reply}}]{
		ModuleName: %vService.Name,
		Name:       "{{.MethodName}}",
		MessageType: %vService.Name+"."+"{{.MethodName}}",
	}
	{{end}}
)

`, time.Now().String(), packageName, moduleName, moduleName, version, moduleName, moduleName)
	tm := template.New("apiStruct")
	tp, _ := tm.Parse(tpl)
	fileName := strings.ToLower(moduleName) + "_" + autoGenerateFileNameSuffix + ".go"
	path := autoGenerateAPICodePath
	if dir != "" {
		path = path + string(filepath.Separator) + dir
	}
	// 创建路径中的所有必要的目录
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(path + string(filepath.Separator) + fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tp.Execute(file, a)
}

type CSStruct struct {
	PackageImports []string
	Apis           []struct {
		Args            string
		Reply           string
		MethodName      string
		ModuleName      string
		ModuleNameUpper string
	}
	PushServices []string
}

var csStructCache = &CSStruct{}

// GeneratorAPI
// @Description: 生成api文件
// @param ref
func GeneratorAPI[T any](moduleName, version, packageName string, pushServices ...string) {
	var t T
	v := reflect.ValueOf(&t)
	ty := v.Type().Elem()

	tt := make(map[string]bool)
	tt["github.com/thkhxm/tgf/rpc"] = true
	goStructCache := &CSStruct{}
	for i := 0; i < ty.NumMethod(); i++ {
		m := ty.Method(i)
		// 遍历方法的参数列表
		for j := 0; j < m.Type.NumIn(); j++ {
			// 获取参数类型对象
			argType := m.Type.In(j)
			pkg := argType.PkgPath()
			if argType.Kind() == reflect.Pointer {
				pkg = argType.Elem().PkgPath()
				tt[pkg] = true
			}
		}
		d := struct {
			Args            string
			Reply           string
			MethodName      string
			ModuleName      string
			ModuleNameUpper string
		}{Args: m.Type.In(1).String(), Reply: m.Type.In(2).String(), MethodName: m.Name, ModuleName: moduleName, ModuleNameUpper: cases.Title(language.English).String(moduleName)}

		var r = regexp.MustCompile("[A-Za-z0-9_]+\\.[A-Za-z0-9_]+\\[(.*)\\]")
		match := r.FindStringSubmatch(d.Args)
		if len(match) > 1 {
			pointIndex := strings.LastIndex(match[1], ".")
			pk := match[1][1:pointIndex]
			l := strings.LastIndex(pk, "/")
			d.Args = "*" + pk[l+1:] + match[1][pointIndex:]
			tt[pk] = true
		}

		match = r.FindStringSubmatch(d.Reply)
		if len(match) > 1 {
			pointIndex := strings.LastIndex(match[1], ".")
			pk := match[1][1:pointIndex]
			l := strings.LastIndex(pk, "/")
			d.Reply = "*" + pk[l+1:] + match[1][pointIndex:]
			tt[pk] = true
		}
		csStructCache.Apis = append(csStructCache.Apis, d)
		goStructCache.Apis = append(goStructCache.Apis, d)
	}
	pi := make([]string, 0)
	for k, _ := range tt {
		pi = append(pi, k)
	}

	goStructCache.PackageImports = pi
	goStructCache.PushServices = pushServices
	csStructCache.PushServices = append(csStructCache.PushServices, pushServices...)
	tpl := fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

package %v

import (
{{range .PackageImports}}
"{{.}}"
{{end}}
)
var %vService = &rpc.Module{Name: "%v", Version: "%v"}

var (
	{{range .Apis}}
	{{.ModuleNameUpper}}_{{.MethodName}} = rpc.ServiceAPI[{{.Args}}, {{.Reply}}]{
		ModuleName: %vService.Name,
		Name:       "{{.MethodName}}",
		MessageType: %vService.Name+"."+"{{.MethodName}}",
	}
	{{end}}
)
const(
	{{range .PushServices}}
	{{.}} = "{{.}}" 
	{{end}}
)

`, time.Now().String(), packageName, moduleName, moduleName, version, moduleName, moduleName)
	tm := template.New("apiStruct")
	tp, _ := tm.Parse(tpl)
	fileName := strings.ToLower(moduleName) + "_" + autoGenerateFileNameSuffix + ".go"
	// 创建路径中的所有必要的目录
	err := os.MkdirAll(autoGenerateAPICodePath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(autoGenerateAPICodePath + string(filepath.Separator) + fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tp.Execute(file, goStructCache)
	//tp.Execute(os.Stdout, a)
}

func GenerateCSApiService() {
	if autoGenerateAPICSCodeNamespace != "" {

		tplCS := fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

using AOT;

namespace %v
{
    public struct ServerApi
    {
	{{range .Apis}}
 		public static readonly Api {{.ModuleNameUpper}}_{{.MethodName}} = new("{{.ModuleName}}","{{.MethodName}}");
	{{end}}
	{{range .PushServices}}
		public static readonly string {{.}} = "{{.}}"; 
	{{end}}
	}
	
}
`, time.Now().String(), autoGenerateAPICSCodeNamespace)
		tmCS := template.New("apiCSStruct")
		tpCS, _ := tmCS.Parse(tplCS)
		fileCSName := "ServerApi.cs"
		err := os.MkdirAll(autoGenerateAPICodePath, os.ModePerm)
		if err != nil {
			panic(err)
		}
		fileCS, errCS := os.Create(autoGenerateAPICSCodePath + string(filepath.Separator) + fileCSName)
		if errCS != nil {
			panic(errCS)
		}
		defer fileCS.Close()
		tpCS.Execute(fileCS, csStructCache)
	}
}
