package zinc

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type bindingPlan struct {
	pathFields   []bindingField
	queryFields  []bindingField
	formFields   []bindingField
	headerFields []bindingField
}

type bindingField struct {
	index  int
	name   string
	label  string
	setter fieldSetter
}

type fieldSetter struct {
	kind             fieldSetterKind
	bits             int
	elemBits         int
	unsupportedKind  reflect.Kind
	unsupportedSlice reflect.Type
}

type fieldSetterKind uint8

const (
	fieldSetterString fieldSetterKind = iota
	fieldSetterBool
	fieldSetterInt
	fieldSetterUint
	fieldSetterFloat
	fieldSetterSliceString
	fieldSetterSliceInt
	fieldSetterUnsupportedKind
	fieldSetterUnsupportedSlice
)

var bindingPlanCache sync.Map

func bindTargetPlan(ptr any) (reflect.Value, *bindingPlan, error) {
	if ptr == nil {
		return reflect.Value{}, nil, fmt.Errorf("binding target must not be nil")
	}

	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return reflect.Value{}, nil, fmt.Errorf("binding target must be a pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("binding target must point to a struct")
	}

	return val, bindingPlanFor(val.Type()), nil
}

func bindingPlanFor(typ reflect.Type) *bindingPlan {
	if cached, ok := bindingPlanCache.Load(typ); ok {
		return cached.(*bindingPlan)
	}

	plan := compileBindingPlan(typ)
	actual, _ := bindingPlanCache.LoadOrStore(typ, plan)
	return actual.(*bindingPlan)
}

func compileBindingPlan(typ reflect.Type) *bindingPlan {
	plan := &bindingPlan{}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}

		setter := compileFieldSetter(field.Type)
		if compiled, ok := compileBindingField(i, field, setter, "path"); ok {
			plan.pathFields = append(plan.pathFields, compiled)
		}
		if compiled, ok := compileBindingField(i, field, setter, "query"); ok {
			plan.queryFields = append(plan.queryFields, compiled)
		}
		if compiled, ok := compileBindingField(i, field, setter, "form"); ok {
			plan.formFields = append(plan.formFields, compiled)
		}
		if compiled, ok := compileBindingField(i, field, setter, "header"); ok {
			plan.headerFields = append(plan.headerFields, compiled)
		}
	}
	return plan
}

func compileBindingField(index int, field reflect.StructField, setter fieldSetter, tag string) (bindingField, bool) {
	name, ok := bindingFieldName(field, tag)
	if !ok {
		return bindingField{}, false
	}
	return bindingField{
		index:  index,
		name:   name,
		label:  field.Name,
		setter: setter,
	}, true
}

func bindingFieldName(field reflect.StructField, tag string) (string, bool) {
	name := field.Tag.Get(tag)
	if name == "-" {
		return "", false
	}
	if idx := strings.IndexByte(name, ','); idx >= 0 {
		name = name[:idx]
	}
	if name == "" {
		name = strings.ToLower(field.Name)
	}
	if tag == "header" {
		name = strings.ToLower(name)
	}
	return name, true
}

func compileFieldSetter(typ reflect.Type) fieldSetter {
	switch typ.Kind() {
	case reflect.String:
		return fieldSetter{kind: fieldSetterString}
	case reflect.Bool:
		return fieldSetter{kind: fieldSetterBool}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fieldSetter{kind: fieldSetterInt, bits: int(typ.Bits())}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fieldSetter{kind: fieldSetterUint, bits: int(typ.Bits())}
	case reflect.Float32, reflect.Float64:
		return fieldSetter{kind: fieldSetterFloat, bits: int(typ.Bits())}
	case reflect.Slice:
		switch typ.Elem().Kind() {
		case reflect.String:
			return fieldSetter{kind: fieldSetterSliceString}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fieldSetter{kind: fieldSetterSliceInt, elemBits: int(typ.Elem().Bits())}
		default:
			return fieldSetter{kind: fieldSetterUnsupportedSlice, unsupportedSlice: typ.Elem()}
		}
	default:
		return fieldSetter{kind: fieldSetterUnsupportedKind, unsupportedKind: typ.Kind()}
	}
}

