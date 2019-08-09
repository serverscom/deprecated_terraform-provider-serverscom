package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"strings"
)

type L2DataListResp struct {
	Data *[]L2Resp `json:"data"`
}

type L2DataResp struct {
	Data *L2Resp `json:"data"`
}

type L2Resp struct {
	Id         int           `json:"id"`
	Hosts      *[]L2HostResp `json:"hosts"`
	Location   interface{}   `json:"location"`
	Name       string        `json:"name"`
	Status     string        `json:"status"`
	Type       int           `json:"type"`
}

type L2HostResp struct {
	Id     int           `json:"id"`
	Mode   string        `json:"mode"`
	Title  string        `json:"title"`
	Vlan   interface{}   `json:"vlan"`
}

func ListL2(url, email, token string) (*[]L2Resp, error) {
	body, err := GetResponse(fmt.Sprintf("%s/rest/l2_segments", url),
		email, token, "GET", nil)
	if err != nil {
		return nil, err
	}
	var l2Data L2DataListResp
	err = json.Unmarshal([]byte(string(*body)), &l2Data)
	if err != nil {
		return nil, err
	}
	fmt.Println(fmt.Sprintf("Segments len: %d", len(*l2Data.Data)))
	for _, l2 := range *l2Data.Data {
		fmt.Println(fmt.Sprintf("Id: %d, Status: %s", l2.Id, l2.Status))
	}
	return l2Data.Data, nil
}

type L2Req struct {
	DeleteIps   interface{}    `json:"delete_ips"`
	Hosts       *[]L2HostReq   `json:"hosts"`
	LocationId  int            `json:"location_id"`
	Name        string         `json:"name"`
	Type        int            `json:"type"`
}

type L2HostReq struct {
	Id     int           `json:"id"`
	Mode   string        `json:"mode"`
}

func CreateL2(url, email, token, name string, l2Type int, hostNames []HostnameWithType) (*L2Resp, error) {
	return CreateUpdateL2(url, fmt.Sprintf("%s/rest/l2_segments/", url), "POST",
		email, token, name, l2Type, hostNames)
}

func UpdateL2(url, email, token, name, id string, l2Type int, hostNames []HostnameWithType) (*L2Resp, error) {
	return CreateUpdateL2(url, fmt.Sprintf("%s/rest/l2_segments/%s", url, id), "PUT",
		email, token, name, l2Type, hostNames)
}

