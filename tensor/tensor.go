package tensor

import (
	"reflect"
)

type dimension struct{ Channels, Height, Width int }

func (d dimension) Volume() int              { return d.Channels * d.Width * d.Height }
func (d dimension) Dimension() (c, h, w int) { return d.Channels, d.Height, d.Width }

type tensor interface {
	Dimension() (c, h, w int)
	Volume() int
	Type() reflect.Type
	Magic() byte
	Values() interface{}
	Index(index int) interface{}
	ConvertElem(val string, index int) error
	HotOne() int
	Floats32(copy ...bool) []float32
	CopyTo(interface{})
}

type Tensor struct{ tensor }

/*


type tensor64f struct {
	dimension
	values []float64
}


type tensori struct {
	dimension
	values []int
}

type tensor8f struct {
	dimension
	values []fu.Fixed8
}


func (t tensor64f) ConvertElem(val string, index int) (err error) {
	t.values[index], err = strconv.ParseFloat(val, 64)
	return
}

func (t tensori) ConvertElem(val string, index int) (err error) {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return
	}
	t.values[index] = int(i)
	return
}

func (t tensor8f) ConvertElem(val string, index int) (err error) {
	t.values[index], err = fu.Fast8f(val)
	return
}


func (t tensori) Index(index int) interface{}   { return t.values[index] }
func (t tensor8f) Index(index int) interface{}  { return t.values[index] }

func (t tensor64f) Index(index int) interface{} { return t.values[index] }

func (t tensori) Values() interface{}   { return t.values }
func (t tensor64f) Values() interface{} { return t.values }

func (t tensori) Type() reflect.Type   { return Int }
func (t tensor8f) Type() reflect.Type  { return Fixed8Type }
func (t tensor8f) Values() interface{}  { return t.values }
func (t tensor64f) Type() reflect.Type { return Float64 }

func (t tensori) Magic() byte   { return 'i' }
func (t tensor64f) Magic() byte { return 'F' }

func (t tensori) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j] < v {
			j = i
		}
	}
	return
}

func (t tensor8f) Magic() byte  { return '8' }
func (t tensor8f) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j].int8 < v.int8 {
			j = i
		}
	}
	return
}



func (t tensor64f) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j] < v {
			j = i
		}
	}
	return
}

func (t tensori) Extract(r []reflect.Value) {
	for i, v := range t.values {
		r[i] = reflect.ValueOf(v)
	}
}

func (t tensori) Floats32(...bool) (r []float32) {
	r = make([]float32, len(t.values))
	for i, v := range t.values {
		r[i] = float32(v)
	}
	return
}

func (t tensor8f) Extract(r []reflect.Value) {
	for i, v := range t.values {
		r[i] = reflect.ValueOf(v)
	}
}

func (t tensor8f) Floats32(...bool) (r []float32) {
	r = make([]float32, len(t.values))
	for i, v := range t.values {
		r[i] = v.Float32()
	}
	return
}


func (t tensor64f) Extract(r []reflect.Value) {
	for i, v := range t.values {
		r[i] = reflect.ValueOf(v)
	}
}

func (t tensor64f) Floats32(...bool) (r []float32) {
	r = make([]float32, len(t.values))
	for i, v := range t.values {
		r[i] = float32(v)
	}
	return
}




//	gets base64-encoded compressed stream as a string prefixed by \xE2\x9C\x97` (✗`)
func DecodeTensor(string) (t Tensor, err error) {
	return
}

func (t Tensor) Width() int {
	_, _, w := t.Dimension()
	return w
}

func (t Tensor) Height() int {
	_, h, _ := t.Dimension()
	return h
}

func (t Tensor) Depth() int {
	c, _, _ := t.Dimension()
	return c
}

func (t Tensor) String() (str string) {
	return t.Encode(false)
}

func (t Tensor) Encode(compress bool) (str string) {
	//t.Magic()
	//t.Dimension()
	//t.Values()
	//gzip => base64
	return
}

func MakeFloat64Tensor(channels, height, width int, values []float64, docopy ...bool) Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]float64, len(values))
			copy(v, values)
		}
	} else {
		v = make([]float64, channels*height*width)
	}
	x := tensor64f{
		dimension: dimension{
			Channels: channels,
			Height:   height,
			Width:    width,
		},
		values: v,
	}
	return Tensor{x}
}



func MakeFixed8Tensor(channels, height, width int, values []Fixed8, docopy ...bool) Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]Fixed8, len(values))
			copy(v, values)
		}
	} else {
		v = make([]Fixed8, channels*height*width)
	}
	x := tensor8f{
		dimension: dimension{
			Channels: channels,
			Height:   height,
			Width:    width},
		values: v,
	}
	return Tensor{x}
}

func MakeIntTensor(channels, height, width int, values []int, docopy ...bool) Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]int, len(values))
			copy(v, values)
		}
	} else {
		v = make([]int, channels*height*width)
	}
	x := tensori{
		dimension: dimension{
			Channels: channels,
			Height:   height,
			Width:    width},
		values: v,
	}
	return Tensor{x}
}

var TensorType = reflect.TypeOf(Tensor{})

*/