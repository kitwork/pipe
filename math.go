package pipe

import "reflect"

func Add(vals ...reflect.Value) reflect.Value {
	ln := len(vals)
	if ln == 0 {
		return RV(float64(0))
	}
	// left is last (accumulator)
	left := toFloatRV(vals[ln-1])
	// add all previous
	for i := 0; i < ln-1; i++ {
		left += toFloatRV(vals[i])
	}
	return RV(left)
}

func Sub(vals ...reflect.Value) reflect.Value {
	ln := len(vals)
	if ln == 0 {
		return RV(float64(0))
	}
	// left is last (accumulator)
	left := toFloatRV(vals[ln-1])
	// subtract all previous in order
	for i := 0; i < ln-1; i++ {
		left -= toFloatRV(vals[i])
	}
	return RV(left)
}

func Mul(vals ...reflect.Value) reflect.Value {
	ln := len(vals)
	if ln == 0 {
		return RV(float64(0))
	}
	left := toFloatRV(vals[ln-1])
	for i := 0; i < ln-1; i++ {
		left *= toFloatRV(vals[i])
	}
	return RV(left)
}

func Div(vals ...reflect.Value) reflect.Value {
	ln := len(vals)
	if ln == 0 {
		return RV(float64(0))
	}
	left := toFloatRV(vals[ln-1])
	for i := 0; i < ln-1; i++ {
		d := toFloatRV(vals[i])
		if d == 0 {
			// tránh panic / inf: trả 0 (hoặc bạn có thể chọn trả math.Inf)
			return RV(float64(0))
		}
		left /= d
	}
	return RV(left)
}

func Nullish(vals ...reflect.Value) reflect.Value {
	if vals[0].IsZero() || !vals[0].IsValid() {
		return vals[1]
	}
	return vals[0]
}

func Ternary(vals ...reflect.Value) reflect.Value {
	cond := vals[0]
	var ok bool

	switch cond.Kind() {
	case reflect.Bool:
		ok = cond.Bool()
	default:
		ok = toFloatRV(cond) != 0
	}

	if ok {
		return vals[1]
	}
	return vals[2]
}
