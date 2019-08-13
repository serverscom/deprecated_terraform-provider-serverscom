package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"strings"
)

type HostsList struct {
	Data []Host `json:"data"`
}

type Network struct {
	Id				   int    `json:"id"`
	HostIp             string `json:"host_ip"`
	PoolType           string `json:"pool_type"`
	Size               int    `json:"size"`
	Netmask            string `json:"netmask"`
}

type Location struct {
	Id				   int        `json:"id"`
	Name               string     `json:"name"`
}

type Host struct {
	Id                 int         `json:"id"`
	Type               int         `json:"type"`
	Title              string      `json:"title"`
	Conf               string      `json:"conf"`
	ServiceType        int         `json:"service_type"`
	LeaseEnd           interface{} `json:"lease_end"`
	Networks           []Network   `json:"networks"`
	Location           Location    `json:"location"`
	ProjectId          interface{} `json:"project_id"`
	ProjectName        interface{} `json:"project_name"`
	ScheduledReleaseAt interface{} `json:"scheduled_release_at"`
	RackName           interface{} `json:"rack_name"`
	RackId             interface{} `json:"rack_id"`
	L2Segments         interface{} `json:"l2_segments"`
}

type OrdersList struct {
	Data []Order `json:"data"`
}

type Order struct {
	Amount             float64  `json:"amount"`
	AmountTax          float64  `json:"amount_tax"`
	AmountTotal        float64  `json:"amount_total"`
	CreatedTime        string   `json:"created_time"`
	Currency           string   `json:"currency"`
	Description        []string `json:"description"`
	Id                 int      `json:"id"`
	OriginalAmount     float64  `json:"original_amount"`
	OriginalCurrency   string   `json:"original_currency"`
	Status             int      `json:"status"`
}

func AddToCart(url, email, token, hostname, config string) error {
	url = fmt.Sprintf(`%s/rest/server_cart_items`, url)
	data := strings.NewReader(fmt.Sprintf(config, hostname))
	_, err := GetResponse(url, email, token, "POST", data)
	return err
}

func CheckoutOrder(url, email, token string) error {
	data := strings.NewReader("{\"ts\":1456817777230}")
	_, err := GetResponse(fmt.Sprintf("%s/rest/orders", url),
		email, token, "POST", data)
	return err
}

func CancelServer(url string, serverid int, email, pwd, token string) error {
	data := strings.NewReader(fmt.Sprintf("{\"token\":\"%s\"}", pwd))
	_, err := GetResponse(fmt.Sprintf("%s/rest/hosts/%d/schedule_release", url, serverid),
		email, token, "POST", data)
	return err
}

func GetServers(url, email, token string) ([]Host, error) {
	body, err := GetResponse(fmt.Sprintf(`%s/rest/hosts`, url), email, token, "GET", nil)
	if err != nil {
		return nil, err
	}
	var hosts HostsList
	errUnmarshal := json.Unmarshal([]byte(string(*body)), &hosts)
	if errUnmarshal != nil {
		return []Host{}, errUnmarshal
	}
	return hosts.Data, err
}

func GetPendingServers(url, email, token string) ([]Host, error) {
	body, err := GetResponse(fmt.Sprintf(`%s/rest/hosts_pending`, url), email, token, "GET", nil)
	if err != nil {
		return nil, err
	}
	var hosts HostsList
	errUnmarshal := json.Unmarshal([]byte(string(*body)), &hosts)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}
	return hosts.Data, err
}

func GetOrders(url, email, token string) ([]Order, error) {
	body, err := GetResponse(fmt.Sprintf(`%s/rest/orders`, url), email, token, "GET", nil)
	if err != nil {
		return nil, err
	}
	var orders OrdersList
	errUnmarshal := json.Unmarshal([]byte(string(*body)), &orders)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}
	for _, order := range orders.Data {
		fmt.Println(fmt.Sprintf("host: %d", order.Id))
	}
	return orders.Data, nil
}

func GetServer(url, email, token, hostname string) (*Host, error) {
	body, err := GetResponse(fmt.Sprintf(`%s/rest/hosts?title=%s`, url, hostname), email, token, "GET", nil)
	if err != nil {
		return nil, err
	}
	var hosts HostsList
	errUnmarshal := json.Unmarshal([]byte(string(*body)), &hosts)
	if errUnmarshal != nil {
		hosts = HostsList{}
		return nil, errUnmarshal
	}
	if len(hosts.Data) > 1 {
		hosts = HostsList{}
		return nil, errors.New(fmt.Sprintf("Hostname: %s is not unique.", hostname))
	} else if len(hosts.Data) == 1 {
		return &hosts.Data[0], err
	} else {
		return nil, err
	}
}

