package dynamic

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/ConsenSys/orchestrate/pkg/http/config/dynamic"
	"github.com/ConsenSys/orchestrate/pkg/http/handler"
	"github.com/ConsenSys/orchestrate/pkg/http/handler/healthcheck"
	"github.com/ConsenSys/orchestrate/pkg/http/handler/prometheus"
	reflecthandler "github.com/ConsenSys/orchestrate/pkg/http/handler/reflect"
	"github.com/ConsenSys/orchestrate/pkg/http/handler/swagger"
)

type Builder struct {
	reflect *reflecthandler.Builder
}

func NewBuilder() *Builder {
	b := &Builder{
		reflect: reflecthandler.NewBuilder(),
	}

	b.AddBuilder(reflect.TypeOf(&dynamic.HealthCheck{}), healthcheck.NewTraefikBuilder())
	b.AddBuilder(reflect.TypeOf(&dynamic.Swagger{}), swagger.NewBuilder())
	b.AddBuilder(reflect.TypeOf(&dynamic.Prometheus{}), prometheus.NewBuilder(nil))

	return b
}

func (b *Builder) AddBuilder(typ reflect.Type, builder handler.Builder) {
	b.reflect.AddBuilder(typ, builder)
}

func (b *Builder) Build(ctx context.Context, name string, configuration interface{}, respModifier func(*http.Response) error) (http.Handler, error) {
	cfg, ok := configuration.(*dynamic.Service)
	if !ok {
		return nil, fmt.Errorf("invalid configuration type (expected %T but got %T)", cfg, configuration)
	}

	field, err := cfg.Field()
	if err != nil {
		return nil, err
	}

	return b.reflect.Build(ctx, name, field, respModifier)
}
