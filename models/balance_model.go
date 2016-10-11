package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Balance struct {
	Namespace string  `json:"namespace"`
	CreateAt  string  `json:"create_at"`
	UpdateAt  string  `json:"update_at"`
	Balance   float64 `json:"balance"`
	Status    string  `json:"state,omitempty"`
}

func GetBalanceByNamespace(db *sql.DB, ns string) (*Balance, error) {
	balance := new(Balance)
	var err error

	err = db.QueryRow(`SELECT 
		balance, 
		state
		FROM DF_balance 
		WHERE 
		namespace=?`,
		ns).Scan(
		&balance.Balance,
		&balance.Status)

	if err == sql.ErrNoRows {
		CreateNamespace(db, ns)
	} else if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	logger.Debug("%#v", balance)

	balance.Namespace = ns

	return balance, nil

}

func CreateNamespace(db *sql.DB, ns string) (err error) {

	if _, err = db.Exec(`INSERT INTO DF_balance
			(namespace) VALUES(?)`, ns); err != nil {
		logger.Error("INSERT INTO DF_balance error:", err.Error())

	}
	return err
}

func UpdateBalance(db *sql.DB, balance *Balance) (*Balance, error) {

	sqlstr := fmt.Sprintf(`update DF_balance SET balance = '%v' where namespace = '%v'`, balance.Balance, balance.Namespace)

	_, err := db.Exec(sqlstr)

	if err == sql.ErrNoRows {
		logger.Error("no such rows.")
	} else if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return GetBalanceByNamespace(db, balance.Namespace)

}

func RechargeBalance(db *sql.DB, ns string, amount float64) (*Balance, error) {

	balance, err := GetBalanceByNamespace(db, ns)
	if err != nil {
		return nil, err
	}
	balance.Balance += amount
	return UpdateBalance(db, balance)

}

func DeductionBalance(db *sql.DB, ns string, amount float64) (*Balance, error) {

	balance, err := GetBalanceByNamespace(db, ns)
	if err != nil {
		return nil, err
	}
	balance.Balance -= amount
	if balance.Balance < 0 {
		return balance, errors.New("need recharge first.")
	}
	return UpdateBalance(db, balance)
}

func checkSqlErr(err error) {
	switch {
	case err == sql.ErrNoRows:
		logger.Error("No such rows:", err)

	case err != nil:
		log.Fatal(err)
	}
}

func logRollback(tx *sql.Tx) {
	err := tx.Rollback()
	if err != sql.ErrTxDone && err != nil {
		logger.Error(err.Error())
	}
}
