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
	logger.Info("Request url: GET %v.", r.URL)

	logger.Info("Begin do recharge handler.")

	r.ParseForm()

	token := r.Header.Get("Authorization")

	user, err := getDFUserame(token)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeAuthFailed, err.Error()), nil)
		return
	}

	db := models.GetDB()
	balance, err := models.GetBalanceByNamespace(db, user)

	defer logger.Info("End do recharge handler.")

	//todo create in database

	api.JsonResult(w, http.StatusOK, nil, balance)
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
