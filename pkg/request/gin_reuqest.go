package gin_request

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

const (
	tag_uri  = "uri"
	tag_json = "json"
	tag_form = "form"
)

// GinRequest get the request from gin context
func GinRequest[T any](ctx *gin.Context) *T {
	objects := make([]T, 0)
	tags := parseTags(new(T))

	for _, tag := range tags {
		switch tag {
		case tag_json:
			request := new(T)
			ctx.ShouldBindJSON(request)
			objects = append(objects, *request)
		case tag_form:
			request := new(T)
			ctx.ShouldBindQuery(request)
			objects = append(objects, *request)
		case tag_uri:
			request := new(T)
			ctx.ShouldBindUri(request)
			objects = append(objects, *request)
		}
	}

	return updateObjectFromObjects(objects)
}

// parseTags get the tags related to the request method
func parseTags[T any](request T) []string {
	m := make(map[string]bool)

	for i := 0; i < reflect.TypeOf(request).Elem().NumField(); i++ {
		f := reflect.TypeOf(request).Elem().Field(i)

		if f.Anonymous {
			tags := parseTags(reflect.New(f.Type).Interface())
			for _, tag := range tags {
				m[tag] = true
			}
			continue
		}

		tag := reflect.TypeOf(request).Elem().Field(i).Tag
		for _, key := range []string{tag_json, tag_form, tag_uri} {
			value := tag.Get(key)
			if len(value) > 0 {
				m[key] = true
			}
		}
	}

	result := make([]string, 0)
	for key := range m {
		result = append(result, key)
	}
	return result
}

func updateObjectFromObjects[T any](objects []T) *T {
	if len(objects) == 0 {
		return nil
	}

	objectVal := reflect.ValueOf(new(T)).Elem()
	for _, o := range objects {
		val := reflect.ValueOf(o)
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if !field.IsZero() && objectVal.Field(i).CanSet() {
				if objectVal.Field(i).IsZero() && !objectVal.Field(i).CanSet() {
					continue
				}
				switch field.Kind() {
				case reflect.String:
					objectVal.Field(i).SetString(field.String())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					objectVal.Field(i).SetInt(field.Int())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					objectVal.Field(i).SetUint(field.Uint())
				case reflect.Float32, reflect.Float64:
					objectVal.Field(i).SetFloat(field.Float())
				case reflect.Bool:
					objectVal.Field(i).SetBool(field.Bool())
				case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct, reflect.Ptr:
					objectVal.Field(i).Set(field)
				}
			}
		}
	}

	output := objectVal.Interface().(T)
	return &output
}
