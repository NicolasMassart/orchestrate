package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPtr(t *testing.T) {
	_ = ToPtr("a").(*string)
	_ = ToPtr(1).(*int)
	_ = ToPtr([]string{"a"}).(*[]string)
}

func TestCopyPtr(t *testing.T) {
	oldCar := &struct {
		make    string
		model   string
		mileage int
	}{
		make:    "Ford",
		model:   "Taurus",
		mileage: 200000,
	}

	newCar := &struct {
		make    string
		model   string
		mileage int
	}{}

	CopyPtr(oldCar, newCar)
	assert.Equal(t, oldCar.make, newCar.make)
	assert.Equal(t, oldCar.model, newCar.model)
	assert.Equal(t, oldCar.mileage, newCar.mileage)
}
