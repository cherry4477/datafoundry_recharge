package handler

import (
	//"encoding/json"
	"github.com/asiainfoLDP/datafoundry_recharge/api"
	"github.com/asiainfoLDP/datafoundry_recharge/models"
	"github.com/julienschmidt/httprouter"
	//"io/ioutil"
	"fmt"
	"net/http"
	//"strconv"
	//"time"
)

func Balance(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: POST %v.", r.URL)

	logger.Info("Begin do recharge handler.")
	defer logger.Info("End do recharge handler.")

	//todo create in database

	dotestbalance()

	api.JsonResult(w, http.StatusOK, nil, nil)
}

func dotestbalance() {
	db := models.GetDB()

	balance, err := models.GetBalanceByNamespace(db, "chaizs")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("#%#v\n", balance)
	}

	balance, err = models.RechargeBalance(db, "chaizs", 12.34)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("####%#v\n", balance)
	}
	balance, err = models.RechargeBalance(db, "yuanwm", 12.34)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("###%#v\n", balance)
	}
	balance, err = models.DeductionBalance(db, "liuxu", 12.34)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("###%#v\n", balance)
	}
}
