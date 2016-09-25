package handler

import (
	//"encoding/json"
	"github.com/asiainfoLDP/datafoundry_recharge/api"
	"github.com/asiainfoLDP/datafoundry_recharge/log"
	//"github.com/asiainfoLDP/datafoundry_recharge/models"
	"github.com/julienschmidt/httprouter"
	//"io/ioutil"
	"net/http"
	//"strconv"
	//"time"
)

var logger = log.GetLogger()

func DoRecharge(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: POST %v.", r.URL)

	logger.Info("Begin do recharge handler.")
	defer logger.Info("End do recharge handler.")

	//todo create in database

	api.JsonResult(w, http.StatusOK, nil, nil)
}
