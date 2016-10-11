package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asiainfoLDP/datafoundry_recharge/common"
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

const DataFoundryHost = "https://dev.dataos.io:8443"

func authDF(token string) (*User, error) {
	url := fmt.Sprintf("%s/oapi/v1/users/~", DataFoundryHost)

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

func getDFUserame(token string) (string, error) {
	//Logger.Info("token = ", token)

	user, err := authDF(token)
	if err != nil {
		return "", err
	}
	return dfUser(user), nil
}
