package pipe

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// New creates a completely new template.FuncMap by copying from the global `functions` map.
// Use this when you want to run multiple independent template engines,
// or add new pipes without affecting the original map.
// Each call to New() returns a separate, independent map.
func New() template.FuncMap { // copy map / clone map
	funcs := template.FuncMap{}
	for k, v := range functions {
		funcs[k] = v
	}
	return funcs
}

// Functions returns the global `functions` map.
// Use this when you only need a single template engine, with static functions,
// and you won't be adding new pipes at runtime. This map is shared across all usages.
func Functions() template.FuncMap { // map global
	return functions
}

var functions = template.FuncMap{
	"json":     Json,
	"thousand": Thousand,
	"dollar":   Dollar,
}

// Thousand formats a number with thousand separators and optional decimal places.
// Accepts flexible parameters:
//   - vals[0] (optional) string: thousand separator (default ".")
//   - vals[1] (optional) int: number of decimal places (default 0)
//   - vals[2] (optional) string: decimal point character (default ",")
//   - vals[last] the main value to format (int, uint, float, or string)
func Thousand(vals ...reflect.Value) reflect.Value {
	if len(vals) == 0 {
		return reflect.ValueOf("")
	}

	val := vals[len(vals)-1] // the main value to format

	// Defaults
	dot := "."   // thousand separator
	decimal := 0 // number of decimal digits
	point := "," // decimal separator

	// Optional parameters
	if len(vals) > 1 {
		if s, ok := vals[0].Interface().(string); ok {
			dot = s
		}
	}
	if len(vals) > 2 {
		if n, ok := vals[1].Interface().(int); ok && n >= 0 {
			decimal = n
		}
	}
	if len(vals) > 3 {
		if p, ok := vals[2].Interface().(string); ok {
			point = p
		}
	}

	// Convert val to float64
	var num float64
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num = float64(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num = float64(val.Uint())
	case reflect.Float32, reflect.Float64:
		num = val.Float()
	case reflect.String:
		f, err := strconv.ParseFloat(val.String(), 64)
		if err != nil {
			return reflect.ValueOf("")
		}
		num = f
	default:
		return reflect.ValueOf("")
	}

	// Round the number
	num = math.Round(num*pow10(decimal)) / pow10(decimal)

	// Convert to string
	s := strconv.FormatFloat(num, 'f', decimal, 64)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	fracPart := ""
	if len(parts) > 1 {
		fracPart = parts[1]
	}

	// Insert thousand separators
	result := ""
	count := 0
	for i := len(intPart) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			result = dot + result
		}
		result = string(intPart[i]) + result
		count++
	}

	// Append decimal part if needed
	if decimal > 0 && fracPart != "" {
		result += point + fracPart
	}

	return reflect.ValueOf(result)
}

// pow10 returns 10^n
func pow10(n int) float64 {
	p := 1.0
	for i := 0; i < n; i++ {
		p *= 10
	}
	return p
}

// Dollar formats a value as a US dollar string, e.g., $1,234.56
func Dollar(val reflect.Value) reflect.Value {
	// Use Thousand with fixed parameters for dollars
	dot := reflect.ValueOf(".")   // thousand separator
	decimal := reflect.ValueOf(2) // decimal digits
	point := reflect.ValueOf(",") // decimal separator

	num := Thousand(dot, decimal, point, val)

	return reflect.ValueOf("$" + num.String())
}

// Json placeholder function, returns empty reflect.Value

func Json(val reflect.Value) reflect.Value {
	data := val.Interface()

	// Marshal sang JSON string
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("json.Marshal error:", err)
		return reflect.Value{}
	}

	// Trả về reflect.Value chứa string JSON
	return reflect.ValueOf(string(b))
}
