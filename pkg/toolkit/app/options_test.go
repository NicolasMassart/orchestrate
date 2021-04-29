package app

import (
	"context"
	"reflect"
	"testing"
	"time"

	mockauth "github.com/ConsenSys/orchestrate/pkg/toolkit/app/auth/mock"
	"github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/config/dynamic"
	mockprovider "github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/configwatcher/provider/mock"
	mockhandler "github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/handler/mock"
	mockmid "github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/middleware/mock"
	dynrouter "github.com/ConsenSys/orchestrate/pkg/toolkit/app/http/router/dynamic"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderOpt(t *testing.T) {
	prvdr := mockprovider.New()

	app, err := New(newTestConfig())
	assert.NoError(t, err, "Creating app should not error")

	err = ProviderOpt(prvdr)(app)
	assert.NoError(t, err, "Applying option should not error")

	var listened bool
	app.AddListener(func(context.Context, interface{}) error {
		listened = true
		return nil
	})

	err = app.Start(context.Background())
	require.NoError(t, err, "App should have started properly")

	// Wait for application to properly start
	time.Sleep(100 * time.Millisecond)

	// Provide a message
	_ = prvdr.ProvideMsg(context.Background(), dynamic.NewMessage("test", &dynamic.Configuration{}))

	// Wait for application to properly process provided message
	time.Sleep(100 * time.Millisecond)

	// Stop app
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = app.Stop(ctx)
	assert.NoError(t, err, "App should have stop properly")

	assert.True(t, listened, "Listener should have been called")
}

func TestMiddlewareOpt(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()

	midBuilder := mockmid.NewMockBuilder(ctrlr)

	app, err := New(newTestConfig())
	assert.NoError(t, err, "Creating app should not error")

	err = MiddlewareOpt(
		reflect.TypeOf(&dynamic.Mock{}),
		midBuilder,
	)(app)
	assert.NoError(t, err, "Applying option should not error")

	testCfg := &dynamic.Mock{}
	cfg := &dynamic.Middleware{
		Mock: testCfg,
	}
	midBuilder.EXPECT().Build(gomock.Any(), "test", testCfg)
	_, _, err = app.HTTP().(*dynrouter.Builder).Middleware.Build(context.Background(), "test", cfg)
	assert.NoError(t, err, "Building middleware should not error")
}

func TestHandlerOpt(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()

	handlerBuilder := mockhandler.NewMockBuilder(ctrlr)

	app, err := New(newTestConfig())
	assert.NoError(t, err, "Creating app should not error")

	err = HandlerOpt(
		reflect.TypeOf(&dynamic.Mock{}),
		handlerBuilder,
	)(app)
	assert.NoError(t, err, "Applying option should not error")

	testCfg := &dynamic.Mock{}
	cfg := &dynamic.Service{
		Mock: testCfg,
	}
	handlerBuilder.EXPECT().Build(gomock.Any(), "test", testCfg, nil)
	_, err = app.HTTP().(*dynrouter.Builder).Handler.Build(context.Background(), "test", cfg, nil)
	assert.NoError(t, err, "Building middleware should not error")
}

func TestMultitenancyOpt(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()

	jwtChecker := mockauth.NewMockChecker(ctrlr)
	keyChecker := mockauth.NewMockChecker(ctrlr)

	opt := MultiTenancyOpt("auth", jwtChecker, keyChecker, true)

	_, err := New(newTestConfig(), opt)
	assert.NoError(t, err, "Creating app should not error")
}

func TestLoggerOpt(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()

	opt := LoggerOpt("test")

	_, err := New(newTestConfig(), opt)
	assert.NoError(t, err, "Creating app should not error")
}

func TestSwaggerOpt(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()

	opt := SwaggerOpt("test")

	_, err := New(newTestConfig(), opt)
	assert.NoError(t, err, "Creating app should not error")
}

func TestMetricsOpt(t *testing.T) {
	ctrlr := gomock.NewController(t)
	defer ctrlr.Finish()

	opt := MetricsOpt()

	testCfg := newTestConfig()
	_, err := New(testCfg, opt)
	assert.NoError(t, err, "Creating app should not error")
}
