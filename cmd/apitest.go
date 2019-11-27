package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func HttpGet(getUrl string, p string) string {

	if strings.Index(getUrl, "https://") == -1 && strings.Index(getUrl, "http://") == -1 {

		GoAssetError("未设置BASE_HOST")
	}

	allUrl := fmt.Sprintf("%s?%s", getUrl, p)
	//log.Println(allUrl)
	resp, err := http.Get(allUrl)
	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	return string(body)
}

func HttpPostForm(postUrl string, p string) string {
	if strings.Index(postUrl, "https://") == -1 && strings.Index(postUrl, "http://") == -1 {
		GoAssetError("未设置BASE_HOST")
	}

	resp, err := http.Post(postUrl, "application/x-www-form-urlencoded", strings.NewReader(p))

	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	return string(body)

}

func GoAssetError(a ...interface{}) {
	result := []interface{}{"[fail]"}
	result = append(result, a...)
	fmt.Println(result...)
	os.Exit(0)
}

func AssetError(call otto.FunctionCall) otto.Value {
	data := call.Argument(0).String()
	GoAssetError(data)
	return otto.Value{}
}

func Ajax(call otto.FunctionCall) otto.Value {
	method := call.Argument(0).String()
	urls := call.Argument(1).String()

	parameter := call.Argument(2).String()

	data := ""
	if strings.ToLower(method) == "post" {
		data = HttpPostForm(urls, parameter)
	} else {
		data = HttpGet(urls, parameter)
	}
	result, _ := getOttoInstance().ToValue(data)
	return result
}

var vm *otto.Otto

func getOttoInstance() *otto.Otto {
	if vm == nil {
		vm = otto.New()
	}
	return vm
}

type YapiConfigList struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}
type YapiConfig struct {
	Name string           `json:"name"`
	List []YapiConfigList `json:"list"`
}

func JsonByFile(file string, v interface{}) {
	data, _ := ioutil.ReadFile(file)
	json.Unmarshal(data, v)
}
func JsonToFile(file string, v interface{}) bool {
	data, err := json.Marshal(v)
	if err != nil {
		return false
	}
	return ioutil.WriteFile(file, data, 0644) == nil
}

const TEMPLATE = `"%s": function () {
			var url = BASE_HOST + "%s";
			var result = Ajax("get", url, "");
			try {
				JSON.parse(result);
			} catch (e) {
				AssetError("%s:"+result);
			}
		}`

const PLACEHOLDER = "/* PLACEHOLDER 占位符请勿删除*/"

func generateYapiTest(yapi string, out string) {

	fileRaw, _ := ioutil.ReadFile(out)
	fileText := string(fileRaw)
	if fileText == "" {
		fileText = fmt.Sprintf(`
var BASE_HOST = "";
function GetAllTest(){
	return {
		%s
	}
}
`, PLACEHOLDER)
	}

	var yc []YapiConfig
	JsonByFile(yapi, &yc)

	var fileItem []string
	for e := range yc {
		for k := range yc[e].List {
			id := fmt.Sprintf("%s-%s", yc[e].Name, yc[e].List[k].Title)
			id = strings.Replace(id, `\`, "", -1)
			id = strings.Replace(id, `/`, "", -1)
			id = strings.Replace(id, `"`, "", -1)
			if strings.Index(fileText, id) == -1 {
				fileItem = append(fileItem, fmt.Sprintf(TEMPLATE, id, yc[e].List[k].Path, id))
			}
		}
	}

	if len(fileItem) > 0 {
		fileText = strings.Replace(fileText, PLACEHOLDER, fmt.Sprintf(`%s,
		%s`, strings.Join(fileItem, ","), PLACEHOLDER), -1)

		ioutil.WriteFile(out, []byte(fileText), 0644)
	}

}

func main() {
	var yapi string
	var js string
	flag.StringVar(&yapi, "yapi", "", "yapi配置文件")
	flag.StringVar(&js, "js", "apitest.js", "输出文件")
	flag.Parse()

	if yapi != "" {
		generateYapiTest(yapi, js)
		return
	}

	getOttoInstance().Set("Ajax", Ajax)
	getOttoInstance().Set("AssetError", AssetError)

	jss, _ := ioutil.ReadFile(js)

	getOttoInstance().Run(fmt.Sprintf("%s\n%s", jss, `
		var _fnjkadhaskjdkas = GetAllTest();
		for (var _fadsdi in _fnjkadhaskjdkas){
			_fnjkadhaskjdkas[_fadsdi]();
		}
`))

	fmt.Println("[pass]")
	os.Exit(0)
}
