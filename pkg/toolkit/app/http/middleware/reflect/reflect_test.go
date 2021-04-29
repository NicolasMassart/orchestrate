// +build unit

package reflect_test

import (
	"context"
	"reflect"
	"testing"

	mockmid "github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/middleware/mock"
	reflectmid "github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/middleware/reflect"
	"github.com/golang/mock/gomock"
)

func TestBuilder(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()

	b := reflectmid.NewBuilder()

	type Foo struct{}
	type Bar struct{}

	fooBuilder := mockmid.NewMockBuilder(ctrlr)
	barBuilder := mockmid.NewMockBuilder(ctrlr)

	b.AddBuilder(reflect.TypeOf(&Foo{}), fooBuilder)
	b.AddBuilder(reflect.TypeOf(&Bar{}), barBuilder)

	foo := &Foo{}
	fooBuilder.EXPECT().Build(gomock.Any(), "test-foo", foo).Return(nil, nil, nil)
	_, _, _ = b.Build(context.Background(), "test-foo", foo)

	bar := &Bar{}
	barBuilder.EXPECT().Build(gomock.Any(), "test-bar", bar).Return(nil, nil, nil)
	_, _, _ = b.Build(context.Background(), "test-bar", bar)
}
