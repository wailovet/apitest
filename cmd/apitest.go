package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/robertkrimen/otto"
	"github.com/wailovet/osmanthuswine"
	"github.com/wailovet/osmanthuswine/src/core"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func HttpGet(getUrl string, p string) string {

	if strings.Index(getUrl, "https://") == -1 && strings.Index(getUrl, "http://") == -1 {

		GoAssetError("未设置BASE_HOST")
	}

	allUrl := fmt.Sprintf("%s?%s", getUrl, p)
	//log.Println(allUrl)
	resp, err := http.Get(allUrl)
	if err != nil {
		GoAssetError("http err", "[", allUrl, "]", err.Error())
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		GoAssetError("http err", "[", allUrl, "]", err.Error())
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
		GoAssetError("http err", "[", postUrl, "]", err.Error())
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		GoAssetError("http err", "[", postUrl, "]", err.Error())
	}

	return string(body)

}

func GoAssetError(a ...interface{}) {
	result := []interface{}{"[fail]"}
	result = append(result, a...)
	panic(result)
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

const PASS = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="98" height="20"><linearGradient id="b" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="a"><rect width="98" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#a)"><path fill="#555" d="M0 0h47v20H0z"/><path fill="#4c1" d="M47 0h51v20H47z"/><path fill="url(#b)" d="M0 0h98v20H0z"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="110"> <text x="245" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="370">apitest</text><text x="245" y="140" transform="scale(.1)" textLength="370">apitest</text><text x="715" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="410">passing</text><text x="715" y="140" transform="scale(.1)" textLength="410">passing</text></g> </svg>`
const FAIL = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="86" height="20"><linearGradient id="b" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="a"><rect width="86" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#a)"><path fill="#555" d="M0 0h47v20H0z"/><path fill="#e05d44" d="M47 0h39v20H47z"/><path fill="url(#b)" d="M0 0h86v20H0z"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="110"> <text x="245" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="370">apitest</text><text x="245" y="140" transform="scale(.1)" textLength="370">apitest</text><text x="655" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="290">failed</text><text x="655" y="140" transform="scale(.1)" textLength="290">failed</text></g> </svg>`
const PENDING = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="102" height="20"><linearGradient id="b" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><clipPath id="a"><rect width="102" height="20" rx="3" fill="#fff"/></clipPath><g clip-path="url(#a)"><path fill="#555" d="M0 0h47v20H0z"/><path fill="#dbab09" d="M47 0h55v20H47z"/><path fill="url(#b)" d="M0 0h102v20H0z"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="110"> <text x="245" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="370">apitest</text><text x="245" y="140" transform="scale(.1)" textLength="370">apitest</text><text x="735" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="450">pending</text><text x="735" y="140" transform="scale(.1)" textLength="450">pending</text></g> </svg>`

func start(web string, js string) {
	jss := []byte("")
	if web != "" {
		jss = []byte(HttpGet(web, ""))
	} else {
		jss, _ = ioutil.ReadFile(js)
	}

	getOttoInstance().Run(fmt.Sprintf("%s\n%s", jss, `
		var _fnjkadhaskjdkas = GetAllTest();
		for (var _fadsdi in _fnjkadhaskjdkas){
			_fnjkadhaskjdkas[_fadsdi]();
		}
`))

	fmt.Println("[pass]")
}

var webMap = new(sync.Map)

type testResult struct {
	Result     string `json:"result"`
	UpdateTime int64  `json:"update_time"`
}

func badgeWeb() {

	osmanthuswine.HandleFunc("/badge.svg", func(request core.Request, response core.Response) {

		if request.REQUEST["url"] == "" {
			response.DisplayByError("请传入url参数", 500)
		}
		response.OriginResponseWriter.Header().Set("content-type", "image/svg+xml;charset=utf-8")

		tmp, ok := webMap.Load(request.REQUEST["url"])
		if ok {
			cache := tmp.(testResult)
			if cache.UpdateTime+60 > time.Now().Unix() {
				response.OriginResponseWriter.Write([]byte(cache.Result))
				return
			}
		}

		(func() {
			defer func() {
				a := recover()
				if a != nil {
					response.OriginResponseWriter.Write([]byte(FAIL))
					webMap.Store(request.REQUEST["url"], testResult{
						Result:     FAIL,
						UpdateTime: time.Now().Unix(),
					})
					log.Println(a.([]interface{})...)
				}
			}()

			webMap.Store(request.REQUEST["url"], testResult{
				Result:     PENDING,
				UpdateTime: time.Now().Unix(),
			})
			start(request.REQUEST["url"], "")
			webMap.Store(request.REQUEST["url"], testResult{
				Result:     PASS,
				UpdateTime: time.Now().Unix(),
			})
			response.OriginResponseWriter.Write([]byte(PASS))
		})()

	})

	osmanthuswine.Run()
}

func main() {
	var yapi string
	var js string
	var web string
	var badge bool
	var port string
	flag.StringVar(&yapi, "yapi", "", "yapi配置文件")
	flag.StringVar(&js, "js", "apitest.js", "输出文件")
	flag.StringVar(&web, "web", "", "网络文件")
	flag.StringVar(&port, "p", "80", "徽章模式端口")
	flag.BoolVar(&badge, "badge", false, "徽章模式")
	flag.Parse()


	if yapi != "" {
		generateYapiTest(yapi, js)
		return
	}

	getOttoInstance().Set("Ajax", Ajax)
	getOttoInstance().Set("AssetError", AssetError)

	if badge {
		core.GetInstanceConfig().Port = port
		badgeWeb()
	} else {

		(func() {
			defer func() {
				a := recover()
				if a != nil {

					fmt.Println(a.([]interface{})...)
				}
			}()
			start(web, js)
		})()
		os.Exit(0)
	}
}
