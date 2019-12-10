package grpcserver

import (
	"runtime"

	log "github.com/sirupsen/logrus"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
)

// RecoverPanicHandler functions used by gRPC interceptor to recover panic
func RecoverPanicHandler(p interface{}) error {
	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)
	log.Errorf("panic recovered:%+v", string(buf))
	return errors.InternalError("%s", p)
}