func GetPendingServer(url, email, token, hostname string) (*Host, error) {
	hosts, err := GetPendingServers(url, email, token)
	if err != nil {
		return nil, err
	}
	for _, h := range hosts {
		if h.Title == hostname {
			return &h, nil
		}
	}
	return nil, nil
}

func IsServerOrOrderExists(url, email, token, hostname string) (bool, error) {
	s, err := GetServer(url, email, token, hostname)
	if err != nil {
		return false, err
	}
	if s != nil {
		return true, nil
	}

	s, err = GetPendingServer(url, email, token, hostname)
	if err != nil {
		return false, err
	}
	if s != nil {
		return true, nil
	}

	orders, err := GetOrders(url, email, token)
	if err != nil {
		return false, err
	}
	for _, order := range orders {
		if order.Status != 2 {
			d := order.Description
			for _, h := range d {
				if h == hostname {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	hostname := d.Get("hostname").(string)
	isExist, err := IsServerOrOrderExists(url, email, token, hostname)
	if err != nil {
		return err
	}
	if isExist {
		return errors.New(fmt.Sprintf("Order cannot be created. Hostname: %s is not unique.", hostname))
	}
	config := d.Get("config").(string)
	err = AddToCart(url, email, token, hostname, config)
	if err != nil {
		return err
	}
	err = CheckoutOrder(url, email, token)
	if err != nil {
		return err
	}
	d.SetId(hostname)
	d.Set("hostname", hostname)
	return nil
	//return resourceServerRead(d, m)
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	hostname := d.Get("hostname").(string)
	isExist, err := IsServerOrOrderExists(url, email, token, hostname)
	if err != nil {
		return err
	}

	if isExist {
		d.Set("hostname", hostname)
		return nil
	} else {
		d.SetId("")
		return errors.New(fmt.Sprintf("Hostname: %s not found.", hostname))
	}
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	url := m.(*Client).Url
	email := m.(*Client).Email
	token := m.(*Client).Token
	pwd := m.(*Client).Pwd
	hostname := d.Get("hostname").(string)
	s, err := GetServer(url, email, token, hostname)

	if err != nil {
		return err
	}

	if s != nil && s.ScheduledReleaseAt == nil {
		err = CancelServer(url, s.Id, email, pwd, token)
		if err != nil {
			return err
		}
		d.SetId("")
	} else if s != nil && s.ScheduledReleaseAt != nil {
		return errors.New(fmt.Sprintf("Server %s is already in cancellation state!", hostname))
	} else {
		return errors.New(fmt.Sprintf("Server %s cannot be released!", hostname))
	}
	return resourceServerRead(d, m)
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	if d.HasChange("hostname") {
		url := m.(*Client).Url
		email := m.(*Client).Email
		token := m.(*Client).Token
		pwd := m.(*Client).Pwd
		hostname := d.Id()
		s, err := GetServer(url, email, token, hostname)
		if err != nil {
			return err
		}
		if s != nil && s.ScheduledReleaseAt == nil {
			err = CancelServer(url, s.Id, email, pwd, token)
			if err != nil {
				return err
			}
			d.SetId("")
		} else if s != nil && s.ScheduledReleaseAt != nil {
			return errors.New(fmt.Sprintf("Server %s is already in cancellation state!", hostname))
		} else {
			return errors.New(fmt.Sprintf("Server %s cannot be updated!", hostname))
		}
		hostname = d.Get("hostname").(string)
		isExist, err := IsServerOrOrderExists(url, email, token, hostname)
		if err != nil {
			return err
		}
		if isExist {
			return errors.New(fmt.Sprintf("Order cannot be created. Hostname: %s is not unique.", hostname))
		}
		err = AddToCart(url, email, token, hostname, d.Get("config").(string))
		if err != nil {
			return err
		}
		err = CheckoutOrder(url, email, token)
		if err != nil {
			return err
		}
		d.SetId(hostname)
		d.SetPartial("hostname")
	}
	d.Partial(false)

	return resourceServerRead(d, m)
}


func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Delete: resourceServerDelete,
		Update: resourceServerUpdate,

		Schema: map[string]*schema.Schema{
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.NoZeroValues,
			},
			"config": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}