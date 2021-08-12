package line

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLeastSquares(t *testing.T) {
	/*
		1	5.919042797
		1	23.33703046
		1	25.14476282
		5	36.06610548
		11	53.49233359
		22	86.56219916
	*/
	x := []float64{1, 1, 1, 5, 11, 22}
	y := []float64{5.919042797, 23.33703046, 25.14476282, 36.06610548, 53.49233359, 86.56219916}
	res, err := LeastSquares(x, y)
	require.NoError(t, err)
	spew.Dump(res)
}