func CreateUpdateL2(url, fullUrl, method, email, token, name string, l2Type int, hostNames []HostnameWithType) (*L2Resp, error) {
	l2req, err := GetL2ReqData(url, email, token, name, l2Type, hostNames)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(l2req)
	if err != nil {
		return nil, err
	}
	body, err := GetResponse(fullUrl, email, token, method, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	var l2Data L2DataResp
	err = json.Unmarshal([]byte(string(*body)), &l2Data)
	if err != nil {
		return nil, err
	}
	return l2Data.Data, nil
}

type HostnameWithType struct {
	Name     string
	Mode     string
}

func GetL2ReqData(url, email, token, name string, l2type int, hostnames []HostnameWithType) (*L2Req, error) {
	data, err := GetServers(url, email, token)
	if err != nil {
		return nil ,err
	}
	locations := []int{}
	hosts := []L2HostReq{}
	for _, hostname := range hostnames {
		for _, server := range data {
			if server.Title == hostname.Name {
				if len(locations) > 0 && server.Location.Id != locations[0] {
					locations = append(locations, server.Location.Id)
				} else if len(locations) == 0 {
					locations = append(locations, server.Location.Id)
				}
				hosts = append(hosts, L2HostReq{Id: server.Id, Mode: hostname.Mode})
			}
		}
	}
	if len(locations) > 1 {
		return nil, errors.New("Hosts have different locations.")
	}
	if len(hosts) != len(hostnames) {
		return nil, errors.New("Not all hosts are ready.")
	}
	out := &L2Req { Hosts: &hosts, LocationId: locations[0], Name: name, Type: l2type }
	return out, nil
}

func GetL2(url, email, token, id string) (*L2Resp, error) {
	l2list, err := ListL2(url, email, token)
	if err != nil {
		return nil, err
	}
	for _, l2 := range *l2list {
		if fmt.Sprintf("%d", l2.Id) == id {
			return &l2, nil
		}
	}
	return nil, nil
}

type Success struct {
	Success bool `json:"success"`
}

func DeleteL2(url, email, token, id string) (bool, error) {
	body, err := GetResponse(fmt.Sprintf("%s/rest/l2_segments/%s/", url, id),
		email, token, "DELETE", nil)
	if err != nil {
		return false, err
	}
	var success Success
	err = json.Unmarshal([]byte(string(*body)), &success)
	if err != nil {
		return false, nil
	}
	return success.Success, nil
}

type hostsData []HostnameWithType

func retrieveHostNames(list interface{}) ([]string, []HostnameWithType, error) {
	hostNamesWithType := make(hostsData, len(list.([]interface{})))
	var hostNames []string
	for i, v := range list.([]interface{}) {
		p, castOk := v.(map[string]interface{})
		if !castOk {
			return nil, nil, fmt.Errorf("Unable to parse parts in hostnames resource declaration")
		}
		host := HostnameWithType{}
		if p, ok := p["name"]; ok {
			host.Name = p.(string)
			hostNames = append(hostNames, host.Name)
		}
		if p, ok := p["mode"]; ok {
			host.Mode = p.(string)
		}
		hostNamesWithType[i] = host
	}
	return hostNames, hostNamesWithType, nil
}

func resourceL2Create(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	name := d.Get("name").(string)
	hostNames, hostNamesWithType, err := retrieveHostNames(d.Get("hostnames"))
	if err != nil {
		return err
	}
	l2Type, err := getType(d.Get("type").(string))
	if err != nil {
		return err
	}
	r, err := CreateL2(url, email, token, name, l2Type, hostNamesWithType)
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%d", r.Id))
	d.Set("name", name)
	d.Set("hostnames", strings.Join(hostNames[:], ","))
	return resourceL2Read(d, m)
}

func resourceL2Read(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	id := d.Id()
	l2, err := GetL2(url, email, token, id)
	if err != nil {
		return err
	}
	if l2 == nil {
		d.SetId("")
		d.Set("name", "")
		d.Set("hostnames", "")
		return errors.New(fmt.Sprintf("L2 not found."))
	} else {
		d.Set("name", d.Get("name").(string))
		hostNames, _, err := retrieveHostNames(d.Get("hostnames"))
		if err != nil {
			return err
		}
		d.Set("hostnames", strings.Join(hostNames, ","))
		return nil
	}
}

func resourceL2Delete(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	id := d.Id()
	l2, err := GetL2(url, email, token, id)
	if err != nil {
		return err
	}
	if l2 != nil && l2.Status == "active" {
		_, err := DeleteL2(url, email, token, id)
		if err != nil {
			return err
		}
		d.SetId("")
		d.Set("name", "")
		d.Set("hostnames", "")
	} else if l2 != nil && l2.Status != "active" {
		return errors.New(fmt.Sprintf("Cannot delete %s segment, because of it's status.", l2.Name))
	} else if l2 == nil {
		return errors.New(fmt.Sprintf("Cannot delete segment with id: %s.", id))
	}
	return resourceL2Read(d, m)
}

func getType(typeName string) (int, error) {
	if typeName == "public" {
		return 1, nil
	} else if typeName == "private" {
		return 0, nil
	} else {
		return 0, errors.New(fmt.Sprintf(`Unrecognized type: %s. Please provide "public" or "private".`, typeName))
	}
}

func resourceL2Update(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	if d.HasChange("name") || d.HasChange("hostnames") {
		url := m.(*Client).Url
		email := m.(*Client).Email
		token := m.(*Client).Token
		id := d.Id()
		l2Type, err := getType(d.Get("type").(string))
		if err != nil {
			return err
		}
		l2, err := GetL2(url, email, token, id)
		if err != nil {
			return err
		}
		if l2 != nil && l2.Status == "active" {
			url := m.(*Client).Url
			email := m.(*Client).Email
			token := m.(*Client).Token
			name := d.Get("name").(string)
			_, hostNamesWithType, err := retrieveHostNames(d.Get("hostnames"))
			if err != nil {
				return err
			}
			_, err = UpdateL2(url, email, token, name, d.Id(), l2Type, hostNamesWithType)
			if err != nil {
				return err
			}
		} else if l2 != nil && l2.Status != "active" {
			return errors.New(fmt.Sprintf("Cannot update %s segment, because of it's status.", l2.Name))
		} else if l2 == nil {
			return errors.New(fmt.Sprintf("Cannot update segment with id: %s.", id))
		}
		d.SetPartial("name")
		d.SetPartial("hostnames")
	}
	d.Partial(false)
	return resourceL2Read(d, m)
}

func resourceL2() *schema.Resource {
	return &schema.Resource{
		Create: resourceL2Create,
		Read:   resourceL2Read,
		Delete: resourceL2Delete,
		Update: resourceL2Update,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.NoZeroValues,
			},
			"hostnames": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				MinItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.NoZeroValues,
						},
						"mode": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default: "native",
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default: "public",
			},
		},
	}
}
