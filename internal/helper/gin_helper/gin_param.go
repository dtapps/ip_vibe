package gin_helper

import (
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"vibe_ip/internal/global"

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
	// 日志
	logger := slog.Default()
	if global.GvaGinSlog != nil {
		logger = global.GvaGinSlog.WithLogger()
	}

	setDefaults(obj)             // ← 先设默认值
	err = b.Bind(c.Request, obj) // ← 再绑定，覆盖传了的字段

	var (
		method        = c.Request.Method
		path          = c.Request.URL.Path
		contentLength = c.Request.ContentLength
		contentType   = c.GetHeader("Content-Type")
	)

	if global.GvaEnv.DEBUG {
		logger.DebugContext(c, "BindWithDefaults",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int64("contentLength", contentLength),
			slog.String("contentType", contentType),
			slog.Any("obj", obj),
			slog.Any("err", err),
		)
	}
	return err
}

// 快捷方法
func BindQueryWithDefaults(c *gin.Context, obj any) (err error) {
	// 日志
	logger := slog.Default()
	if global.GvaGinSlog != nil {
		logger = global.GvaGinSlog.WithLogger()
	}

	// 先设默认值
	setDefaults(obj)

	var (
		method        = c.Request.Method
		path          = c.Request.URL.Path
		contentLength = c.Request.ContentLength
		contentType   = c.GetHeader("Content-Type")
	)

	// 再绑定，覆盖传了的字段
	err = c.ShouldBindQuery(obj)
	if err != nil {
		if global.GvaEnv.DEBUG {
			logger.ErrorContext(c, "ShouldBindQuery",
				slog.String("method", method),
				slog.String("path", path),
				slog.Int64("contentLength", contentLength),
				slog.String("contentType", contentType),
				slog.Any("obj", obj),
				slog.Any("err", err),
			)
		}
		return err
	}

	if global.GvaEnv.DEBUG {
		logger.DebugContext(c, "BindQueryWithDefaults",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int64("contentLength", contentLength),
			slog.String("contentType", contentType),
			slog.Any("obj", obj),
			slog.Any("err", err),
		)
	}
	return nil
}

func BindJSONWithDefaults(c *gin.Context, obj any) (err error) {
	// 日志
	logger := slog.Default()
	if global.GvaGinSlog != nil {
		logger = global.GvaGinSlog.WithLogger()
	}

	// 先设默认值
	setDefaults(obj)

	var (
		method        = c.Request.Method
		path          = c.Request.URL.Path
		contentLength = c.Request.ContentLength
		contentType   = c.GetHeader("Content-Type")
	)

	// 仅当有 JSON body 时才绑定
	if contentLength > 0 && strings.Contains(contentType, "application/json") {
		// 再绑定，覆盖传了的字段
		err = c.ShouldBindJSON(obj)
		if err != nil {
			if global.GvaEnv.DEBUG {
				logger.ErrorContext(c, "ShouldBindJSON",
					slog.String("method", method),
					slog.String("path", path),
					slog.Int64("contentLength", contentLength),
					slog.String("contentType", contentType),
					slog.Any("obj", obj),
					slog.Any("err", err),
				)
			}
			return err
		}
	}

	if global.GvaEnv.DEBUG {
		logger.DebugContext(c, "BindJSONWithDefaults",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int64("contentLength", contentLength),
			slog.String("contentType", contentType),
			slog.Any("obj", obj),
			slog.Any("err", err),
		)
	}
	return nil
}

func BindUriWithDefaults(c *gin.Context, obj any) (err error) {
	// 日志
	logger := slog.Default()
	if global.GvaGinSlog != nil {
		logger = global.GvaGinSlog.WithLogger()
	}

	setDefaults(obj)           // ← 先设默认值
	err = c.ShouldBindUri(obj) // ← 再绑定，覆盖传了的字段

	var (
		method        = c.Request.Method
		path          = c.Request.URL.Path
		contentLength = c.Request.ContentLength
		contentType   = c.GetHeader("Content-Type")
	)

	if err != nil {
		if global.GvaEnv.DEBUG {
			logger.ErrorContext(c, "BindUriWithDefaults",
				slog.String("method", method),
				slog.String("path", path),
				slog.Int64("contentLength", contentLength),
				slog.String("contentType", contentType),
				slog.Any("obj", obj),
				slog.Any("err", err),
			)
		}
		return err
	}

	if global.GvaEnv.DEBUG {
		logger.DebugContext(c, "BindUriWithDefaults",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int64("contentLength", contentLength),
			slog.String("contentType", contentType),
			slog.Any("obj", obj),
			slog.Any("err", err),
		)
	}
	return nil
}

// func BindUriAndJSONWithDefaults(c *gin.Context, obj any) (err error) {
// 	setDefaults(obj)           // ← 先设默认值
// 	err = c.ShouldBindUri(obj) // ← 再绑定，覆盖传了的字段
// 	if err != nil {
// 		if global.GvaEnv.DEBUG {
// 			fmt.Printf("%s %s ShouldBindUri: %+v ;Err: %v \n", c.Request.Method, c.Request.URL.Path, obj, err)
// 		}
// 		return err
// 	}
//
// 	// 仅当有 JSON body 时才绑定
// 	if c.Request.ContentLength > 0 && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
// 		err = c.ShouldBindJSON(obj) // ← 再绑定，覆盖传了的字段
// 		if err != nil {
// 			if global.GvaEnv.DEBUG {
// 				fmt.Printf("%s %s ShouldBindJSON: %+v ;Err: %v \n", c.Request.Method, c.Request.URL.Path, obj, err)
// 			}
// 			return err
// 		}
// 	}

// 	if global.GvaEnv.DEBUG {
// 		fmt.Printf("%s %s BindUriAndJSONWithDefaults: %+v ;Err: %v \n", c.Request.Method, c.Request.URL.Path, obj, err)
// 	}
// 	return nil
// }
