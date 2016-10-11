package models

import (
	"time"
)

type Recharge struct {
	RechargeId int64     `json:"rechargeid"`
	Amount     float64   `json:"amount"`
	Namespace  string    `json:"namespace"`
	User       string    `json:"user,omitempty"`
	CreateTime time.Time `json:"createtime,omitempty"`
	Status     string    `json:"status,omitempty"`
	StatusTime time.Time `json:"statustime,omitempty"`
}

/*func CreatePlan(db *sql.DB, planInfo *Plan) (string, error) {
	logger.Info("Model begin create a plan.")
	defer logger.Info("Model end create a plan.")

	nowstr := time.Now().Format("2006-01-02 15:04:05.999999")
	sqlstr := fmt.Sprintf(`insert into DF_PLAN (
				PLAN_ID, PLAN_NAME, PLAN_TYPE, SPECIFICATION1, SPECIFICATION2,
				PRICE, CYCLE, REGION, CREATE_TIME, STATUS
				) values (
				?, ?, ?, ?, ?, ?, ?, ?,
				'%s', '%s')`,
		nowstr, "A")

	_, err := db.Exec(sqlstr,
		planInfo.Plan_id, planInfo.Plan_name, planInfo.Plan_type, planInfo.Specification1, planInfo.Specification2,
		planInfo.Price, planInfo.Cycle, planInfo.Region)

	return planInfo.Plan_id, err
}

func DeletePlan(db *sql.DB, planId string) error {
	logger.Info("Model begin delete a plan.")
	defer logger.Info("Model begin delete a plan.")

	//sqlstr := fmt.Sprintf(`update DF_PLAN set status = "N" where PLAN_ID = '%s'`, planId)
	//_, err := db.Exec(sqlstr)

	err := modifyPlanStatusToN(db, planId)
	if err != nil {
		return err
	}

	return err
}

func ModifyPlan(db *sql.DB, planInfo *Plan) error {
	logger.Info("Model begin modify a plan.")
	defer logger.Info("Model begin modify a plan.")

	plan, err := RetrievePlanByID(db, planInfo.Plan_id)
	if err != nil {
		return err
	}
	logger.Debug("Retrieve plan: %v", plan)

	err = modifyPlanStatusToN(db, plan.Plan_id)
	if err != nil {
		return err
	}

	//planInfo.Plan_id = 0
	_, err = CreatePlan(db, planInfo)
	if err != nil {
		return err
	}

	return err
}

func RetrievePlanByID(db *sql.DB, planID string) (*Plan, error) {
	logger.Info("Model begin get a plan by id.")
	defer logger.Info("Model end get a plan by id.")

	return getSinglePlan(db, fmt.Sprintf("PLAN_ID = '%s' and STATUS = 'A'", planID))
}

func getSinglePlan(db *sql.DB, sqlWhere string) (*Plan, error) {
	apps, err := queryPlans(db, sqlWhere, 1, 0)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	if len(apps) == 0 {
		return nil, nil
	}

	return apps[0], nil
}

func queryPlans(db *sql.DB, sqlWhere string, limit int, offset int64, sqlParams ...interface{}) ([]*Plan, error) {
	offset_str := ""
	if offset > 0 {
		offset_str = fmt.Sprintf("offset %d", offset)
	}

	sqlWhereAll := ""
	if sqlWhere != "" {
		sqlWhereAll = fmt.Sprintf("where %s", sqlWhere)
	}

	sql_str := fmt.Sprintf(`select
					PLAN_ID, PLAN_NAME, PLAN_TYPE,
					SPECIFICATION1,
					SPECIFICATION2,
					PRICE, CYCLE, REGION,
					CREATE_TIME, STATUS
					from DF_PLAN
					%s
					limit %d
					%s
					`,
		sqlWhereAll,
		limit,
		offset_str)
	rows, err := db.Query(sql_str, sqlParams...)

	logger.Info(">>> %v", sql_str)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := make([]*Plan, 0, 100)
	for rows.Next() {
		plan := &Plan{}
		err := rows.Scan(
			&plan.Plan_id, &plan.Plan_name, &plan.Plan_type, &plan.Specification1, &plan.Specification2,
			&plan.Price, &plan.Cycle, &plan.Region, &plan.Create_time, &plan.Status,
		)
		if err != nil {
			return nil, err
		}
		//validateApp(s) // already done in scanAppWithRows
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func modifyPlanStatusToN(db *sql.DB, planId string) error {
	sqlstr := fmt.Sprintf(`update DF_PLAN set status = "N" where PLAN_ID = '%s' and STATUS = 'A'`, planId)

	_, err := db.Exec(sqlstr)
	if err != nil {
		return err
	}

	return err
}

func QueryPlans(db *sql.DB, orderBy string, sortOrder bool, offset int64, limit int) (int64, []*Plan, error) {
	logger.Info("Model begin get plan list.")
	defer logger.Info("Model end get plan list.")

	sqlParams := make([]interface{}, 0, 4)

	// ...

	sqlWhere := "STATUS = 'A'"
	//provider = strings.ToLower(provider)
	//if provider != "" {
	//	if sqlWhere == "" {
	//		sqlWhere = "PROVIDER=?"
	//	} else {
	//		sqlWhere = sqlWhere + " and PROVIDER=?"
	//	}
	//	sqlParams = append(sqlParams, provider)
	//}
	//if category != "" {
	//	if sqlWhere == "" {
	//		sqlWhere = "CATEGORY=?"
	//	} else {
	//		sqlWhere = sqlWhere + " and CATEGORY=?"
	//	}
	//	sqlParams = append(sqlParams, category)
	//}

	// ...

	switch strings.ToLower(orderBy) {
	default:
		orderBy = "CREATE_TIME"
		sortOrder = false
	case "createtime":
		orderBy = "CREATE_TIME"
	case "hotness":
		orderBy = "HOTNESS"
	}

	sqlSort := fmt.Sprintf("%s %s", orderBy, sortOrderText[sortOrder])

	// ...

	return getPlanList(db, offset, limit, sqlWhere, sqlSort, sqlParams...)
}

const (
	SortOrder_Asc  = "asc"
	SortOrder_Desc = "desc"
)

// true: asc
// false: desc
var sortOrderText = map[bool]string{true: "asc", false: "desc"}

func ValidateSortOrder(sortOrder string, defaultOrder bool) bool {
	switch strings.ToLower(sortOrder) {
	case SortOrder_Asc:
		return true
	case SortOrder_Desc:
		return false
	}

	return defaultOrder
}

func ValidateOrderBy(orderBy string) string {
	switch orderBy {
	case "createtime":
		return "CREATE_TIME"
	case "hotness":
		return "HOTNESS"
	}

	return ""
}

func getPlanList(db *sql.DB, offset int64, limit int, sqlWhere string, sqlSort string, sqlParams ...interface{}) (int64, []*Plan, error) {
	//if strings.TrimSpace(sqlWhere) == "" {
	//	return 0, nil, errors.New("sqlWhere can't be blank")
	//}

	count, err := queryPlansCount(db, sqlWhere)
	if err != nil {
		return 0, nil, err
	}
	if count == 0 {
		return 0, []*Plan{}, nil
	}
	validateOffsetAndLimit(count, &offset, &limit)

	subs, err := queryPlans(db,
		fmt.Sprintf(`%s order by %s`, sqlWhere, sqlSort),
		limit, offset, sqlParams...)

	return count, subs, err
}

func queryPlansCount(db *sql.DB, sqlWhere string, sqlParams ...interface{}) (int64, error) {
	sqlWhere = strings.TrimSpace(sqlWhere)
	sql_where_all := ""
	if sqlWhere != "" {
		sql_where_all = fmt.Sprintf("where %s", sqlWhere)
	}

	count := int64(0)
	sql_str := fmt.Sprintf(`select COUNT(*) from DF_PLAN %s`, sql_where_all)
	logger.Debug(">>>\n"+
		"	%s", sql_str)
	err := db.QueryRow(sql_str, sqlParams...).Scan(&count)

	return count, err
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
*/
