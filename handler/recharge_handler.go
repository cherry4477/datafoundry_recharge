package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/asiainfoLDP/datafoundry_recharge/api"
	"github.com/asiainfoLDP/datafoundry_recharge/common"

	"github.com/asiainfoLDP/datafoundry_recharge/models"
	"github.com/julienschmidt/httprouter"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	TransTypeDEDUCTION = "deduction"
	TransTypeRECHARGE  = "recharge"

	AdminUser = "admin"

	letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Aipayrecharge struct {
	Order_id  string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	ReturnUrl string  `json:"returnUrl"`
}

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type AipayRequestInfo struct {
	Aiurl   string      `json:"aiurl"`
	Method  string      `json:"method"`
	Payload PayloadInfo `json:"payload"`
}

type PayloadInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
	recharge.User = user
	if recharge.Namespace == "" {
		recharge.Namespace = user
	}
	recharge.TransactionId = genUUID()
	logger.Debug("recharge: %v", recharge.TransactionId)

	if recharge.Type == TransTypeDEDUCTION {
		_doDeduction(w, r, recharge, db, user)
	} else {
		_doRecharge(w, r, recharge, db)
	}

}

func AipayCallBack(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Debug("AipayCallBack begin")
}

func _doDeduction(w http.ResponseWriter, r *http.Request, recharge *models.Transaction, db *sql.DB, user string) {
	if user != AdminUser {
		logger.Warn("Only admin user can deduction! user:%v", user)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeAuthFailed, "Only admin user can deduction!"), nil)
		return
	}

	//record recharge in database
	err := models.RecordRecharge(db, recharge)
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

func _doRecharge(w http.ResponseWriter, r *http.Request, recharge *models.Transaction, db *sql.DB) {
	xmlMsg, err := GetAipayRechargeMsg(recharge)
	if err != nil {
		logger.Error("GetAipayRechargeMsg  err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeRecordRecharge, err.Error()), nil)
		return
	}

	aipayRequestInfo := &AipayRequestInfo{
		Aiurl:   os.Getenv("AIPAY_WEB_URL"),
		Method:  "POST",
		Payload: PayloadInfo{Name: "requestPacket", Value: xmlMsg},
	}

	api.JsonResult(w, http.StatusOK, nil, aipayRequestInfo)

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

	logger.Debug(fmt.Sprintf("%v", balance.Balance))

	//api.JsonResult(w, http.StatusOK, nil, balance)

}

func GetAipayRechargeMsg(recharge *models.Transaction) (xmlMsg string, err error) {
	if recharge.Type != TransTypeRECHARGE {
		return "", nil
	}
	aipayrecharge := &Aipayrecharge{Order_id: recharge.TransactionId,
		Amount: recharge.Amount, ReturnUrl: os.Getenv("RETURN_URL")}
	logger.Debug(aipayrecharge.ReturnUrl)

	body, err := json.Marshal(aipayrecharge)

	url := fmt.Sprintf("%s/bill/%s/recharge",
		os.Getenv("JAVA_AIPAY_REQUESTPACKET_URL"), recharge.Namespace)

	response, data, err := common.RemoteCallWithJsonBody("PUT", url, "", "", body)
	if err != nil {
		logger.Error("error: ", err.Error())
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		logger.Error("remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return "", fmt.Errorf("remote (%s) status code: %d.", url, response.StatusCode)
	}

	result := &Result{}
	err = json.Unmarshal(data, result)
	if err != nil {
		logger.Error("Parse body err: %v", err)
		return
	}

	xmlMsg = fmt.Sprintf("%v", result.Data)
	logger.Debug(xmlMsg)

	return
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
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
