// General program utilities
package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/levigross/grequests"
	"github.com/paralus/cli/pkg/authprofile"
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
		diags.AddError(fmt.Sprintf("%s: expected not empty string", message), "")
	} else {
		diags.AddError("expected not empty string", "")
	}
	return diags
}

// error types
var (
	ErrResourceNotExists   = errors.New("resource does not exist")
	ErrOperationNotAllowed = errors.New("operation not allowed")
	ErrInvalidCredentials  = errors.New("invalid credentials")
)

// Makes the desired REST call
func makeRestCall(ctx context.Context, uri string, method string, payload interface{}, auth *authprofile.Profile) (resp string, err error) {

	resp = ""

	defer func() {
		err = handleRestPanic(uri, method, payload, err)
	}()

	if auth == nil {
		auth = config.GetConfig().GetAppAuthProfile()
	}

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
	req.SetURI(url)          // copy url into request
	fasthttp.ReleaseURI(url) // now you may release the URI

	req.Header.SetMethod(method)

	if payload != nil {
		body, err := json.MarshalIndent(payload, "", "\t")
		if err != nil {
			return "", err
		}
		tflog.Debug(ctx, fmt.Sprintf("payload body: %s", body))
		req.SetBodyRaw(body)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	fastResp := fasthttp.AcquireResponse()
	err = client.Do(req, fastResp)
	if err != nil {
		return "", fmt.Errorf("connection error: %v", err)
	}
	fasthttp.ReleaseRequest(req)

	statusCode := fastResp.StatusCode()
	respBody := fastResp.Body()
	fasthttp.ReleaseResponse(fastResp)
	if statusCode != http.StatusOK {
		// check if error type is permission issue
		if strings.Contains(string(respBody), "no or invalid credentials") {
			return "", ErrInvalidCredentials
		}
		// check if error type is resource not found
		if strings.Contains(string(respBody), "no rows in result set") {
			return "", ErrResourceNotExists
		}
		// check if error type is permission issue
		if strings.Contains(string(respBody), "method or route not allowed") {
			return "", ErrOperationNotAllowed
		}
		// check if error type is permission issue
		if strings.Contains(string(respBody), "You do not have enough privileges") {
			return "", ErrOperationNotAllowed
		}

		if string(respBody) == "" {
			return "", fmt.Errorf("invalid HTTP response code: %d", statusCode)
		}
		return "", errors.New(string(respBody))
	}

	if len(respBody) <= 0 {
		return "", nil
	}

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

// recover from panic while making rest call
func handleRestPanic(uri string, method string, payload interface{}, err error) error {

	if a := recover(); a != nil {
		if payload != nil {
			body, err := json.MarshalIndent(payload, "", "\t")
			if err != nil {
				return fmt.Errorf("error converting payload %v: %v",
					payload, err)
			}
			return fmt.Errorf("panic while making %s rest call against %s with payload %v: %v",
				uri, method, body, a)
		} else {
			return fmt.Errorf("panic while making %s rest call against %s: %v",
				uri, method, a)
		}
	}
	return err
}

// Converts string pointer to string if not nul, other returns empty string
func DerefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}
