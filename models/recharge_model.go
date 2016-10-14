package models

import (
	"database/sql"
	"fmt"
	"strings"

	"time"
)

const (
	SortOrderDesc = "desc"
	SortOrderAsc  = "asc"
)

type Transaction struct {
	TransactionId string    `json:"transactionId"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	Namespace     string    `json:"namespace"`
	User          string    `json:"user,omitempty"`
	Reason        string    `json:"reason,omitempty"`
	CreateTime    time.Time `json:"createtime,omitempty"`
	Status        string    `json:"status,omitempty"`
	StatusTime    time.Time `json:"statustime,omitempty"`
}

func RecordRecharge(db *sql.DB, rechargeInfo *Transaction) error {
	logger.Info("Model begin record recharge")
	defer logger.Info("Model end record recharge")

	nowstr := time.Now().Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`insert into DF_TRANSACTION (
				TRANSACTION_ID, TYPE, AMOUNT, NAMESPACE, USER, REASON, 
				CREATE_TIME, STATUS, STATUS_TIME
				) values (
				?, ?, ?, ?, ?, ?,
				'%s', '%s', '%s')`,
		nowstr, "I", nowstr)

	_, err := db.Exec(sqlstr,
		rechargeInfo.TransactionId, rechargeInfo.Type, rechargeInfo.Amount,
		rechargeInfo.Namespace, rechargeInfo.User, rechargeInfo.Reason)

	return err
}

func QueryTransactionList(db *sql.DB, transType, namespace, status, orderBy, sortOrder string,
	offset int64, limit int) (int64, []*Transaction, error) {

	logger.Debug("QueryTransactions begin")

	sqlParams := make([]interface{}, 0, 3)
	sqlwhere := ""
	if status != "" {
		if sqlwhere == "" {
			sqlwhere = "status=?"
		} else {
			sqlwhere = sqlwhere + " and status=?"
		}
		sqlParams = append(sqlParams, status)
	}

	if transType != "" {
		if sqlwhere == "" {
			sqlwhere = "type=?"
		} else {
			sqlwhere = sqlwhere + " and type=?"
		}
		sqlParams = append(sqlParams, transType)
	}

	if namespace != "" {
		if sqlwhere == "" {
			sqlwhere = "namespace=?"
		} else {
			sqlwhere = sqlwhere + " and namespace=?"
		}
		sqlParams = append(sqlParams, namespace)
	}

	sqlorder := ""
	if orderBy != "" {
		sqlorder = fmt.Sprintf(" order by %s %s", orderBy, sortOrder)
	}

	count, err := queryTransactionsCount(db, sqlwhere, sqlParams...)
	if err != nil {
		logger.Error(err.Error())
		return 0, nil, err
	}

	validateOffsetAndLimit(count, &offset, &limit)

	trans, err := queryTransactions(db,
		sqlwhere, sqlorder,
		limit, offset, sqlParams...)

	return count, trans, err
}

func ValidateSortOrder(sortOrder string, defaultOrder string) string {
	switch strings.ToLower(sortOrder) {
	case SortOrderAsc:
		return SortOrderAsc
	case SortOrderDesc:
		return SortOrderDesc
	}

	return defaultOrder
}

func ValidateOrderBy(orderBy string) string {
	switch orderBy {
	case "createtime":
		return "CREATE_TIME"
	}
	return ""
}

func ValidateTransType(transtype string) string {
	switch transtype {
	case "deduction":
		return "deduction"
	case "recharge":
		return "recharge"
	}

	return ""
}

func ValidateStatus(status string) string {
	switch strings.ToUpper(status) {
	case "O":
		return "O"
	case "I":
		return "I"
	case "ALL":
		return ""
	default:
		return "O"
	}

}

func queryTransactionsCount(db *sql.DB, sqlwhere string, sqlParams ...interface{}) (int64, error) {

	count := int64(0)

	sqlwhereall := ""
	if sqlwhere != "" {
		sqlwhereall = fmt.Sprintf("where %s", sqlwhere)
	}
	sqlstr := fmt.Sprintf(`select COUNT(*) from DF_TRANSACTION %s `, sqlwhereall)
	logger.Debug(">>>\n"+
		"	%s", sqlstr)
	err := db.QueryRow(sqlstr, sqlParams...).Scan(&count)

	return count, err
}

func queryTransactions(db *sql.DB, sqlwhere, sqlorder string,
	limit int, offset int64, sqlParams ...interface{}) ([]*Transaction, error) {

	logger.Info("Model begin queryTransactions")
	defer logger.Info("Model end queryTransactions")

	sqlwhereall := ""
	if sqlwhere != "" {
		sqlwhereall = fmt.Sprintf("where %s", sqlwhere)
	}
	sqlstr := fmt.Sprintf(`SELECT TRANSACTION_ID, TYPE, 
		AMOUNT, NAMESPACE, USER, REASON, CREATE_TIME, STATUS,  STATUS_TIME
		FROM DF_TRANSACTION 
		%s 
		%s 
		LIMIT %d OFFSET %d`,
		sqlwhereall,
		sqlorder,
		limit, offset)

	logger.Info(">>> %v", sqlstr)
	rows, err := db.Query(sqlstr, sqlParams...)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	trans := make([]*Transaction, 0, 32)
	for rows.Next() {
		tran := &Transaction{}
		err := rows.Scan(&tran.TransactionId, &tran.Type, &tran.Amount, &tran.Namespace,
			&tran.User, &tran.Reason, &tran.CreateTime, &tran.Status, &tran.StatusTime)
		if err != nil {
			return nil, err
		}
		trans = append(trans, tran)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return trans, nil
}

func validateOffsetAndLimit(count int64, offset *int64, limit *int) {
	if *limit < 1 {
		*limit = 1
	}
	if *offset >= count {
		*offset = count - int64(*limit)
	}
	if *offset < 0 {
		*offset = 0
	}
	if *offset+int64(*limit) > count {
		*limit = int(count - *offset)
	}
}
