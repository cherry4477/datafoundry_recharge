package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asiainfoLDP/datafoundry_recharge/common"
	"os"
)

type ObjectMeta struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

type User struct {
	ObjectMeta `json:"metadata,omitempty"`

	// FullName is the full name of user
	FullName string `json:"fullName,omitempty"`

	// Identities are the identities associated with this user
	Identities []string `json:"identities"`

	// Groups are the groups that this user is a member of
	Groups []string `json:"groups"`
}

const (
	RegionOne = "cn-north-1"
	RegionTwo = "cn-north-2"
)

var (
	DataFoundryHost  = os.Getenv("DataFoundryRegionOneHost")
	DataFoundryHost2 = os.Getenv("DataFoundryRegionTwoHost")
)

func getHost(region string) string {
	var host string
	if region == "" || region == RegionOne {
		if DataFoundryHost == "" {
			DataFoundryHost = "https://dev.dataos.io:8443"
		}
		host = DataFoundryHost
	} else if region == RegionTwo {
		if DataFoundryHost2 == "" {
			DataFoundryHost2 = "https://lab.asiainfodata.com:8443"
		}
		host = DataFoundryHost2
	} else {
		return ""
	}
	return host
}
func authDF(token, region string) (*User, error) {
	host := getHost(region)
	if host == "" {
		return nil, fmt.Errorf("Invalid region request :%s", region)
	}
	url := fmt.Sprintf("%s/oapi/v1/users/~", host)

	response, data, err := common.RemoteCall("GET", url, token, "")
	if err != nil {
		logger.Error("authDF error: ", err.Error())
		return nil, err
	}

	// todo: use return code and msg instead
	if response.StatusCode != http.StatusOK {
		logger.Error("remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return nil, fmt.Errorf("remote (%s) status code: %d.", url, response.StatusCode)
	}

	user := new(User)
	err = json.Unmarshal(data, user)
	if err != nil {
		logger.Error("authDF Unmarshal error: %s. Data: %s\n", err.Error(), string(data))
		return nil, err
	}

	return user, nil
}

func dfUser(user *User) string {
	return user.Name
}

func getDFUserame(token, region string) (string, error) {
	//Logger.Info("token = ", token)

	user, err := authDF(token, region)
	if err != nil {
		return "", err
	}
	return dfUser(user), nil
}

func checkNameSpacePermission(ns, token, region string) error {
	host := getHost(region)
	if host == "" {
		return fmt.Errorf("Invalid region request :%s, namespace :%s\n", region, ns)
	}
	url := fmt.Sprintf("%s/oapi/v1/projects/%s", host, ns)

	response, data, err := common.RemoteCall("GET", url, token, "")
	if err != nil {
		logger.Error("get projects error: ", err.Error())
		return err
	}

	if response.StatusCode != http.StatusOK {
		logger.Error("remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return fmt.Errorf("remote (%s) status code: %d.", url, response.StatusCode)
	}

	return err
}
