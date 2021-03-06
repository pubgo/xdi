package dix

import (
	"reflect"
	"sort"

	"github.com/pubgo/xerror"
)

type node struct {
	c          *dix
	fn         reflect.Value
	input      []value
	outputType map[group]key
}

func newNode(c *dix, data reflect.Value) (nd *node, err error) {
	nd = &node{fn: data, c: c, outputType: make(map[group]key)}
	nd.outputType, err = nd.returnedType()
	return
}

func (n *node) returnedType() (map[group]key, error) {
	retType := make(map[group]key)
	v := n.fn
	for i := 0; i < v.Type().NumOut(); i++ {
		out := v.Type().Out(i)
		switch out.Kind() {
		case reflect.Ptr:
			retType[_default] = getIndirectType(out)
		case reflect.Struct:
			for j := 0; j < v.NumField(); j++ {
				feTye := v.Type().Field(j)
				xerror.Assert(feTye.Type.Kind() != reflect.Ptr,
					"the struct field[%s] should be Ptr type", feTye.Type.Kind())
				retType[n.c.getNS(feTye)] = getIndirectType(feTye.Type)
			}
		default:
			if isError(out) {
				continue
			}
			return nil, xerror.Fmt("provide type kind error, (kind %v)", out.Kind())
		}
	}
	return retType, nil
}

func (n *node) handleCall(params []reflect.Value) (err error) {
	defer xerror.RespErr(&err)

	values := defaultInvoker(n.fn, params[:])

	if len(values) == 0 {
		return nil
	}

	// the returned value should be error
	vErr := values[len(values)-1]
	xerror.Assert(!isError(vErr.Type()), "the last returned value should be error type, got(%v)", vErr.Type())

	if vErr.IsValid() && !vErr.IsNil() {
		xerror.PanicF(vErr.Interface().(error), "func error, func: %s, params: %s", callerWithFunc(n.fn), params)
	}

	var vas []interface{}
	for i := range values[:len(values)-1] {
		vas = append(vas, values[i].Interface())
	}

	if len(vas) == 0 {
		return
	}

	return xerror.Wrap(n.c.dix(vas...))
}

func (n *node) isNil(v reflect.Value) bool {
	return n.c.isNil(v)
}

type sortValue struct {
	Key   string
	Value reflect.Value
}

func (n *node) call() (err error) {
	defer xerror.RespErr(&err)

	var values []reflect.Value
	var params []reflect.Value
	var input []reflect.Value
	for i := 0; i < n.fn.Type().NumIn(); i++ {
		inType := n.fn.Type().In(i)
		switch inType.Kind() {
		case reflect.Interface:
			val := n.c.getAbcValue(getIndirectType(inType), _default)
			if !n.isNil(val) {
				params = append(params, val)
				input = append(input, val)
			}
		case reflect.Ptr:
			val := n.c.getValue(getIndirectType(inType), _default)
			if !n.isNil(val) {
				params = append(params, val)
				input = append(input, val)
			}
		case reflect.Struct:
			mt := reflect.New(inType).Elem()
			var sv []sortValue
			for i := 0; i < inType.NumField(); i++ {
				field := inType.Field(i)

				if _, ok := field.Tag.Lookup(_tagName); n.c.opts.Strict && !ok {
					continue
				}

				// 结构体里面所有的属性值全部有值, 且不为nil
				var val reflect.Value
				if getIndirectType(field.Type).Kind() == reflect.Interface {
					// 如果value为nil, 就不触发初始化
					val = n.c.getAbcValue(getIndirectType(field.Type), n.c.getNS(field))
					if n.isNil(val) {
						return nil
					}

					values = append(values, val)
					mt.Field(i).Set(val)
				} else {
					// 如果value为nil, 就不触发初始化
					val = n.c.getValue(getIndirectType(field.Type), n.c.getNS(field))
					if n.isNil(val) {
						return nil
					}

					values = append(values, val)
					mt.Field(i).Set(val)
				}

				sv = append(sv, sortValue{Key: n.c.getNS(field), Value: val})
			}

			sort.Slice(sv, func(i, j int) bool { return sv[i].Key > sv[j].Key })
			for i := range sv {
				input = append(input, sv[i].Value)
			}

			params = append(params, mt)
		default:
			return xerror.Fmt("incorrect input parameter type, got(%s)", inType.Kind())
		}
	}

	if equal(n.input, input) {
		return nil
	}

	if err := n.handleCall(params); err != nil {
		return xerror.Wrap(err)
	}

	n.input = input
	return nil
}
