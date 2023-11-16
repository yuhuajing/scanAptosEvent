package aptosclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/coming-chat/go-aptos/aptostypes"
)

const (
	VERSIONNONE = ""
	VERSION1    = "v1"
)

type RestClient struct {
	chainId int
	c       *http.Client
	rpcUrl  string
	version string
}

func Dial(ctx context.Context, rpcUrl string) (client *RestClient, err error) {
	client = &RestClient{
		rpcUrl: strings.TrimRight(rpcUrl, "/"),
		c: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    3,
				IdleConnTimeout: 30 * time.Second,
			},
			Timeout: 30 * time.Second,
		},
		version: VERSION1,
	}
	err = client.setChainId()
	return
}

func DialWithClient(ctx context.Context, rpcUrl string, c *http.Client) (client *RestClient, err error) {
	client = &RestClient{
		rpcUrl:  strings.TrimRight(rpcUrl, "/"),
		c:       c,
		version: VERSION1,
	}
	err = client.setChainId()
	return
}

func (c *RestClient) SetVersion(version string) {
	c.version = version
}

func (c *RestClient) GetVersion() string {
	return c.version
}

func (c *RestClient) GetVersionedRpcUrl() string {
	return c.rpcUrl + "/" + c.version
}

func (c *RestClient) LedgerInfo() (res *aptostypes.LedgerInfo, err error) {
	req, err := http.NewRequest("GET", c.GetVersionedRpcUrl(), nil)
	if err != nil {
		return
	}
	res = &aptostypes.LedgerInfo{}
	err = c.doReq(req, res)
	return
}

func (c *RestClient) RawQuery(urlWithoutVersion string, params map[string]string) (data []byte, err error) {
	urlWithoutVersion = "/" + strings.TrimPrefix(urlWithoutVersion, "/")
	req, err := http.NewRequest("GET", c.GetVersionedRpcUrl()+urlWithoutVersion, nil)
	if err != nil {
		return
	}
	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return
	}
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode >= http.StatusBadRequest {
		restError := &aptostypes.RestError{}
		json.Unmarshal(data, &restError)
		restError.Code = resp.StatusCode
		return nil, restError
	}
	return
}

func (c *RestClient) setChainId() (err error) {
	ledger, err := c.LedgerInfo()
	if err != nil {
		return
	}
	c.chainId = ledger.ChainId
	return
}

func (c *RestClient) ChainId() int {
	return c.chainId
}

// doReq send request and unmarshal response body to result
func (c *RestClient) doReq(req *http.Request, result interface{}) error {
	return doReqWithClient(req, result, c.c)
}

func doReqWithClient(req *http.Request, result interface{}, client *http.Client) error {
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return handleResponse(result, resp)
}

// handleResponse read response data and unmarshal to result
// if http status code >= 400, function will return error with raw content
func handleResponse(result interface{}, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		restError := &aptostypes.RestError{}
		json.Unmarshal(body, &restError)
		restError.Code = resp.StatusCode
		return restError
	}
	return json.Unmarshal(body, &result)
}
