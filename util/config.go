package util

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
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
//2023/4/10
//***************************************************

var (
	excelToJsonPath       = make([]string, 0)
	excelToClientJsonPath = make([]string, 0)
	excelToGoPath         = ""
	excelToUnityPath      = ""
	excelPath             = ""
	goPackage             = "conf"
	unityPackage          = "HotFix.Config"
	fileExt               = ".xlsx"
)

// ExcelExport
// @Description: Excel导出json文件
func ExcelExport() {
	fmt.Println("---------------start export------------------")
	fmt.Println("")
	fmt.Println("")
	files := GetFileList(excelPath, fileExt)
	structs := make([]*configStruct, 0)
	for _, file := range files {
		d := parseFile(file)
		if d != nil {
			structs = append(structs, d...)
		}
	}
	//
	if excelToGoPath != "" {
		toGolang(structs)
	}
	//
	if excelToUnityPath != "" {
		toUnity(structs)
	}

	fmt.Println("")
	fmt.Println("")
	fmt.Println("---------------end export-------------------")

}

// SetExcelToJsonPath
// @Description: 设置Excel导出Json地址,可以追加多个输出地址
// @param path
func SetExcelToJsonPath(path string) {
	p, _ := filepath.Abs(path)
	excelToJsonPath = append(excelToJsonPath, p)
	fmt.Println("set excel to json path", excelToJsonPath)
}

// SetExcelToClientJsonPath
// @Description: 设置Excel导出客户端Json地址,可以追加多个输出地址
// @param path
func SetExcelToClientJsonPath(path string) {
	p, _ := filepath.Abs(path)
	excelToClientJsonPath = append(excelToClientJsonPath, p)
	fmt.Println("set excel to json path", excelToClientJsonPath)
}

// SetExcelToGoPath
// @Description: 设置Excel导出Go地址
// @param path
func SetExcelToGoPath(path string) {
	excelToGoPath, _ = filepath.Abs(path)
	fmt.Println("set excel to go path", excelToGoPath)
}

// SetExcelToUnityPath
// @Description: 设置Excel导出Unity地址
// @param path
func SetExcelToUnityPath(path string) {
	excelToUnityPath, _ = filepath.Abs(path)
	fmt.Println("set excel to unity path", excelToUnityPath)
}

// SetExcelToUnityNamespace
// @Description: 设置excel导出到unity的命名空间,默认不设置的话为 HotFix.Config
// @param namespace
func SetExcelToUnityNamespace(namespace string) {
	unityPackage = namespace
	fmt.Println("set excel to unity namespace", unityPackage)
}

// SetExcelPath
// @Description: 设置Excel文件所在路径
func SetExcelPath(path string) {
	excelPath, _ = filepath.Abs(path)
	fmt.Println("set excel file path", excelPath)
}

