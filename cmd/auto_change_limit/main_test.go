package main_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
	"math"
	"testing"
)

func TestExecConfig_RunOneOsd(t *testing.T) {
	// y = 11.533x-0.809
	spew.Dump(11.533 * math.Pow(1, -0.809))
	spew.Dump(11.533 * math.Pow(5, -0.809))
	spew.Dump(11.533 * math.Pow(11, -0.809))
	spew.Dump(11.533 * math.Pow(22, -0.809))

	logrus.Infof("%0.2f", 11.222222)
}
