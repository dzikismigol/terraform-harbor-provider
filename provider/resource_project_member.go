package provider

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/bestsellerit/terraform-provider-harbor/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type members struct {
	RoleID      int   `json:"role_id"`
	GroupMember group `json:"member_group"`
}

type group struct {
	GroupType int    `json:"group_type"`
	GroupName string `json:"group_name"`
}

type entity struct {
	ID     int `json:"id"`
	RoleID int `json:"role_id"`
}

func resourceMembers() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"member_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "projectadmin" && v != "developer" && v != "guest" && v != "master" {
						errs = append(errs, fmt.Errorf("%q must be either projectadmin, developer, guest or master, got: %s", key, v))
					}
					return
				},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "ldap" && v != "internal" && v != "oidc" {
						errs = append(errs, fmt.Errorf("%q must be either ldap, internal or oidc, got: %s", key, v))
					}
					return
				},
			},
		},
		Create: resourceMembersCreate,
		Read:   resourceMembersRead,
		Update: resourceMembersUpdate,
		Delete: resourceMembersDelete,
	}
}

func groupType(group string) (x int) {
	switch group {
	case "ldap":
		x = 1
	case "internal":
		x = 2
	case "oidc":
		x = 3
	}
	return x
}

func roleTypeNumber(role int) (x string) {
	switch role {
	case 1:
		x = "projectadmin"
	case 2:
		x = "developer"
	case 3:
		x = "guest"
	case 4:
		x = "master"
	}
	return x
}

func roleType(role string) (x int) {
	switch role {
	case "projectadmin":
		x = 1
	case "developer":
		x = 2
	case "guest":
		x = 3
	case "master":
		x = 4
	}
	return x
}

func resourceMembersCreate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	projectid := checkProjectid(d.Get("project_id").(string))
	path := projectid + "/members"

	body := members{
		RoleID: roleType(d.Get("role").(string)),
		GroupMember: group{
			GroupType: groupType(d.Get("type").(string)),
			GroupName: d.Get("name").(string),
		},
	}

	_, err := apiClient.SendRequest("POST", path, body, 201)
	if err != nil {
		return err
	}

	return resourceMembersRead(d, m)
}

func resourceMembersRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	projectid := checkProjectid(d.Get("project_id").(string))
	path := projectid + "/members?entityname=" + d.Get("name").(string)

	resp, err := apiClient.SendRequest("GET", path, nil, 200)
	if err != nil {
		fmt.Println(err)
	}

	var entityData []entity
	json.Unmarshal([]byte(resp), &entityData)
	if err != nil {
		return fmt.Errorf("[ERROR] Unable to unmarshal: %s", err)
	}

	d.SetId(projectid + "/members/" + strconv.Itoa(entityData[0].ID))
	d.Set("member_id", entityData[0].ID)
	d.Set("role", roleTypeNumber(entityData[0].RoleID))
	return nil
}

func resourceMembersUpdate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	body := members{
		RoleID: roleType(d.Get("role").(string)),
	}

	_, err := apiClient.SendRequest("GET", d.Id(), body, 200)
	if err != nil {
		fmt.Println(err)
	}

	return resourceMembersRead(d, m)
}

func resourceMembersDelete(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	_, err := apiClient.SendRequest("DELETE", d.Id(), nil, 200)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
