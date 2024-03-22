package main

import (
	"github.com/CloudNativeGame/aigc-gateway/pkg/autofree"
	"github.com/CloudNativeGame/aigc-gateway/pkg/options"
	"github.com/CloudNativeGame/aigc-gateway/pkg/routers"
	"github.com/CloudNativeGame/aigc-gateway/pkg/signals"
	"github.com/CloudNativeGame/aigc-gateway/pkg/storage"
	"github.com/CloudNativeGame/aigc-gateway/pkg/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/logto-io/go/client"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/wait"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"os"
	"time"
)

var logFlushFreq = pflag.Duration("log-flush-frequency", 5*time.Second, "Maximum number of seconds between log flushes")

func main() {
	klog.InitFlags(nil)

	serverOpts := options.NewServerOption()
	serverOpts.AddFlags(pflag.CommandLine)

	cliflag.InitFlags()

	go wait.Until(klog.Flush, *logFlushFreq, wait.NeverStop)
	defer klog.Flush()

	ctx := signals.SetupSignalContext()

	router := gin.Default()
	// load templates
	router.Delims("{[{", "}]}")
	router.LoadHTMLGlob("aigc-dashboard/dist/*.html")
	router.Use(static.Serve("/assets", static.LocalFile("aigc-dashboard/dist/assets", true)))

	endpoint := os.Getenv("Endpoint")
	logtoConfig := &client.LogtoConfig{

		Endpoint:  endpoint,
		AppId:     os.Getenv("App_Id"),
		AppSecret: os.Getenv("App_Secret"),
		Scopes:    []string{"email", "custom_data"},
	}
	// We use memory-based session in this example
	store := memstore.NewStore([]byte("your session secret"))
	store.Options(sessions.Options{
		Domain: utils.GetDomainFromEndpoint(endpoint),
		Path:   "/",
		MaxAge: 604800,
	})
	router.Use(sessions.Sessions("logto-session", store))
	routers.RegisterSignRouters(router, logtoConfig)
	routers.RegisterResourceRouters(router, logtoConfig)

	storage.Initialize(serverOpts)
	autofree.Run(ctx, serverOpts)

	router.Run(":8090")
}
