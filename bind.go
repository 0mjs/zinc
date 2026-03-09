package zinc

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
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
	val, plan, err := bindTargetPlan(v)
	if err != nil {
		return err
	}
	if err := bindFieldsFromPath(val, plan.pathFields, c); err != nil {
		return err
	}
	if len(plan.queryFields) > 0 && c.Request().URL.RawQuery != "" {
		if err := bindFieldsFromValues(val, plan.queryFields, c.QueryValues()); err != nil {
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
	mediaType := requestMediaType(c.GetHeader(HeaderContentType))
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
		if err := bindFieldsFromValues(val, plan.formFields, c.Request().Form); err != nil {
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

	mediaType := requestMediaType(c.GetHeader(HeaderContentType))
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
	val, plan, err := bindTargetPlan(v)
	if err != nil {
		return err
	}
	if err := bindFieldsFromValues(val, plan.queryFields, c.QueryValues()); err != nil {
		return err
	}
	return c.Validate(v)
}

func (b defaultBinder) BindForm(c *Context, v any) error {
	if err := c.Request().ParseForm(); err != nil {
		return fmt.Errorf("parse form: %w", err)
	}
	val, plan, err := bindTargetPlan(v)
	if err != nil {
		return err
	}
	if err := bindFieldsFromValues(val, plan.formFields, c.Request().Form); err != nil {
		return err
	}
	return c.Validate(v)
}

func (b defaultBinder) BindHeader(c *Context, v any) error {
	val, plan, err := bindTargetPlan(v)
	if err != nil {
		return err
	}
	if err := bindFieldsFromHeader(val, plan.headerFields, c.Request().Header); err != nil {
		return err
	}
	return c.Validate(v)
}

func (b defaultBinder) BindPath(c *Context, v any) error {
	val, plan, err := bindTargetPlan(v)
	if err != nil {
		return err
	}
	if err := bindFieldsFromPath(val, plan.pathFields, c); err != nil {
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
	val, plan, err := bindTargetPlan(ptr)
	if err != nil {
		return err
	}
	switch tag {
	case "path":
		return bindFieldsFromValues(val, plan.pathFields, data)
	case "query":
		return bindFieldsFromValues(val, plan.queryFields, data)
	case "form":
		return bindFieldsFromValues(val, plan.formFields, data)
	case "header":
		return bindFieldsFromValues(val, plan.headerFields, data)
	default:
		return nil
	}
}

func setFieldValue(value reflect.Value, inputs []string) error {
	return compileFieldSetter(value.Type()).set(value, inputs)
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
