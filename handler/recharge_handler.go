package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/asiainfoLDP/datafoundry_recharge/api"
	"github.com/asiainfoLDP/datafoundry_recharge/common"
	"github.com/asiainfoLDP/datafoundry_recharge/models"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	TransTypeDEDUCTION = "deduction"
	TransTypeRECHARGE  = "recharge"

	letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var AdminUsers = []string{"admin", "datafoundry"}

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
	Aiurl    string        `json:"aiurl"`
	Method   string        `json:"method"`
	Payloads []PayloadInfo `json:"payloads"`
}

type PayloadInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type NotifyResult struct {
	SignPayNotifyMsg string `json:"signPayNotifyMsg"`
	Order_id         string `json:"order_id"`
	Result           int    `json:"result"`
}

func init() {
	rand.Seed(time.Now().UnixNano())

	initAdminUser()
}

func DoRecharge(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: POST %v.", r.URL)

	logger.Info("Begin do recharge handler.")
	defer logger.Info("End do recharge handler.")

	//

	token := r.Header.Get("Authorization")

	r.ParseForm()
	region := r.Form.Get("region")

	user, err := getDFUserame(token, region)
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

	balance, err := models.GetBalanceByNamespace(db, recharge.Namespace)
	if err != nil {
		logger.Error("GetBalance:%v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeQUERYBALANCE, err.Error()), nil)
		return
	}
	recharge.Balance = balance.Balance

	if len(recharge.Region) == 0 {
		recharge.Region = region
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

	rbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	logger.Debug(string(rbody))

	url := fmt.Sprintf("%s/payconfirm/message",
		os.Getenv("JAVA_AIPAY_REQUESTPACKET_URL"))

	response, data, err := common.RemoteCallWithJsonBody("POST", url, "", "", rbody)
	if err != nil {
		logger.Error("error: ", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		logger.Error("remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return
	}

	logger.Debug("%s ----:%s", url, data)

	notifyResult := &NotifyResult{}
	result := &Result{Data: notifyResult}
	err = json.Unmarshal(data, result)
	if err != nil {
		logger.Error("Parse body err: %v", err)
		return
	}

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}
	w.Header().Set("charset", "utf-8")

	switch result.Code {
	case 0:
		{
			if notifyResult.Result == 0 {
				logger.Debug("aipay succeeded!")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(notifyResult.SignPayNotifyMsg))
				//update record recharge in database
				err = models.UpdateRechargeAndBalance(db, notifyResult.Order_id, "O")

			} else {
				logger.Debug("aipay failed! notifyResult.result:%d, error:%s", notifyResult.Result, result.Msg)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(notifyResult.SignPayNotifyMsg))
				//update record recharge in database
				err = models.UpdateRechargeAndBalance(db, notifyResult.Order_id, "E")
			}
		}
	case 1001:
		{
			logger.Debug("aipay failed! code:%d, error:%s", result.Code, result.Msg)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(notifyResult.SignPayNotifyMsg))
			//update record recharge in database
			err = models.UpdateRechargeAndBalance(db, notifyResult.Order_id, "E")
		}
	}

	if err != nil {
		logger.Error(err.Error())
	}

}

func Testsql(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}
	models.UpdateRechargeAndBalance(db, "LA0IIC0VVX", "O")
}

func checkAdminUser(user string) bool {
	for _, v := range AdminUsers {
		if user == v {
			return true
		}
	}
	return false
}

func _doDeduction(w http.ResponseWriter, r *http.Request, trans *models.Transaction, db *sql.DB, user string) {
	if false == checkAdminUser(user) {
		logger.Warn("Only admin user can deduction! user:%v", user)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeAuthFailed, "Only admin user can deduction!"), nil)
		return
	}

	//record recharge in database
	trans.Status = "O"
	err := models.RecordRecharge(db, trans)
	if err != nil {
		logger.Error("Record recharge err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeRecordRecharge, err.Error()), nil)
		return
	}

	balance, e := updateBalance(db, trans)
	if e != nil {
		logger.Error("udateBalance err: %v", e)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeUpdateBalance, e.Error()), nil)
		//todo rollback RecordRecharge
		if err := models.UpdateTransaction(db, trans.TransactionId, "F"); err != nil {
			logger.Error("UpdateTransaction err: %v", err)
		}
		return
	}

	api.JsonResult(w, http.StatusOK, nil, balance)

}

func CheckAmount(amount float64) uint {

	amount = float64(int64(amount*100)) * 0.01
	/*if (amount*100 - float64(int64(amount*100))) > 0 {
		logger.Error("%v, %v, %v, %v", api.ErrorCodeAmountsInvalid, amount, amount*100, float64(int(amount*100)))
		return api.ErrorCodeAmountsInvalid
	}*/
	logger.Info("amount:%v", amount)
	if amount < 0 {
		logger.Error("%v, %v", api.ErrorCodeAmountsNegative, amount)
		return api.ErrorCodeAmountsNegative
	}
	if amount > 99999999.99 {
		logger.Error("%v, %v", api.ErrorCodeAmountsTooBig, amount)
		return api.ErrorCodeAmountsTooBig
	}

	return api.ErrorCodeNone
}

func EnsureLeTowDecimal(amount float64) bool {
	var bret = false
	var sinput = strconv.FormatFloat(amount, 'f', -1, 64)

	count := strings.Count(sinput, ".")
	if count == 0 {
		bret = true
	} else if count == 1 {
		i := strings.LastIndex(sinput, ".")
		logger.Info("i:%v", i)
		logger.Info("len:%v", len(sinput))
		if i+3 < len(sinput) {
			bret = false
		} else {
			bret = true
		}
	} else if count > 1 {

	}
	return bret
}

