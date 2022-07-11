package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type CredentialIface interface {
	GetSecretId() string
	GetToken() string
	GetSecretKey() string
	Refresh() error
	GetRole() string
}

type Credential struct {
	sync.Mutex
	SecretId    string
	SecretKey   string
	Token       string
	ExpiredTime int64
	Role        string
}

type metadataResponse struct {
	TmpSecretId  string
	TmpSecretKey string
	Token        string
	ExpiredTime  int64
	Code         string
}

func (c *Credential) Refresh() error {
	tick := time.NewTicker(10 * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if time.Now().Unix() > c.ExpiredTime-720 {
				if err := c.refresh(); err != nil {
					return err
				}
			}
		}
	}
}

func (c *Credential) refresh() error {
	res, err := http.Get(fmt.Sprintf("http://metadata.tencentyun.com/meta-data/cam/service-role-security-credentials/%s", c.Role))
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		_ = res.Body.Close()
		res, err = http.Get(fmt.Sprintf("http://metadata.tencentyun.com/meta-data/cam/security-credentials/%s", c.Role))
		if err != nil {
			return err
		}
		if res.StatusCode != 200 {
			return fmt.Errorf("status code is %d", res.StatusCode)
		}
	}

	defer func() { _ = res.Body.Close() }()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	metaData := &metadataResponse{}
	if err := json.Unmarshal(data, metaData); err != nil {
		return err
	}

	if metaData.Code != "Success" {
		return fmt.Errorf("get Code is %s", metaData.Code)
	}

	c.SecretId = metaData.TmpSecretId
	c.SecretKey = metaData.TmpSecretKey
	c.Token = metaData.Token
	c.ExpiredTime = metaData.ExpiredTime
	return nil
}

func NewCredential(role string) (*Credential, error) {
	c := &Credential{
		Role: role,
	}
	err := c.refresh()
	return c, err
}

func NewTokenCredential(secretId, secretKey, token string) *Credential {
	return &Credential{
		SecretId:  secretId,
		SecretKey: secretKey,
		Token:     token,
	}
}

func (c *Credential) GetSecretKey() string {
	return c.SecretKey
}

func (c *Credential) GetSecretId() string {
	return c.SecretId
}

func (c *Credential) GetToken() string {
	return c.Token
}

func (c *Credential) GetRole() string {
	return c.Role
}
