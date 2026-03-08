package zinc

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"reflect"
	"strconv"
	"strings"
)

type Binder interface {
	Bind(*Context, any) error
	BindBody(*Context, any) error
	BindQuery(*Context, any) error
	BindForm(*Context, any) error
	BindHeader(*Context, any) error
	BindPath(*Context, any) error
}

type Validator interface {
	Validate(any) error
}

type Renderer interface {
	Render(w io.Writer, name string, data any, c *Context) error
}

type JSONCodec interface {
	Encode(w io.Writer, v any, indent string) error
	Decode(r io.Reader, v any) error
}

type defaultBinder struct {
	codec JSONCodec
}

func (b defaultBinder) Bind(c *Context, v any) error {
	pathValues := make(map[string][]string, c.paramCount)
	for i := 0; i < c.paramCount; i++ {
		pathValues[c.PathParams[i].key] = []string{c.pathParamValueAt(i)}
	}
	if err := bindData(v, pathValues, "path"); err != nil {
		return err
	}
	if c.Request().URL.RawQuery != "" {
		if err := bindData(v, c.QueryValues(), "query"); err != nil {
			return err
		}
	}
	if c.Request().Body == nil {
		return c.Validate(v)
	}
	body, err := c.bodyBytes()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return c.Validate(v)
	}
	mediaType, _, _ := mime.ParseMediaType(c.GetHeader(HeaderContentType))
	switch mediaType {
	case "", "application/json":
		if err := b.codec.Decode(bytes.NewReader(body), v); err != nil {
			return fmt.Errorf("bind body: %w", err)
		}
	case "application/xml", "text/xml":
		if err := xml.NewDecoder(bytes.NewReader(body)).Decode(v); err != nil {
			return fmt.Errorf("bind body: %w", err)
		}
	case "application/x-www-form-urlencoded", "multipart/form-data":
		if err := c.Request().ParseForm(); err != nil {
			return fmt.Errorf("parse form: %w", err)
		}
		if err := bindData(v, c.Request().Form, "form"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported content type: %s", mediaType)
	}
	return c.Validate(v)
}

func (b defaultBinder) BindBody(c *Context, v any) error {
	body, err := c.bodyBytes()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("request body is empty")
	}

	mediaType, _, _ := mime.ParseMediaType(c.GetHeader(HeaderContentType))
	switch mediaType {
	case "", "application/json":
		if err := b.codec.Decode(bytes.NewReader(body), v); err != nil {
			return fmt.Errorf("bind body: %w", err)
		}
	case "application/xml", "text/xml":
		if err := xml.NewDecoder(bytes.NewReader(body)).Decode(v); err != nil {
			return fmt.Errorf("bind body: %w", err)
		}
	case "application/x-www-form-urlencoded", "multipart/form-data":
		return b.BindForm(c, v)
	default:
		return fmt.Errorf("unsupported content type: %s", mediaType)
	}
	return c.Validate(v)
}

func (b defaultBinder) BindQuery(c *Context, v any) error {
	if err := bindData(v, c.QueryValues(), "query"); err != nil {
		return err
	}
	return c.Validate(v)
}

func (b defaultBinder) BindForm(c *Context, v any) error {
	if err := c.Request().ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", err)
	}
	if err := bindData(v, c.Request().Form, "form"); err != nil {
		return err
	}
	return c.Validate(v)
}

func (b defaultBinder) BindHeader(c *Context, v any) error {
	values := make(map[string][]string, len(c.Request().Header))
	for key, val := range c.Request().Header {
		values[strings.ToLower(key)] = val
	}
	if err := bindData(v, values, "header"); err != nil {
		return err
	}
	return c.Validate(v)
}

func (b defaultBinder) BindPath(c *Context, v any) error {
	values := make(map[string][]string, c.paramCount)
	for i := 0; i < c.paramCount; i++ {
		values[c.PathParams[i].key] = []string{c.pathParamValueAt(i)}
	}
	if err := bindData(v, values, "path"); err != nil {
		return err
	}
	return c.Validate(v)
}

func (c *Context) Bind(v any) error {
	return c.app.config.Binder.Bind(c, v)
}

func (c *Context) BindJSON(v any) error {
	return c.app.config.Binder.BindBody(c, v)
}

func (c *Context) BindXML(v any) error {
	body, err := c.bodyBytes()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.New("request body is empty")
	}
	if err := xml.NewDecoder(bytes.NewReader(body)).Decode(v); err != nil {
		return err
	}
	return c.Validate(v)
}

func (c *Context) BindForm(v any) error {
	return c.app.config.Binder.BindForm(c, v)
}

func (c *Context) BindQuery(v any) error {
	return c.app.config.Binder.BindQuery(c, v)
}

func (c *Context) BindHeader(v any) error {
	return c.app.config.Binder.BindHeader(c, v)
}

func (c *Context) BindPath(v any) error {
	return c.app.config.Binder.BindPath(c, v)
}

func (c *Context) Validate(v any) error {
	if c.app == nil || c.app.config.Validator == nil {
		return nil
	}
	return c.app.config.Validator.Validate(v)
}

func bindData(ptr any, data map[string][]string, tag string) error {
	if ptr == nil {
		return errors.New("binding target must not be nil")
	}

	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return errors.New("binding target must be a pointer")
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("binding target must point to a struct")
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		name := field.Tag.Get(tag)
		if name == "-" {
			continue
		}
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		if idx := strings.IndexByte(name, ','); idx >= 0 {
			name = name[:idx]
		}
		lookup := name
		if tag == "header" {
			lookup = strings.ToLower(name)
		}
		values, ok := data[lookup]
		if !ok || len(values) == 0 {
			continue
		}
		if err := setFieldValue(fieldValue, values); err != nil {
			return fmt.Errorf("bind %s: %w", field.Name, err)
		}
	}
	return nil
}

func setFieldValue(value reflect.Value, inputs []string) error {
	if !value.CanSet() {
		return nil
	}
	if len(inputs) == 0 {
		return nil
	}

	switch value.Kind() {
	case reflect.String:
		value.SetString(inputs[0])
	case reflect.Bool:
		parsed, err := strconv.ParseBool(inputs[0])
		if err != nil {
			return err
		}
		value.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(inputs[0], 10, 64)
		if err != nil {
			return err
		}
		value.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(inputs[0], 10, 64)
		if err != nil {
			return err
		}
		value.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(inputs[0], 64)
		if err != nil {
			return err
		}
		value.SetFloat(parsed)
	case reflect.Slice:
		elem := value.Type().Elem().Kind()
		slice := reflect.MakeSlice(value.Type(), 0, len(inputs))
		for _, input := range inputs {
			switch elem {
			case reflect.String:
				slice = reflect.Append(slice, reflect.ValueOf(input))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				parsed, err := strconv.ParseInt(input, 10, 64)
				if err != nil {
					return err
				}
				item := reflect.New(value.Type().Elem()).Elem()
				item.SetInt(parsed)
				slice = reflect.Append(slice, item)
			default:
				return fmt.Errorf("unsupported slice element type %s", value.Type().Elem())
			}
		}
		value.Set(slice)
	default:
		return fmt.Errorf("unsupported kind %s", value.Kind())
	}
	return nil
}

type defaultJSONCodec struct{}

func (defaultJSONCodec) Encode(w io.Writer, v any, indent string) error {
	enc := json.NewEncoder(w)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(v)
}

func (defaultJSONCodec) Decode(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	return dec.Decode(v)
}
