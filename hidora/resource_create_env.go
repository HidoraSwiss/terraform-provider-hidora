package hidora

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type VolumeMounts struct {
	Protocol          string `json:"protocol"`
	ReadOnly          bool   `json:"readonly"`
	Sourceaddresstype string `json:"sourceAddressType"`
	Sourcenodeid      int    `json:"sourceNodeId"`
	Sourcenodegroup   string `json:"sourceNodeGroup"`
	Sourcepath        string `json:"sourcePath"`
}

type Envsettings struct {
	Displayname string `json:"displayname"` // Deprecated
	Engine      string `json:"engine"`      // Deprecated
	Ishaenabled bool   `json:"ishaenabled"`
	Region      string `json:"region"`
	Shortdomain string `json:"shortdomain"`
	Sslstate    bool   `json:"sslstate"`
}

type Nodes struct {
	Cmd               string                   `json:"cmd"`
	Count             uint8                    `json:"count"`
	Disklimit         uint8                    `json:"diskLimit"`
	Env               map[string]string        `json:"env"`
	Extip             bool                     `json:"extip"`
	Extipv6           bool                     `json:"extipv6"`
	Fixedcloudlets    uint8                    `json:"fixedCloudlets"`
	Flexiblecloudlets uint8                    `json:"flexibleCloudlets"`
	Image             string                   `json:"image"`
	Mission           string                   `json:"mission"`
	Nodegroup         string                   `json:"nodeGroup"`
	Nodetype          string                   `json:"nodeType"`
	Restartdelay      uint16                   `json:"restartDelay"`
	Scalingmode       string                   `json:"scalingMode"`
	Tag               string                   `json:"tag"`
	Volumemounts      map[string]*VolumeMounts `json:"volumeMounts"` // Don't used
	Volumes           []string                 `json:"volumes"`
	Volumesfrom       []string                 `json:"volumesFrom"` // Don't know :/
}

type Createenvironment struct {
	Actionkey   string
	Appid       string
	Envgroups   string // Just one envgroup
	Environment *Envsettings
	Nodes       []*Nodes
	Owneruid    uint32
	Session     string
}

type Region struct {
	Displayname string
	Uniquename  string
}

const (
	API_ENV_CONTROL_GETREGIONS_ENDPOINT     string = "environment/control/rest/getregions"
	API_ENV_CONTROL_CREATEENV_ENDPOINT      string = "environment/control/rest/createenvironment"
	API_ENV_CONTROL_GETENVINFO_ENDPOINT     string = "environment/control/rest/getenvinfo"
	API_ENV_CONTROL_DELETEENV_ENDPOINT      string = "environment/control/rest/deleteenv"
	API_ENV_CONTROL_CHANGETOPOLOGY_ENDPOINT string = "environment/control/rest/changetopology"
	API_ENV_CONTROL_SETENVGROUP_ENDPOINT    string = "environment/control/rest/setenvgroup"
	API_ENV_CONTROL_MIGRATE_ENDPOINT        string = "environment/control/rest/migrate"
	APPID_LENGTH                            int    = 32
	SHORTDOMAIN_MIN_LENGTH                  int    = 5  // Not be so sure
	SHORTDOMAIN_MAX_LENGTH                  int    = 41 // Not be so sure
)

func resourceJelasticCreateEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceJelasticCreateEnvironmentCreate,
		ReadContext:   resourceJelasticCreateEnvironmentRead,
		UpdateContext: resourceJelasticCreateEnvironmentUpdate,
		DeleteContext: resourceJelasticCreateEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"appid": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     PLATFORM_APPID,
				Description: "Application Identity in Jelastic Platform",
			},
			"environment": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ishaenabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "",
						},
						"region": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"shortdomain": { // Verify policy
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "",
						},
						"sslstate": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "",
						},
						"createdon": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsRFC3339Time,
							Description:  "",
						},
						"appid": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "",
						},
						"domain": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "",
						},
						"hardwarenodegroup": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "",
						},
					},
				},
			},
			"nodes": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The date and time of the creation of the Project (Format ISO 8601)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cmd": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"count": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1,
							Description: "",
						},
						"disklimit": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "",
						},
						"env": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"extip": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "",
						},
						"extipv6": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "",
						},
						"fixedcloudlets": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1,
							Description: "",
						},
						"flexiblecloudlets": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     4,
							Description: "",
						},
						"image": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"mission": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						"nodegroup": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"nodetype": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"restartdelay": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     30,
							Description: "",
						},
						"scalingmode": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "STATEFUL",
							Description: "",
						},
						"tag": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
						},
						// "volumemounts": {
						// 	Type:        schema.TypeMap,
						// 	Optional:    true,
						// 	Description: "",
						// 	Elem: &schema.Resource{
						// 		Schema: map[string]*schema.Schema{},
						// 	},
						// },
						"volumes": { // Suspicious
							Type:        schema.TypeList,
							Optional:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"volumesfrom": { // Suspicious
							Type:        schema.TypeList,
							Optional:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"actionkey": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
			},
			"owneruid": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "UID of the owner of environment",
			},
			"envgroups": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Define which group is chosen for the environment",
			},
		},
	}
}

func resourceJelasticCreateEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Statement of m and type assertion with *CLient
	m := meta.(*Client)
	// Statement of http client to do API requests
	client := m.HTTPClient

	// Declare diag variable for debugging
	var diags diag.Diagnostics

	// Define REST URL
	u := *m.BaseUrl
	u.Path += API_ENV_CONTROL_CREATEENV_ENDPOINT
	urlStr := u.String()

	// Allocation of CreateEnvironment struct
	createenv := new(Createenvironment)
	createenv.Environment = new(Envsettings)

	tf_nodes := d.Get("nodes").([]interface{})
	tf_nodes_len := len(tf_nodes)
	createenv.Nodes = initObjRefsWithPreallocation(tf_nodes_len)

	// Statement of env and nodes to ease
	// the insertion of value
	env := &createenv.Environment
	nodes := &createenv.Nodes

	// Fill Struct Createenvironment
	// Check appid format
	appid := d.Get("appid").(string)
	is_string_alphabetic := regexp.MustCompile(`^[a-z0-9]*$`).MatchString
	appid_isalphanumeric := is_string_alphabetic(appid)
	appid_length := len([]rune(appid))
	if appid_isalphanumeric &&
		appid_length == APPID_LENGTH {
		createenv.Appid = appid
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Incorrect appid format",
			Detail: fmt.Sprintf("appid doesn't correspond to format policy ! Must have only alphanumeric with %d characters",
				APPID_LENGTH),
		})
		return diags
	}

	// Check session format
	createenv.Session = m.Token

	// Check actionkey
	createenv.Actionkey = d.Get("actionkey").(string)

	// Check owneruid
	createenv.Owneruid = uint32(d.Get("owneruid").(int))

	// Check envgroups
	// API method can create a new envgroup if it doesn't exist
	createenv.Envgroups = d.Get("envgroups").(string)

	// Get all values of environment
	tf_env := d.Get("environment").([]interface{})[0]
	tf_env_data := tf_env.(map[string]interface{})

	// Fill Struct Envsettings
	// Check ishaenabled
	ishaenabled, ok := tf_env_data["ishaenabled"].(bool)
	if ok {
		(*env).Ishaenabled = ishaenabled
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong value type",
			Detail:   "Value type has to be a boolean type.",
		})
		return diags
	}

	// Check region
	// Define REST URL
	u_getregions := *m.BaseUrl
	u_getregions.Path += API_ENV_CONTROL_GETREGIONS_ENDPOINT
	urlStr_getregions := u_getregions.String()

	// Define request configuration
	var req_region JelasticRequest = JelasticRequest{
		Method:  http.MethodPost,
		Headers: client_headers,
	}

	req_region.Query = url.Values{
		"appid":   {PLATFORM_APPID},
		"session": {m.Token},
	}
	req_region.Body = strings.NewReader(req_region.Query.Encode())
	req, _ := http.NewRequest(req_region.Method, urlStr_getregions, req_region.Body)
	req.Header = req_region.Headers
	resp_getregions, err := client.Do(req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get regions",
			Detail:   "Wrong appid or session token",
		})
		return diags
	}
	defer resp_getregions.Body.Close()
	body, _ := ioutil.ReadAll(resp_getregions.Body)
	var result_getregions map[string]interface{}
	json.Unmarshal(body, &result_getregions)
	if result_getregions["result"].(float64) != 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get regions",
			Detail:   "Wrong appid or session token",
		})
		return diags
	}
	result_array := result_getregions["array"].([]interface{})
	result_array_len := len(result_array)
	regions := make([]Region, result_array_len)
	is_region_accepted := false
	detail_region_message := ""
	for i := 0; i < result_array_len; i++ {
		result_array_id := result_array[i].(map[string]interface{})
		hardnodes_array := result_array_id["hardNodeGroups"].([]interface{})
		hardnodes_array_len := len(hardnodes_array)
		for j := 0; j < hardnodes_array_len; j++ {
			hardnodes_array_id := hardnodes_array[j].(map[string]interface{})
			if hardnodes_array_id["isEnabled"] == true {
				regions[i].Uniquename = hardnodes_array_id["uniqueName"].(string)
				regions[i].Displayname = hardnodes_array_id["displayName"].(string)
				detail_region_message += fmt.Sprintf("Region: %s, value: %s ",
					regions[i].Displayname,
					regions[i].Uniquename)
				if tf_env_data["region"].(string) == regions[i].Uniquename {
					is_region_accepted = true
					break
				}
			}
		}
		if is_region_accepted {
			break
		}
	}
	region, ok := tf_env_data["region"].(string)
	if ok && is_region_accepted {
		(*env).Region = region
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong region selected",
			Detail: fmt.Sprintf("Selected region unknown, please use one of these regions instead : %s",
				detail_region_message),
		})
		return diags
	}

	// Check shortdomain
	shortdomain, ok := tf_env_data["shortdomain"].(string)
	shortdomain_len := len(shortdomain)
	if ok {
		is_string_alphabetic = regexp.MustCompile(`^[a-zA-Z0-9-]*$`).MatchString
		shortdomain_isalphanumeric := is_string_alphabetic(shortdomain)
		if shortdomain_isalphanumeric &&
			shortdomain_len >= SHORTDOMAIN_MIN_LENGTH &&
			shortdomain_len <= SHORTDOMAIN_MAX_LENGTH {
			(*env).Shortdomain = shortdomain
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "shortdomain value is not alphanumerical.",
				Detail:   "shortdomain can contain special characters but they are prohibited",
			})
			return diags
		}
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong type for shortdomain",
			Detail:   "shortdomain is not a string.",
		})
		return diags
	}

	// Check sslstate
	sslstate, ok := tf_env_data["sslstate"].(bool)
	if ok {
		(*env).Sslstate = sslstate
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong value type",
			Detail:   "Value type has to be a boolean type.",
		})
		return diags
	}

	// Fill Struct Nodes
	// without check fields !
	for i, node := range *nodes {
		tf_node := tf_nodes[i].(map[string]interface{})
		node.Cmd = tf_node["cmd"].(string)
		if ishaenabled && tf_node["count"].(int) == 1 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Not enough computed nodes for replication",
				Detail:   "You activated the haenabled and count have to be greater than 1",
			})
			return diags
		}
		node.Count = uint8(tf_node["count"].(int))
		node.Disklimit = uint8(tf_node["disklimit"].(int))
		var node_env_map = make(map[string]string)
		for k, v := range tf_node["env"].(map[string]interface{}) {
			value, _ := v.(string)
			node_env_map[k] = value
		}
		node.Env = node_env_map
		node.Extip = tf_node["extip"].(bool)
		node.Extipv6 = tf_node["extipv6"].(bool)
		node.Fixedcloudlets = uint8(tf_node["fixedcloudlets"].(int))
		node.Flexiblecloudlets = uint8(tf_node["flexiblecloudlets"].(int))
		node.Image = tf_node["image"].(string)
		node.Mission = tf_node["mission"].(string)
		node.Nodegroup = tf_node["nodegroup"].(string)
		node.Nodetype = tf_node["nodetype"].(string)
		node.Restartdelay = uint16(tf_node["restartdelay"].(int))
		node.Scalingmode = tf_node["scalingmode"].(string)
		node.Tag = tf_node["tag"].(string)
		//node.Volumemounts = tf_node["volumemounts"].(map[string]*VolumeMounts)
		node_volumes_len := len(tf_node["volumes"].([]interface{}))
		var node_volumes = make([]string, node_volumes_len)
		for j, v := range tf_node["volumes"].([]interface{}) {
			node_volume, _ := v.(string)
			node_volumes[j] = node_volume
		}
		node.Volumes = node_volumes
		node_volumesfrom_len := len(tf_node["volumesfrom"].([]interface{}))
		var node_volumesfrom = make([]string, node_volumesfrom_len)
		for j, v := range tf_node["volumes"].([]interface{}) {
			node_volumefrom, _ := v.(string)
			node_volumesfrom[j] = node_volumefrom
		}
		node.Volumesfrom = node_volumesfrom
	}

	// Deserialize Createenvironment & Nodes
	env_json, _ := json.Marshal(*env)
	nodes_json, _ := json.Marshal(*nodes)

	// Convert to string each json element
	env_json_string := string(env_json)
	nodes_json_string := string(nodes_json)

	// Probe API Server with parameters
	var req_config JelasticRequest = JelasticRequest{
		Method:  http.MethodPost,
		Headers: client_headers,
	}

	req_config.Query = url.Values{
		"appid":     {PLATFORM_APPID},
		"session":   {m.Token},
		"env":       {env_json_string},                       // JSON env
		"nodes":     {nodes_json_string},                     // JSON nodes
		"actionkey": {createenv.Actionkey},                   // Optional
		"owneruid":  {strconv.Itoa(int(createenv.Owneruid))}, // Optional
		"envgroups": {createenv.Envgroups},                   // Optional
	}
	if createenv.Actionkey == "" {
		req_config.Query.Del("actionkey")
	}
	if createenv.Owneruid == 0 {
		req_config.Query.Del("owneruid")
	}
	if createenv.Envgroups == "" {
		req_config.Query.Del("envgroups")
	}

	req_config.Body = strings.NewReader(req_config.Query.Encode())
	req, _ = http.NewRequest(req_config.Method, urlStr, req_config.Body)
	req.Header = req_config.Headers

	// Take Response
	resp, err := client.Do(req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create environment",
			Detail:   "Can't create environment because of timeout or bad values in paramaters",
		})
		return diags
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	result_response := result["response"].(map[string]interface{})
	result_number := result_response["result"].(float64)
	if result_number == 2314 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API request failed",
			Detail:   fmt.Sprint(result_response["error"]),
		})
		return diags
	} else if result_number != 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API request failed",
			Detail:   "Incorrect value in fields", // Not so much explicit response :/
		})
		return diags
	}
	d.SetId(result_response["name"].(string)) // Because API only search by shortdomain of environment

	return resourceJelasticCreateEnvironmentRead(ctx, d, meta)
}

func resourceJelasticCreateEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Statement of m and type assertion with *CLient
	m := meta.(*Client)
	// Statement of http client to do API requests
	client := m.HTTPClient
	session := m.Token

	// Declare diag variable for debugging
	var diags diag.Diagnostics

	// Define REST URL
	u := *m.BaseUrl
	u.Path += API_ENV_CONTROL_GETENVINFO_ENDPOINT
	urlStr := u.String()

	var req_config JelasticRequest = JelasticRequest{
		Method:  http.MethodPost,
		Headers: client_headers,
	}

	req_config.Query = url.Values{
		"envName": {d.Id()},
		"session": {session},
		"lazy":    {"true"}, // To have less informations, can be changed
	}
	req_config.Body = strings.NewReader(req_config.Query.Encode())
	req, _ := http.NewRequest(req_config.Method, urlStr, req_config.Body)
	req.Header = req_config.Headers

	resp, err := client.Do(req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to get environment informations",
			Detail:   "Can't get informations about environment because of bad envName",
		})
		return diags
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if result["result"].(float64) != 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API request to getenvinfo failed",
			Detail:   fmt.Sprintf("Cannot get environment informations from %s", d.Id()),
		})
		return diags
	}
	result_response := result["env"].(map[string]interface{})
	_ = d.Set("environment", flattenCreateEnvironmentEnvironmentData(result_response))
	_ = d.Set("owneruid", result_response["uid"].(float64))

	return nil
}

