package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"strings"
)

type PtrData struct {
	Data Ptr `json:"data"`
}

type PtrDataList struct {
	Data []Ptr `json:"data"`
}

type Ptr struct {
	Id           int         `json:"id"`
	DomainId     int         `json:"domain_id"`
	Type         string      `json:"type"`
	Name         string      `json:"name"`
	Ttl          int         `json:"ttl"`
	Priority     int         `json:"priority"`
	Data         interface{} `json:"data"`
	Disabled     bool        `json:"disabled"`
}

func AddPtr(url, email, token, data, name string) (*Ptr, error) {
	body, err := GetResponse(fmt.Sprintf("%s/rest/dns/records//", url),
		email, token, "POST",
		strings.NewReader(fmt.Sprintf(`{"data":"%s","name":"%s"}`, data, name)))
	if err != nil {
		return nil, err
	}
	var ptrData PtrData
	err = json.Unmarshal([]byte(string(*body)), &ptrData)
	if err != nil {
		return nil, err
	}
	return &ptrData.Data, nil
}

func ListPtr(url, email, token string) (*[]Ptr, error) {
	body, err := GetResponse(fmt.Sprintf("%s/rest/dns/records//", url),
		email, token, "GET", nil)
	if err != nil {
		return nil, err
	}
	var ptrData PtrDataList
	err = json.Unmarshal([]byte(string(*body)), &ptrData)
	if err != nil {
		return nil, err
	}
	return &ptrData.Data, nil
}

func GetPtr(url, email, token, id string) (*Ptr, error) {
	ptrs, err := ListPtr(url, email, token)
	if err != nil {
		return nil, err
	}
	for _, ptr := range *ptrs {
		if fmt.Sprintf("%d", ptr.Id) == id {
			return &ptr, nil
		}
	}
	return nil, nil
}

func DeletePtr(url, email, token, ptrId, domainId string) error {
	_, err := GetResponse(fmt.Sprintf(`%s/rest/dns/records///%s`, url, ptrId),
		email, token, "DELETE",
		strings.NewReader(fmt.Sprintf(`{"domain_id":%s}`, domainId)))
	return err
}

func GetHostNetwork(url, email, token, hostname string) (*Network, error) {
	s, err := GetServer(url, email, token, hostname)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, errors.New(fmt.Sprintf("Hostname: %s doesn't exist or in pending status.", hostname))
	}
	for _, network := range s.Networks {
		if network.PoolType == "public" {
			return &network, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Hostname: %s doesn't have public network.", hostname))
}

func resourcePtrCreate(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	hostname := d.Get("hostname").(string)
	ptrAddress := d.Get("ptr").(string)
	network, err := GetHostNetwork(url, email, token, hostname)
	if err != nil {
		return err
	}
	ptr, err := AddPtr(url, email, token, ptrAddress, network.HostIp)
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%d", ptr.Id))
	d.Set("hostname", hostname)
	d.Set("ptr", ptrAddress)
	d.Set("domainId", fmt.Sprintf("%d", ptr.DomainId))
	return resourcePtrRead(d, m)
}

func resourcePtrRead(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	ptr, err := GetPtr(url, email, token, d.Id())
	if err != nil {
		return err
	}
	if ptr == nil {
		d.SetId("")
		d.Set("ptr", "")
		d.Set("domainId", "")
		d.Set("hostname", "")
		return errors.New(fmt.Sprintf("Ptr not found."))
	} else {
		d.Set("ptr", d.Get("ptr").(string))
		d.Set("domainId", fmt.Sprintf("%d", ptr.DomainId))
		d.Set("hostname", d.Get("hostname").(string))
		return nil
	}
}

func resourcePtrDelete(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	ptr, err := GetPtr(url, email, token, d.Id())
	if err != nil {
		return err
	}
	if err = DeletePtr(url, email, token, d.Id(), fmt.Sprintf("%d", ptr.DomainId)); err != nil{
		return err
	}
	d.SetId("")
	return resourcePtrRead(d, m)
}

func resourcePtrUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	if d.HasChange("hostname") || d.HasChange("ptr") {
		url := m.(*Client).Url
		email := m.(*Client).Email
		token := m.(*Client).Token
		ptr, err := GetPtr(url, email, token, d.Id())
		if err != nil {
			return err
		}
		if err = DeletePtr(url, email, token, d.Id(), fmt.Sprintf("%d", ptr.DomainId)); err != nil {
			return err
		}
		hostname := d.Get("hostname").(string)
		ptrAddress := d.Get("ptr").(string)
		network, err := GetHostNetwork(url, email, token, hostname)
		if err != nil {
			return err
		}
		ptr, err = AddPtr(url, email, token, ptrAddress, network.HostIp)
		if err != nil {
			return err
		}
		d.SetId(fmt.Sprintf("%d", ptr.Id))
		d.SetPartial("hostname")
		d.SetPartial("ptr")
		d.Set("domainId", fmt.Sprintf("%d", ptr.DomainId))
	}
	d.Partial(false)

	return resourcePtrRead(d, m)
}

func resourcePtr() *schema.Resource {
	return &schema.Resource{
		Create: resourcePtrCreate,
		Read:   resourcePtrRead,
		Delete: resourcePtrDelete,
		Update: resourcePtrUpdate,

		Schema: map[string]*schema.Schema{
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.NoZeroValues,
			},
			"ptr": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}