package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type CredentialIface interface {
	GetSecretId() string
	GetToken() string
	GetSecretKey() string
	Refresh() error
	GetRole() string
}

type Credential struct {
	rwLocker    sync.RWMutex
	Transport   http.RoundTripper
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
			err := c.refresh()
			if err != nil {
				fmt.Println("refresh credential error: ", err.Error())
			}
		}
	}
}

func (c *Credential) refresh() error {
	c.rwLocker.RLock()
	expiredTime := c.ExpiredTime
	c.rwLocker.RUnlock()
	if time.Now().Unix() < expiredTime-720 {
		return nil
	}
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

	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()
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

func NewCredentialTransport(role string) *Credential {
	return &Credential{
		Role: role,
	}
}

func (c *Credential) GetSecretKey() string {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	return c.SecretKey
}

func (c *Credential) GetSecretId() string {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	return c.SecretId
}

func (c *Credential) GetToken() string {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()
	return c.Token
}

func (c *Credential) GetRole() string {
	return c.Role
}

func (c *Credential) RoundTrip(req *http.Request) (*http.Response, error) {
	err := c.refresh()
	if err != nil {
		return nil, err
	}
	req = cloneRequest(req)
	// 增加 Authorization header
	authTime := cos.NewAuthTime(time.Hour)
	cos.AddAuthorizationHeader(c.GetSecretId(), c.GetSecretKey(), c.GetToken(), req, authTime)

	resp, err := c.transport().RoundTrip(req)
	return resp, err
}

func (c *Credential) transport() http.RoundTripper {
	if c.Transport != nil {
		return c.Transport
	}
	return http.DefaultTransport
}

// cloneRequest returns a clone of the provided *http.Request. The clone is a
// shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