func bindFieldsFromValues(val reflect.Value, fields []bindingField, values url.Values) error {
	if len(fields) == 0 || len(values) == 0 {
		return nil
	}
	for _, field := range fields {
		inputs, ok := values[field.name]
		if !ok || len(inputs) == 0 {
			continue
		}
		if err := field.setter.set(val.Field(field.index), inputs); err != nil {
			return fmt.Errorf("bind %s: %w", field.label, err)
		}
	}
	return nil
}

func bindFieldsFromHeader(val reflect.Value, fields []bindingField, header http.Header) error {
	if len(fields) == 0 || len(header) == 0 {
		return nil
	}
	for _, field := range fields {
		inputs := header.Values(field.name)
		if len(inputs) == 0 {
			continue
		}
		if err := field.setter.set(val.Field(field.index), inputs); err != nil {
			return fmt.Errorf("bind %s: %w", field.label, err)
		}
	}
	return nil
}

func bindFieldsFromPath(val reflect.Value, fields []bindingField, c *Context) error {
	if len(fields) == 0 || c == nil || c.paramCount == 0 {
		return nil
	}
	for _, field := range fields {
		input, ok := c.lookupPathParam(field.name)
		if !ok {
			continue
		}
		single := [1]string{input}
		if err := field.setter.set(val.Field(field.index), single[:]); err != nil {
			return fmt.Errorf("bind %s: %w", field.label, err)
		}
	}
	return nil
}

func (s fieldSetter) set(value reflect.Value, inputs []string) error {
	if !value.CanSet() || len(inputs) == 0 {
		return nil
	}

	switch s.kind {
	case fieldSetterString:
		value.SetString(inputs[0])
	case fieldSetterBool:
		parsed, err := strconv.ParseBool(inputs[0])
		if err != nil {
			return err
		}
		value.SetBool(parsed)
	case fieldSetterInt:
		parsed, err := strconv.ParseInt(inputs[0], 10, s.bits)
		if err != nil {
			return err
		}
		value.SetInt(parsed)
	case fieldSetterUint:
		parsed, err := strconv.ParseUint(inputs[0], 10, s.bits)
		if err != nil {
			return err
		}
		value.SetUint(parsed)
	case fieldSetterFloat:
		parsed, err := strconv.ParseFloat(inputs[0], s.bits)
		if err != nil {
			return err
		}
		value.SetFloat(parsed)
	case fieldSetterSliceString:
		slice := reflect.MakeSlice(value.Type(), len(inputs), len(inputs))
		for i, input := range inputs {
			slice.Index(i).SetString(input)
		}
		value.Set(slice)
	case fieldSetterSliceInt:
		slice := reflect.MakeSlice(value.Type(), len(inputs), len(inputs))
		for i, input := range inputs {
			parsed, err := strconv.ParseInt(input, 10, s.elemBits)
			if err != nil {
				return err
			}
			slice.Index(i).SetInt(parsed)
		}
		value.Set(slice)
	case fieldSetterUnsupportedSlice:
		return fmt.Errorf("unsupported slice element type %s", s.unsupportedSlice)
	case fieldSetterUnsupportedKind:
		return fmt.Errorf("unsupported kind %s", s.unsupportedKind)
	}
	return nil
}

func (c *Context) lookupPathParam(name string) (string, bool) {
	if route := c.paramRoute; route != nil {
		if c.paramPath != "" && c.paramCount > 1 && len(route.paramIndices) > 0 {
			c.materializePathParams()
		}
		if index, ok := route.paramIndex(name); ok {
			if index >= c.paramCount {
				return "", false
			}
			if c.PathParams[index].start == directParamStart {
				return c.PathParams[index].value, true
			}
			return c.pathParamValueAt(index), true
		}
		return "", false
	}
	if c.paramPath != "" && c.paramCount > 1 {
		c.materializePathParams()
	}
	for i := 0; i < c.paramCount; i++ {
		if c.PathParams[i].key != name {
			continue
		}
		if c.PathParams[i].start == directParamStart {
			return c.PathParams[i].value, true
		}
		return c.pathParamValueAt(i), true
	}
	return "", false
}

func requestMediaType(header string) string {
	base, _, _ := strings.Cut(header, ";")
	return strings.TrimSpace(base)
}
