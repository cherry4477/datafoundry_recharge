package handler

import (
	//"encoding/json"
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/asiainfoLDP/datafoundry_recharge/api"
	"github.com/asiainfoLDP/datafoundry_recharge/common"

	"github.com/asiainfoLDP/datafoundry_recharge/models"
	"github.com/julienschmidt/httprouter"
	//"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"time"
)

func DoRecharge(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: POST %v.", r.URL)

	logger.Info("Begin do recharge handler.")
	defer logger.Info("End do recharge handler.")

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	recharge := &models.Transaction{}
	err := common.ParseRequestJsonInto(r, recharge)
	if err != nil {
		logger.Error("Parse body err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}

	setTransactionType(r, recharge)

	recharge.TransactionId = genUUID()
	logger.Debug("recharge: %v", recharge.TransactionId)

	//record recharge in database
	err = models.RecordRecharge(db, recharge)
	if err != nil {
		logger.Error("Record recharge err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeRecordRecharge, err.Error()), nil)
		return
	}

	balance, e := updateBalance(db, recharge)
	if e != nil {
		logger.Error("udateBalance err: %v", e)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeRecordRecharge, err.Error()), nil)
		//todo rollback RecordRecharge

		return
	}

	api.JsonResult(w, http.StatusOK, nil, balance)
}

func updateBalance(db *sql.DB, recharge *models.Transaction) (*models.Balance, error) {
	if recharge.Type == "deduction" {
		return models.DeductionBalance(db, recharge.Namespace, recharge.Amount)
	} else {
		return models.RechargeBalance(db, recharge.Namespace, recharge.Amount)
	}
}

func setTransactionType(r *http.Request, transaction *models.Transaction) {
	r.ParseForm()
	transType := r.Form.Get("type")
	logger.Debug("Transcation type in url is:%v", transType)

	if transType == "deduction" {
		transaction.Type = "deduction"
	} else {
		transaction.Type = "recharge"
	}
}

func genUUID() string {
	mathrand.Seed(time.Now().UnixNano())

	bs := make([]byte, 16)
	_, err := rand.Read(bs)
	if err != nil {
		logger.Warn("genUUID error: ", err.Error())

		mathrand.Read(bs)
	}

	return fmt.Sprintf("%X-%X-%X", bs[0:4], bs[4:8], bs[8:])
}
