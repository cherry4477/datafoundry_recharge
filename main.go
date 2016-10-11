package main

import (
	"fmt"
	"github.com/asiainfoLDP/datafoundry_recharge/api"
	"github.com/asiainfoLDP/datafoundry_recharge/log"
	"github.com/asiainfoLDP/datafoundry_recharge/models"
	"github.com/asiainfoLDP/datafoundry_recharge/router"
	"github.com/asiainfoLDP/datahub_commons/httputil"
	"net/http"
	"os"
	"strings"
	"time"
)

const SERVERPORT = 8090

var (
	logger = log.GetLogger()

	loglevel = os.Getenv("LOG_LEVEL")
	//init a router
	initRouter = router.InitRouter()
)

type Service struct {
	httpPort int
}

func newService(httpPort int) *Service {
	service := &Service{
		httpPort: httpPort,
	}

	return service
}

func main() {

	//new a router
	router.NewRouter(initRouter)

	//todo init db
	models.InitDB()

	service := newService(SERVERPORT)
	address := fmt.Sprintf(":%d", service.httpPort)
	logger.Debug("address: %v", address)

	logger.Info("Listening http at: %s", address)
	err := http.ListenAndServe(address, httputil.TimeoutHandler(initRouter, 250*time.Millisecond, ""))
	if err != nil {
		logger.Error("http listen and server err: %v", err)
		return
	}

	return
}

func init() {
	if strings.ToUpper(loglevel) == "DEBUG" {
		log.SetDebug = true
	}

	api.InitMQ()
}
