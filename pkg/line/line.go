package line

import (
	"errors"
	"fmt"
	"strconv"
)

// 输入: []float64{}
/*
1	5.919042797
1	23.33703046
1	25.14476282
5	36.06610548
11	53.49233359
22	86.56219916
*/

type LineMetaData struct {
	// a是斜率，b是截距
	A         float64
	B         float64
	DataCount int64
}

// 输出: 3.2658x+16.104
func LeastSquares(x []float64, y []float64) (LineMetaData, error) {
	// x是横坐标数据,y是纵坐标数据
	// a是斜率，b是截距
	xi := float64(0)
	x2 := float64(0)
	yi := float64(0)
	xy := float64(0)

	a := float64(0)
	b := float64(0)
	if len(x) != len(y) {
		return LineMetaData{}, errors.New("len(x)!= len(y)")
	} else {
		length := float64(len(x))
		for i := 0; i < len(x); i++ {
			xi += x[i]
			x2 += x[i] * x[i]
			yi += y[i]
			xy += x[i] * y[i]
		}
		a = (yi*xi - xy*length) / (xi*xi - x2*length)
		b = (yi*x2 - xy*xi) / (x2*length - xi*xi)
	}
	a, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", a), 64)
	b, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", b), 64)
	return LineMetaData{
		A:         a,
		B:         b,
		DataCount: int64(len(x)),
	}, nil
}
