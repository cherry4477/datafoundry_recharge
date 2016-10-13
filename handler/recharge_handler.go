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

const (
	TransTypeDEDUCTION = "deduction"
	TransTypeRECHARGE  = "recharge"

	AdminUser = "admin"
)

func DoRecharge(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: POST %v.", r.URL)

	logger.Info("Begin do recharge handler.")
	defer logger.Info("End do recharge handler.")

	//

	token := r.Header.Get("Authorization")

	user, err := getDFUserame(token)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeAuthFailed, err.Error()), nil)
		return
	}

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	recharge := &models.Transaction{}
	err = common.ParseRequestJsonInto(r, recharge)
	if err != nil {
		logger.Error("Parse body err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}

	setTransactionType(r, recharge)

	if recharge.Type == TransTypeDEDUCTION && user != AdminUser {
		logger.Warn("Only admin user can deduction! user:%v", user)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeAuthFailed, "Only admin user can deduction!"), nil)
	}

	recharge.User = user
	if recharge.Namespace == "" {
		recharge.Namespace = user
	}
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
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeUpdateBalance, e.Error()), nil)
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

func GetRechargeList(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: GET %v.", r.URL)

	logger.Info("Begin get recharge handler.")
	defer logger.Info("End get recharge handler.")

	r.ParseForm()

	token := r.Header.Get("Authorization")

	user, err := getDFUserame(token)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeAuthFailed, err.Error()), nil)
		return
	}

	userparam := ""
	if user == AdminUser {
		userparam = r.Form.Get("username")
	} else {
		userparam = user
	}

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	offset, size := api.OptionalOffsetAndSize(r, 30, 1, 100)

	orderBy := models.ValidateOrderBy(r.Form.Get("orderby"))
	sortOrder := models.ValidateSortOrder(r.Form.Get("sortorder"), models.SortOrderDesc)
	transType := models.ValidateTransType(r.Form.Get("type"))

	count, transactions, err := models.QueryTransactionList(db, transType, userparam, orderBy, sortOrder, offset, size)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeQueryTransactions, err.Error()), nil)
		return
	}

	api.JsonResult(w, http.StatusOK, nil, api.NewQueryListResult(count, transactions))
}

func genUUID() string {
	mathrand.Seed(time.Now().UnixNano())

	bs := make([]byte, 12)
	_, err := rand.Read(bs)
	if err != nil {
		logger.Warn("genUUID error: ", err.Error())

		mathrand.Read(bs)
	}

	return fmt.Sprintf("%X-%X-%X", bs[0:4], bs[4:8], bs[8:])
}