// Update only envgroups and environment fields, not nodes
func resourceJelasticCreateEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Statement of m and type assertion with *CLient
	m := meta.(*Client)
	// Statement of http client to do API requests
	client := m.HTTPClient
	session := m.Token

	// Declare diag variable for debugging
	var diags diag.Diagnostics

	// Define REST URL
	u := *m.BaseUrl

	var req_config JelasticRequest = JelasticRequest{
		Method:  http.MethodPost,
		Headers: client_headers,
	}

	// envgroups -> setenvgroups API method
	// ishaenabled -> ChangeTopology API method
	// region -> migrate API method (don"t forget to check hardwarenodegroup)
	// shortdomain -> None recreate resource
	// sslstate -> ChangeTopology API method

	if d.HasChange("envgroups") {
		u_envgroups := u
		u_envgroups.Path += API_ENV_CONTROL_SETENVGROUP_ENDPOINT
		urlStr := u_envgroups.String()

		req_config.Query = url.Values{
			"envName":  {d.Id()},
			"session":  {session},
			"envGroup": {d.Get("envgroups").(string)},
		}

		req_config.Body = strings.NewReader(req_config.Query.Encode())
		req, _ := http.NewRequest(req_config.Method, urlStr, req_config.Body)
		req.Header = req_config.Headers

		resp, err := client.Do(req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set an environment groups",
				Detail:   "",
			})
			return diags
		}
		defer resp.Body.Close()
	}
	if d.HasChange("environment.0.region") {
		u_region := u
		u_region.Path += API_ENV_CONTROL_MIGRATE_ENDPOINT
		urlStr := u_region.String()

		req_config.Query = url.Values{
			"envName":           {d.Id()},
			"session":           {session},
			"hardwareNodeGroup": {d.Get("environment.0.region").(string)}, // No check /!\
			"isOnline":          {"true"},                                 // arbitrary, can be modified
		}

		req_config.Body = strings.NewReader(req_config.Query.Encode())
		req, _ := http.NewRequest(req_config.Method, urlStr, req_config.Body)
		req.Header = req_config.Headers

		resp, err := client.Do(req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf("Unable to migrate environment to %s",
					d.Get("environment.0.region").(string)),
				Detail: "Cannot find hardwareNodeGroup",
			})
			return diags
		}
		defer resp.Body.Close()
	}

	// Reapeat the same checks as resourceJelasticCreateEnvironmentCreate
	// Not implemented
	if (d.HasChange("environment.0.ishaenabled") &&
		d.HasChange("environment.0.sslstate")) ||
		(d.HasChange("environment.0.ishaenabled") ||
			d.HasChange("environment.0.sslstate")) {
	}

	return resourceJelasticCreateEnvironmentRead(ctx, d, meta)
}

func resourceJelasticCreateEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Statement of m and type assertion with *CLient
	m := meta.(*Client)
	// Statement of http client to do API requests
	client := m.HTTPClient
	session := m.Token

	// Declare diag variable for debugging
	var diags diag.Diagnostics

	// Define REST URL
	u := *m.BaseUrl
	u.Path += API_ENV_CONTROL_DELETEENV_ENDPOINT
	urlStr := u.String()

	var req_config JelasticRequest = JelasticRequest{
		Method:  http.MethodPost,
		Headers: client_headers,
	}

	req_config.Query = url.Values{
		"envName": {d.Id()},
		"session": {session},
	}
	req_config.Body = strings.NewReader(req_config.Query.Encode())
	req, _ := http.NewRequest(req_config.Method, urlStr, req_config.Body)
	req.Header = req_config.Headers

	resp, err := client.Do(req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to delete environment %s", d.Id()),
			Detail:   fmt.Sprintf("The environment %s doesn't exist", d.Id()),
		})
		return diags
	}
	defer resp.Body.Close()

	return nil
}

func initObjRefsWithPreallocation(n int) []*Nodes {
	objs := make([]Nodes, n)
	refs := make([]*Nodes, 0, n)
	for i := 0; i < n; i++ {
		//objs[i].id = i
		refs = append(refs, &objs[i])
	}
	return refs
}

func flattenCreateEnvironmentEnvironmentData(env map[string]interface{}) interface{} {
	if env == nil {
		return nil
	}
	result_response_hostgroup := env["hostGroup"].(map[string]interface{})
	flatten_environment := []map[string]interface{}(nil)
	flatten_environment = append(flatten_environment, map[string]interface{}{
		"appid":             env["appid"].(string),
		"createdon":         env["createdOn"].(string),
		"domain":            env["domain"].(string),
		"hardwarenodegroup": env["hardwareNodeGroup"].(string),
		"ishaenabled":       env["ishaenabled"].(bool),
		"region":            result_response_hostgroup["uniqueName"].(string),
		"shortdomain":       env["shortdomain"].(string),
		"sslstate":          env["sslstate"].(bool),
	})
	return flatten_environment
}