// to golang
func toGolang(metalist []*configStruct) {
	tpl := fmt.Sprintf(`
//Auto Generated by tgf util,DO NOT EDIT.
//created at %v

package %v
		{{range .}}
type {{.StructName}}Conf struct {
		{{range .Fields}}{{if contains .Region "s"}}
		//{{.Des}}
		{{.Key}}	{{.Typ}}{{end}}
		{{end}}
}
{{end}}`, time.Now().String(), goPackage)
	t := template.New("ConfigStruct").Funcs(template.FuncMap{"contains": strings.Contains})
	tp, e := t.Parse(tpl)
	if e != nil {
		panic(e)
	}
	// 创建路径中的所有必要的目录
	err := os.MkdirAll(excelToGoPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(excelToGoPath + string(filepath.Separator) + "conf_struct.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tp.Execute(file, metalist)
}

func toUnity(metalist []*configStruct) {
	tpl := fmt.Sprintf(`
//Auto Generated by tgf util,DO NOT EDIT.
//created at %v

using System.Collections.Generic;

namespace %v
{


{{range .}}
	public class {{.StructName}}Conf {
		{{range .Fields}}{{if contains .Region "c"}}
		//{{.Des}}
		public {{.UnityTyp}} {{.Key}} { get; set; }{{end}}
		{{end}}
	}
{{end}}
}
`, time.Now().String(), unityPackage)
	t := template.New("UnityConfigStruct").Funcs(template.FuncMap{"contains": strings.Contains})
	tp, _ := t.Parse(tpl)
	// 创建路径中的所有必要的目录
	err := os.MkdirAll(excelToUnityPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(excelToUnityPath + string(filepath.Separator) + "AllConfig.cs")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tp.Execute(file, metalist)
}

//

type configStruct struct {
	StructName string
	Fields     []*meta
	Version    string
}

type meta struct {
	Key      string
	Idx      int
	Typ      string
	UnityTyp string
	Des      string
	Region   string
}

type rowdata []interface{}

var C, S = "c", "s"

func parseFile(file string) []*configStruct {
	// Ignoring temporary Excel files
	if strings.HasPrefix(filepath.Base(file), "~$") {
		fmt.Println("Skipping temporary file: [", file, "]")
		return nil
	}

	fmt.Println("excel file [", file, "]")
	fileReader, err := os.OpenFile(file, os.O_RDONLY, 0666)
	xlsx, err := excelize.OpenReader(fileReader)
	if err != nil {
		panic(err.Error())
	}
	sheets := xlsx.GetSheetList()

	rs := make([]*configStruct, 0, len(sheets))

	for _, s := range sheets {
		//如果sheet名字以#开头则忽略
		if strings.HasPrefix(s, "#") {
			continue
		}
		rows, err := xlsx.GetRows(s)
		if err != nil {
			return nil
		}
		if len(rows) < 5 {
			return nil
		}

		colNum := len(rows[1])
		metaList := make([]*meta, 0, colNum)
		dataList := make([]rowdata, 0, len(rows)-4)
		version := ""
		for line, row := range rows {
			switch line {
			case 0: // sheet 名
				version = row[0]
			case 1: // col name
				for idx, colname := range row {
					metaList = append(metaList, &meta{Key: colname, Idx: idx})
				}
			case 2: // data type
				for idx, typ := range row {
					metaList[idx].Typ = typ
					metaList[idx].UnityTyp = convertToUnityFieldType(typ)
				}
			case 3: //Region
				for idx, region := range row {
					metaList[idx].Region = strings.ToLower(region)
				}
			case 4: // desc
				for idx, des := range row {
					metaList[idx].Des = des
				}

			default: //>= 5 row data
				data := make(rowdata, colNum)
				for k := 0; k < colNum; k++ {
					if k < len(row) {
						data[k] = row[k]
					}
				}
				dataList = append(dataList, data)
			}
		}
		jsonFile := fmt.Sprintf("%s.json", s)
		if len(excelToJsonPath) > 0 {
			for _, p := range excelToJsonPath {
				// 创建路径中的所有必要的目录
				err := os.MkdirAll(p, os.ModePerm)
				if err != nil {
					panic(err)
				}
				err = output(p, jsonFile, toJson(dataList, metaList, S))
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		if len(excelToClientJsonPath) > 0 {
			for _, p := range excelToClientJsonPath {
				// 创建路径中的所有必要的目录
				err := os.MkdirAll(p, os.ModePerm)
				if err != nil {
					panic(err)
				}
				err = output(p, jsonFile, toJson(dataList, metaList, C))
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		result := &configStruct{}
		result.Fields = metaList
		result.StructName = s
		result.Version = version
		rs = append(rs, result)
		fmt.Println("excel export : json file", jsonFile, "golang struct :", s+"Conf", "[", version, "]")
	}
	return rs
}

func convertToUnityFieldType(t string) string {
	switch t {
	case "int32", "uint32", "int8", "uint8":
		return "int"
	case "[]int32":
		return "List<int>"
	case "int64", "uint64":
		return "long"
	case "[]int64":
		return "List<long>"
	case "[]string":
		return "List<string>"
	default:
		return t
	}
}

const (
	fileType_string      = "string"
	fileType_time        = "time"
	fileType_arrayInt32  = "[]int32"
	fileType_arrayString = "[]string"
)

var intRegex, _ = regexp.Compile(".*\\[.*\\].*int.*")

func toJson(datarows []rowdata, metalist []*meta, region string) string {
	ret := "["
	for _, row := range datarows {
		ret += "\n\t{"
		for idx, meta := range metalist {
			if strings.Index(meta.Region, region) < 0 {
				continue
			}
			ret += fmt.Sprintf("\n\t\t\"%s\":", meta.Key)
			switch meta.Typ {
			case fileType_time:
				fallthrough
			case fileType_string:
				if row[idx] == nil {
					ret += "\"\""
				} else {
					ret += fmt.Sprintf("\"%s\"", row[idx])
				}
			case fileType_arrayString:
				if row[idx] == nil || row[idx] == "" {
					ret += "[]"
				} else {
					ret += fmt.Sprintf("%s", convertToStringSlice(row[idx].(string)))
				}
			default:
				if intRegex.MatchString(meta.Typ) {
					if row[idx] == nil || row[idx] == "" {
						ret += "[]"
					} else {
						ret += fmt.Sprintf("[%s]", row[idx])
					}
					break
				}
				if row[idx] == nil || row[idx] == "" {
					ret += "0"
				} else {
					ret += fmt.Sprintf("%s", row[idx])
				}
			}
			ret += ","
		}
		ret = ret[:len(ret)-1]
		ret += "\n\t},"
	}
	ret = ret[:len(ret)-1]
	ret += "\n]"
	return ret
}

func convertToStringSlice(input string) string {
	splitInput := strings.Split(input, ",")

	// Convert the string array to json
	jsonData, err := json.Marshal(splitInput)
	if err != nil {
		fmt.Sprintf("string array error %v", err)
	}
	return string(jsonData)
}

func output(path, filename, str string) error {

	f, err := os.OpenFile(path+string(filepath.Separator)+filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		return err
	}

	return nil
}

type TemplateKeyValueData struct {
	FieldName interface{}
	Values    interface{}
	Other     interface{}
}

func JsonToKeyValueGoFile(packageName, fileName, outPath, fieldType string, data []TemplateKeyValueData) {

	tpl := fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

package %v
const(
	{{range .}}
	{{.FieldName}} = "{{.Values}}"
    {{end}}
)
`, time.Now().String(), packageName)
	if "string" != fieldType {
		tpl = fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

package %v
const(
	{{range .}}
	{{.FieldName}} %v = {{.Values}}
    {{end}}
)
`, time.Now().String(), packageName, fieldType)
	}
	t := template.New("KeyValueStruct")
	tp, _ := t.Parse(tpl)
	s, _ := filepath.Abs(outPath)
	// 创建路径中的所有必要的目录
	err := os.MkdirAll(s, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(s + string(filepath.Separator) + fileName + ".go")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tp.Execute(file, data)
}

func JsonToErrorStruct(packageName, fileName, outPath string, data []TemplateKeyValueData) {
	tpl := fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

package %v

type GameError struct {
	msg  string
	code int32
}

var(
	{{range .}}	{{.FieldName}} = newError("{{.Other}}", {{.Values}})
    {{end}}
)
func (c *GameError) Error() string {
	return c.msg
}
func (c *GameError) Code() int32 {
	return c.code
}

func newError(msg string, code int32) *GameError {
	return &GameError{msg: msg, code: code}
}
`, time.Now().String(), packageName)
	t := template.New("GameErrorStruct")
	tp, _ := t.Parse(tpl)
	s, _ := filepath.Abs(outPath)
	// 创建路径中的所有必要的目录
	err := os.MkdirAll(s, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(s + string(filepath.Separator) + fileName + ".go")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tp.Execute(file, data)

}
