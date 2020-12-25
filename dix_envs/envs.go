package dix_envs

import (
	"expvar"
	"os"
	"strconv"
	"strings"

	"github.com/pubgo/dix/internal/envs"
)

const (
	prefix = "DIX_"
	Trace  = "DIX_TRACE"
)

func IsTrace() bool { return envs.IsTrace }
func SetTrace()     { _ = os.Setenv(Trace, "true"); envs.IsTrace = true }

func init() {
	if env := os.Getenv(Trace); env != "" {
		envs.IsTrace, _ = strconv.ParseBool(os.Getenv(strings.ToUpper(env)))
	}

	expvar.Publish("dix", expvar.Func(func() interface{} {
		var data []string
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, prefix) {
				data = append(data, env)
			}
		}
		return data
	}))
}
