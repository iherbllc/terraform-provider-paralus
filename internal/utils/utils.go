// General program utilities
package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/levigross/grequests"
	"github.com/paralus/cli/pkg/config"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/valyala/fasthttp"
)

func MultiEnvSearch(ks []string) string {
	for _, k := range ks {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// Convert a json to a string map interface
func JsonToMap(jsonStr string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Convert a map to json string
func MapToJsonString(jsonMap map[string]string) (string, error) {
	jsonBytes, err := json.Marshal(&jsonMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// AssertStringNotEmpty asserts when the string is not empty
func AssertStringNotEmpty(message, str string) diag.Diagnostics {
	var diags diag.Diagnostics
	str = strings.TrimSpace(str)
	if str != "" {
		return diags
	}

	if message != "" {
		return diag.FromErr(fmt.Errorf("%s: expected not empty string", message))
	} else {
		return diag.FromErr(fmt.Errorf("expected not empty string"))
	}
}

// Makes the desired REST call
func makeRestCall(uri string, method string, payload interface{}) (string, error) {
	auth := config.GetConfig().GetAppAuthProfile()
	s := getSession(auth.SkipServerCertValid)
	sub := auth.SubProfile()
	headers, err := sub.Auth(s)
	if err != nil {
		return "", err
	}
	headers["Content-Type"] = "application/json"

	// Get URI from a pool
	url := fasthttp.AcquireURI()
	url.Parse(nil, []byte(auth.URL+uri))

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetURI(url)          // copy url into request
	fasthttp.ReleaseURI(url) // now you may release the URI

	req.Header.SetMethod(method)

	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		req.SetBodyRaw(body)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	err = client.Do(req, resp)
	if err != nil {
		return "", fmt.Errorf("connection error: %v", err)
	}
	statusCode := resp.StatusCode()
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("invalid HTTP response code: %d", statusCode)
	}

	respBody := resp.Body()
	f := &commonv3.HttpBody{}
	err = json.Unmarshal([]byte(respBody), f)
	if err != nil {
		return "", err
	}

	if string(f.Data) == "" {
		return string(respBody), nil
	}
	return string(f.Data), nil

}

func getSession(skipServerCertCheck bool) *grequests.Session {
	var sessionRequestOption *grequests.RequestOptions
	if skipServerCertCheck {
		sessionRequestOption = &grequests.RequestOptions{
			InsecureSkipVerify: true,
		}
	}
	return grequests.NewSession(sessionRequestOption)
}
