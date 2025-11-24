package pipe

import (
	"html/template"
	"reflect"
	"strconv"
	"strings"
)

func Functions() template.FuncMap {
	return functions
}

var functions = template.FuncMap{
	"json":     Json,
	"thousand": Thousand,
	// "dollar":   Dollar,
}

func Thousand(vals ...reflect.Value) reflect.Value {
	if len(vals) == 0 {
		return reflect.ValueOf("")
	}

	val := vals[len(vals)-1] // Giá trị chính cuối cùng

	// Mặc định
	dot := "."   // thousand separator
	decimal := 0 // số chữ số thập phân
	point := "," // decimal separator

	// Xử lý các tham số tùy chọn
	switch len(vals) {
	case 2:
		if s, ok := vals[0].Interface().(string); ok {
			dot = s
		}
	case 3:
		if s, ok := vals[0].Interface().(string); ok {
			dot = s
		}
		if n, ok := vals[1].Interface().(int); ok {
			decimal = n
		}
	case 4:
		if s, ok := vals[0].Interface().(string); ok {
			dot = s
		}
		if n, ok := vals[1].Interface().(int); ok {
			decimal = n
		}
		if p, ok := vals[2].Interface().(string); ok {
			point = p
		}
	}

	// Chuyển val sang float64
	var num float64
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num = float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num = float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		num = val.Float()
	default:
		return reflect.ValueOf("")
	}

	// Làm tròn
	if decimal > 0 {
		num = float64(int(num*pow10(decimal)+0.5)) / pow10(decimal)
	} else {
		num = float64(int(num))
	}

	// Chuyển số sang string
	s := strconv.FormatFloat(num, 'f', decimal, 64)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	fracPart := ""
	if len(parts) > 1 {
		fracPart = parts[1]
	}

	// Chèn dấu phân cách hàng nghìn
	result := ""
	count := 0
	for i := len(intPart) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			result = dot + result
		}
		result = string(intPart[i]) + result
		count++
	}

	// Thêm phần thập phân nếu decimal > 0
	if decimal > 0 && fracPart != "" {
		result += point + fracPart
	}

	return reflect.ValueOf(result)
}

func pow10(n int) float64 {
	p := 1.0
	for i := 0; i < n; i++ {
		p *= 10
	}
	return p
}

func Json(reflect.Value) reflect.Value {
	return reflect.Value{}
}
