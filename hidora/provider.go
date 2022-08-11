package hidora

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Client struct {
	BaseUrl    *url.URL
	HTTPClient *http.Client
	Token      string
}

type JelasticRequest struct {
	Method  string
	Headers http.Header
	Query   url.Values
	Body    io.Reader
}

const (
	PLATFORM_APPID                 string = "1dd8d191d38fff45e62564fcf67fdcd6" // https://docs.jelastic.com/api
	API_PROTO                      string = "https://"
	API_VERSION                    string = "/1.0/"
	TOKEN_LENGTH                   int    = 40
	API_USERS_AUTH_SIGNIN_ENDPOINT string = "users/authentication/rest/signin"
)

var client_headers = http.Header{
	"Content-type":   {"application/x-www-form-urlencoded"},
	"Accept-Charset": {"UTF-8"},
	"Accept":         {"application/json"},
}

// Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:     schema.TypeString,
				Optional: true,
				//DefaultFunc: schema.EnvDefaultFunc("JELASTIC_HOST", nil),
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("JELASTIC_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("JELASTIC_PASSWORD", nil),
			},
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("JELASTIC_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"jelastic_create_env": resourceJelasticCreateEnvironment(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"jelastic_create_env": dataSourceJelasticCreateEnvironment(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var client *http.Client = &http.Client{
		Timeout: 3600 * time.Second, // Extreme long timeout
	}

	var c Client = Client{
		BaseUrl:    &url.URL{},
		HTTPClient: client,
		Token:      "",
	}

	var req_config JelasticRequest = JelasticRequest{
		Method:  http.MethodPost,
		Headers: client_headers,
	}

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	access_token := d.Get("access_token").(string)

	var host *string

	hVal, ok := d.GetOk("host")
	if ok {
		tempHost := hVal.(string)
		host = &tempHost
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	u, _ := url.ParseRequestURI(API_PROTO + *host + API_VERSION)
	c.BaseUrl = u

	if (username != "") && (password != "") {
		u.Path = API_USERS_AUTH_SIGNIN_ENDPOINT
		urlStr := u.String()

		req_config.Query = url.Values{
			"appid":    {PLATFORM_APPID},
			"login":    {username},
			"password": {password},
		}

		req_config.Body = strings.NewReader(req_config.Query.Encode())
		req, _ := http.NewRequest(req_config.Method, urlStr, req_config.Body)
		req.Header = req_config.Headers
		resp, err := client.Do(req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Access token",
				Detail:   "Unable to authenticate user for authenticated Hidora client",
			})
			return nil, diags
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		c.Token = result["session"].(string)
		return &c, diags
	}
	if access_token != "" {
		is_string_alphabetic := regexp.MustCompile(`^[a-z0-9]*$`).MatchString
		token_isalphanumeric := is_string_alphabetic(access_token)
		token_length := len([]rune(access_token))
		if token_isalphanumeric && (token_length <= TOKEN_LENGTH) {
			c.Token = access_token
			return &c, diags
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Incorrect token format",
				Detail:   fmt.Sprintf("Token doesn't correspond to format policy ! Must have only alphanumeric with %d characters", TOKEN_LENGTH),
			})
			return nil, diags
		}
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create or to get Access token",
		Detail:   "Field username or password are empty and access_token is empty !",
	})
	return nil, diags
}