func _doRecharge(w http.ResponseWriter, r *http.Request, recharge *models.Transaction, db *sql.DB) {
	if errcode := CheckAmount(recharge.Amount); errcode > 0 {
		api.JsonResult(w, http.StatusBadRequest, api.GetError(errcode), nil)
		return
	}

	xmlMsg, err := GetAipayRechargeMsg(recharge)
	if err != nil {
		logger.Error("GetAipayRechargeMsg  err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeGetAiPayMsg, err.Error()), nil)
		return
	}

	aipayRequestInfo := &AipayRequestInfo{
		Aiurl:    os.Getenv("AIPAY_WEB_URL"),
		Method:   "POST",
		Payloads: []PayloadInfo{{Name: "requestPacket", Value: xmlMsg}},
	}

	//record recharge in database
	recharge.Status = "I"
	recharge.Paymode = "hongpay"
	err = models.RecordRecharge(db, recharge)
	if err != nil {
		logger.Error("Record recharge err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeRecordRecharge, err.Error()), nil)
		return
	}

	api.JsonResult(w, http.StatusOK, nil, aipayRequestInfo)

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

	payerPARTNERACCTID := os.Getenv("test100")
	if payerPARTNERACCTID == "" {
		payerPARTNERACCTID = recharge.Namespace
	}

	url := fmt.Sprintf("%s/bill/%s/recharge",
		os.Getenv("JAVA_AIPAY_REQUESTPACKET_URL"), payerPARTNERACCTID)

	response, data, err := common.RemoteCallWithJsonBody("PUT", url, "", "", body)
	if err != nil {
		logger.Error("error:%s", err.Error())
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
		return nil, nil //models.RechargeBalance(db, recharge.Namespace, recharge.Amount)
	}
}

func setTransactionType(r *http.Request, transaction *models.Transaction) {

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

	region := r.Form.Get("region")

	user, err := getDFUserame(token, region)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeAuthFailed, err.Error()), nil)
		return
	}

	ns := r.Form.Get("namespace")
	if true == checkAdminUser(user) {

	} else {
		if ns == "" {
			ns = user
		} else {
			err = checkNameSpacePermission(ns, token, region)
			if err != nil {
				logger.Warn("%s cannot access the namespace:%s.", user, ns)
				api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodePermissionDenied), nil)
				return
			}
		}
	}

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodeDbNotInitlized), nil)
		return
	}

	offset, size := api.OptionalOffsetAndSize(r, 30, 1, 1000)

	orderBy := models.ValidateOrderBy(r.Form.Get("orderby"))
	sortOrder := models.ValidateSortOrder(r.Form.Get("sortorder"), models.SortOrderDesc)
	transType := models.ValidateTransType(r.Form.Get("type"))
	status := models.ValidateStatus(r.Form.Get("status"))

	count, transactions, err := models.QueryTransactionList(db, transType, ns, status, region, orderBy, sortOrder, offset, size)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeQueryTransactions, err.Error()), nil)
		return
	}

	api.JsonResult(w, http.StatusOK, nil, api.NewQueryListResult(count, transactions))
}

func CouponRecharge(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Begin do recharge handler. Request url: POST %v.", r.URL)
	defer logger.Info("End do recharge handler.")

	token := r.Header.Get("Authorization")
	r.ParseForm()
	region := r.Form.Get("region")

	user, err := getDFUserame(token, region)
	if err != nil {
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeAuthFailed, err.Error()), nil)
		return
	}
	if false == checkAdminUser(user) {
		api.JsonResult(w, http.StatusInternalServerError, api.GetError(api.ErrorCodePermissionDenied), nil)
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
	recharge.Type = "recharge"
	if len(recharge.Region) == 0 {
		recharge.Region = region
	}
	recharge.TransactionId = genUUID()
	logger.Debug("coupon recharge: %v", recharge.TransactionId)

	_doCouponRecharge(w, r, recharge, db)
}

func _doCouponRecharge(w http.ResponseWriter, r *http.Request, recharge *models.Transaction, db *sql.DB) {

	if errcode := CheckAmount(recharge.Amount); errcode > 0 {
		api.JsonResult(w, http.StatusBadRequest, api.GetError(errcode), nil)
		return
	}

	//record recharge in database
	recharge.Status = "O"
	if len(recharge.Paymode) == 0 {
		recharge.Paymode = "coupon"
	}
	balance, err := models.RechargeBalance(db, recharge.Namespace, recharge.Amount)
	if err != nil {
		logger.Error("RechargeBalance:%v", err)
		return
	}
	logger.Debug("_doCouponRecharge---RechargeBalance:%v", balance.Balance)

	recharge.Balance = balance.Balance

	if err := models.RecordRecharge(db, recharge); err != nil {
		logger.Error("Record recharge err: %v", err)
		api.JsonResult(w, http.StatusBadRequest, api.GetError2(api.ErrorCodeRecordRecharge, err.Error()), nil)
		return
	}

	api.JsonResult(w, http.StatusOK, nil, balance)
}

func genUUID() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func initAdminUser() {
	admins := os.Getenv("ADMINUSERS")
	if admins == "" {
		logger.Warn("Not set admin users.")
	}
	admins = strings.TrimSpace(admins)
	AdminUsers = strings.Split(admins, " ")
	logger.Info("Admin users: %v.", AdminUsers)
}
