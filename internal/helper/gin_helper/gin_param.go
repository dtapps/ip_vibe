package gin_helper

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func setDefaults(obj any) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}
	setDefaultsRecursive(v)
}

func setDefaultsRecursive(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !value.CanSet() {
			continue
		}

		// 递归嵌套 struct
		if value.Kind() == reflect.Struct {
			setDefaultsRecursive(value)
			continue
		}

		// 仅处理有 default tag 的字段
		if def, ok := field.Tag.Lookup("default"); ok {
			switch value.Kind() {
			case reflect.String:
				value.SetString(def)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if i, err := strconv.ParseInt(def, 10, 64); err == nil {
					value.SetInt(i)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if u, err := strconv.ParseUint(def, 10, 64); err == nil {
					value.SetUint(u)
				}
			case reflect.Float32, reflect.Float64:
				if f, err := strconv.ParseFloat(def, 64); err == nil {
					value.SetFloat(f)
				}
			case reflect.Bool:
				if b, err := strconv.ParseBool(def); err == nil {
					value.SetBool(b)
				}
			}
		}
	}
}

// 通用绑定器（推荐方式）
func BindWithDefaults(c *gin.Context, obj any, b binding.Binding) (err error) {

	setDefaults(obj)             // ← 先设默认值
	err = b.Bind(c.Request, obj) // ← 再绑定，覆盖传了的字段

	// var (
	// 	method        = c.Request.Method
	// 	path          = c.Request.URL.Path
	// 	contentLength = c.Request.ContentLength
	// 	contentType   = c.GetHeader("Content-Type")
	// )

	return err
}

// 快捷方法
func BindQueryWithDefaults(c *gin.Context, obj any) (err error) {

	// 先设默认值
	setDefaults(obj)

	// var (
	// 	method        = c.Request.Method
	// 	path          = c.Request.URL.Path
	// 	contentLength = c.Request.ContentLength
	// 	contentType   = c.GetHeader("Content-Type")
	// )

	// 再绑定，覆盖传了的字段
	err = c.ShouldBindQuery(obj)
	if err != nil {
		return err
	}

	return nil
}

func BindJSONWithDefaults(c *gin.Context, obj any) (err error) {

	// 先设默认值
	setDefaults(obj)

	var (
		// method        = c.Request.Method
		// path          = c.Request.URL.Path
		contentLength = c.Request.ContentLength
		contentType   = c.GetHeader("Content-Type")
	)

	// 仅当有 JSON body 时才绑定
	if contentLength > 0 && strings.Contains(contentType, "application/json") {
		// 再绑定，覆盖传了的字段
		err = c.ShouldBindJSON(obj)
		if err != nil {
			return err
		}
	}

	return nil
}

func BindUriWithDefaults(c *gin.Context, obj any) (err error) {

	setDefaults(obj)           // ← 先设默认值
	err = c.ShouldBindUri(obj) // ← 再绑定，覆盖传了的字段

	// var (
	// 	method        = c.Request.Method
	// 	path          = c.Request.URL.Path
	// 	contentLength = c.Request.ContentLength
	// 	contentType   = c.GetHeader("Content-Type")
	// )

	if err != nil {
		return err
	}

	return nil
}
