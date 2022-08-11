package hidora

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHidoraCreateEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceJelasticCreateEnvironmentRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Application Identity in Jelastic Platform",
			},
			"environment": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ishaenabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "",
						},
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"shortdomain": { // Verify policy
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"sslstate": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "",
						},
						"createdon": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"appid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"hardwarenodegroup": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
					},
				},
			},
			"nodes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cmd": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"disklimit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "",
						},
						"env": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"extip": { // Goal is to import the public IP
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "",
						},
						"extipv6": { // Goal is to import the public IPv6
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "",
						},
						"fixedcloudlets": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "",
						},
						"flexiblecloudlets": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "",
						},
						"image": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"mission": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"nodegroup": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"nodetype": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"restartdelay": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "",
						},
						"scalingmode": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"tag": {
							Type:        schema.TypeString,
							Computed:    true,
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
							Computed:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"volumesfrom": { // Suspicious
							Type:        schema.TypeList,
							Computed:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"owneruid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "UID of the owner of environment",
			},
			"envgroups": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Define which group is chosen for the environment",
			},
		},
	}
}

func dataSourceJelasticCreateEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		"envName": {d.Get("id").(string)},
		"session": {session},
		"lazy":    {"false"}, // Need all informations
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
			Summary:  "Unable to get environment informations",
			Detail:   "Can't get informations about environment because of bad envName",
		})
		return diags
	}
	result_env := result["env"].(map[string]interface{})
	result_envgroups := result["envGroups"].([]interface{})
	result_nodes := result["nodes"].([]interface{})
	result_nodegroups := result["nodeGroups"].([]interface{})
	_ = d.Set("environment", flattenCreateEnvironmentEnvironmentData(result_env))
	// resnodes := flattenCreateEnvironmentNodesData(&result_nodes, &result_nodegroups)
	// diags = append(diags, diag.Diagnostic{
	// 	Severity: diag.Error,
	// 	Summary:  "Unable to get environment informations",
	// 	Detail:   fmt.Sprintf("%v", resnodes),
	// })
	// return diags
	_ = d.Set("nodes", flattenCreateEnvironmentNodesData(&result_nodes, &result_nodegroups))
	_ = d.Set("owneruid", result_env["ownerUid"].(float64))
	_ = d.Set("envgroups", result_envgroups[0].(string)) // envgroups is not a array, can fix later

	d.SetId(d.Get("id").(string))

	return diags
}

func flattenCreateEnvironmentNodesData(nodes *[]interface{}, nodegroups *[]interface{}) interface{} {
	if nodes == nil {
		return nil
	}
	flatten_nodes := []map[string]interface{}(nil)

	for _, node := range *nodes {
		node_map := node.(map[string]interface{})
		flatten_node := make(map[string]interface{})
		customitem := node_map["customitem"].(map[string]interface{})
		dockermanifest := customitem["dockerManifest"].(map[string]interface{})
		flatten_node["cmd"] = dockermanifest["cmd"].([]interface{})[0].(string) // search value in customitem -> dockerManifest -> cmd
		flatten_node["disklimit"] = int(node_map["diskLimit"].(float64)) / 1000
		node_envs_len := len(dockermanifest["env"].([]interface{}))
		node_envs := make([]string, node_envs_len)
		for j, v := range dockermanifest["env"].([]interface{}) {
			node_env, _ := v.(string)
			node_envs[j] = node_env
		}
		envsmap := make(map[string]string)
		for _, env := range node_envs {
			envcut := strings.SplitN(env, "=", 2)
			envsmap[envcut[0]] = envcut[1]
		}
		flatten_node["env"] = envsmap // search value in customitem -> dockerManifest -> env
		extiplist, ok := node_map["extIPs"].([]interface{})
		if !ok {
			flatten_node["extip"] = false
			flatten_node["extipv6"] = false
		} else if ok {
			for _, extip := range extiplist {
				net.ParseIP(extip.(string)).To4()
				if net.ParseIP(extip.(string)).To4() != nil {
					flatten_node["extip"] = true
				} else if net.ParseIP(extip.(string)).To16() != nil {
					flatten_node["extipv6"] = true
				}
			}
		}
		flatten_node["fixedcloudlets"] = int(node_map["fixedCloudlets"].(float64))       // search value in customitem -> fixedCloudlets
		flatten_node["flexiblecloudlets"] = int(node_map["flexibleCloudlets"].(float64)) // search value in customitem -> search value in customitem -> flexibleCloudlets
		flatten_node["image"] = customitem["dockerName"].(string)                        // search value in customitem -> dockerName
		flatten_node["mission"] = node_map["nodemission"].(string)                       // search value in customitem -> nodemission
		flatten_node["nodegroup"] = node_map["nodeGroup"].(string)                       // search value in customitem -> nodeGroup
		flatten_node["nodetype"] = node_map["nodeType"].(string)                         // search value in customitem -> nodeType
		nodegroup := node_map["nodeGroup"].(string)
		for _, nodegroup_infos := range *nodegroups { // Retrieve values of restartdelay and scalingmode of one specific nodegroup
			nodegroup_infos_map := nodegroup_infos.(map[string]interface{})
			if nodegroup_infos_map["name"] == nodegroup {
				flatten_node["restartdelay"] = int(nodegroup_infos_map["restartNodeDelay"].(float64))
				flatten_node["scalingmode"] = nodegroup_infos_map["scalingMode"].(string)
			}
		}
		flatten_node["tag"] = customitem["dockerTag"].(string) // search value in customitem -> dockerTag
		// flatten_node["volumemounts"]
		node_volumes_len := len(customitem["dockerVolumes"].([]interface{}))
		var node_volumes = make([]string, node_volumes_len)
		for j, v := range customitem["dockerVolumes"].([]interface{}) {
			node_volume, _ := v.(string)
			node_volumes[j] = node_volume
		}
		flatten_node["volumes"] = node_volumes // search value in customitem -> dockerVolumes
		node_volumesfrom_len := len(customitem["dockerVolumesFrom"].([]interface{}))
		var node_volumesfrom = make([]string, node_volumesfrom_len)
		for j, v := range customitem["dockerVolumesFrom"].([]interface{}) {
			node_volumefrom, _ := v.(string)
			node_volumesfrom[j] = node_volumefrom
		}
		flatten_node["volumesfrom"] = node_volumesfrom

		flatten_nodes = append(flatten_nodes, flatten_node)
	}
	return flatten_nodes
}
