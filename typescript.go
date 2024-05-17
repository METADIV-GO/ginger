package ginger

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

type ApiInfo struct {
	Imports map[string]bool
	Content string
}

var FileService = new(fileService)

type fileService struct{}

func (s *fileService) RemoveOldFiles() {
	os.RemoveAll("./apis")
}

func (s *fileService) CreateFolder() {
	folder := "./apis/"
	os.MkdirAll(folder, os.ModePerm)
}

var ModelService = new(modelService)

type modelService struct{}

func (s *modelService) GenerateGeneral() {
	os.WriteFile("./apis/general.ts", []byte(`
export interface ErrorResp {
	code: string;
	message: string;
}
export interface Response<T> {
	success: boolean;
	duration: number;
	pagination?: any;
	error?: ErrorResp;
	data?: T;
}
`), os.ModePerm)
}

func (s *modelService) CreateModels() {
	convertor := typescriptify.New()
	convertor.CreateInterface = true
	convertor.BackupDir = ""
	convertor.ManageType(time.Time{}, typescriptify.TypeOptions{TSType: "Date", TSTransform: "new Date(__VALUE__)"})

	for i := range Engine.ApiHandlers {
		if Engine.ApiHandlers[i].Opts == nil || Engine.ApiHandlers[i].Opts.Typescript == nil {
			continue
		}
		opt := Engine.ApiHandlers[i].Opts.Typescript
		for j := range opt.Models {
			convertor.Add(opt.Models[j])
		}
	}

	convertor.ConvertToFile("./apis/models.ts")
}

var ApiService = new(apiService)

type apiService struct{}

func (s *apiService) CreateApis() {

	apiInfo := new(ApiInfo)
	apiInfo.Imports = make(map[string]bool)
	for _, api := range Engine.ApiHandlers {

		if api.Handler == nil {
			continue
		}
		if api.Opts == nil || api.Opts.Typescript == nil {
			continue
		}
		opt := api.Opts.Typescript

		if opt.Body != "" {
			apiInfo.Imports[strings.ReplaceAll(opt.Body, "[]", "")] = true
		}
		if opt.Response != "" {
			apiInfo.Imports[strings.ReplaceAll(opt.Response, "[]", "")] = true
		}

		if opt.FunctionName == "" {
			panic("typescript: function name is empty for " + api.Path)
		}

		apiContent := "export const " + opt.FunctionName + " = ("
		if opt.Paths != nil && len(opt.Paths) > 0 {
			for i := range opt.Paths {
				apiContent += opt.Paths[i] + ": any, "
			}
		}
		if opt.Forms != nil && len(opt.Forms) > 0 {
			for i := range opt.Forms {
				apiContent += opt.Forms[i] + ": any, "
			}
		}
		if opt.Body != "" {
			apiContent += "req: " + opt.Body + ", "
		}
		apiContent = strings.TrimSuffix(apiContent, ", ")
		var resp string
		if opt.Response != "" {
			resp = "Promise<AxiosResponse<Response<" + opt.Response + ">>>"
		} else {
			resp = "Promise<AxiosResponse<Response<void>>>"
		}
		apiContent += "): " + resp + " => {\n"

		var method string
		if api.Method == http.MethodGet {
			method = "get"
		} else if api.Method == http.MethodPost {
			method = "post"
		} else if api.Method == http.MethodPut {
			method = "put"
		} else if api.Method == http.MethodDelete {
			method = "delete"
		} else {
			panic("typescript: method is empty")
		}

		var url string = api.Path
		if strings.Contains(url, ":") {
			elements := strings.Split(url, "/")
			for i := range elements {
				if strings.Contains(elements[i], ":") {
					url = strings.ReplaceAll(url, elements[i], "${"+elements[i][1:]+"}")
				}
			}
		}

		var query string
		if len(opt.Forms) > 0 {
			query = "?"
			for i := range opt.Forms {
				query += opt.Forms[i] + "=${" + opt.Forms[i] + "}&"
			}
		}
		query = strings.TrimSuffix(query, "&")

		apiContent += "\treturn axios." + method + "(`" + url + query + "`"
		if opt.Body != "" {
			apiContent += ", req"
		}
		apiContent += ");\n"
		apiContent += "}\n\n"

		apiInfo.Content += apiContent

	}

	page := "import axios, { AxiosResponse } from 'axios';\n"
	page += "import { Response } from './general';\n"
	for model := range apiInfo.Imports {
		page += "import { " + model + " } from './models';\n"
	}
	page += "\n"
	page += apiInfo.Content
	os.WriteFile("./apis/api.ts", []byte(page), os.ModePerm)
}
