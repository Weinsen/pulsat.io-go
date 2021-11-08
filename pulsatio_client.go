package pulsatio_client

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"net/http"
	"io"
	"fmt"
	"bytes"
	"time"
)

type Pulsatio struct {
	url string
	interval int
	id string
	data map[string]interface{}
	on map[string]func(string)
	_interval int
	_message_id string
	_active bool
	_connected bool
	_update bool
}

func New(id string, url string) (Pulsatio) {
	p := Pulsatio{}
	p.url = url
	p.id = id
	p._interval = 1 * 15000
	p._message_id = ""
	p._active = true
	p._connected = false
	p._update = false
	p.data = map[string]interface{}{}
	p.on = map[string]func(string){}
	p.data["interval"] = p._interval
	return p
}

func (p *Pulsatio) update() {
	p._update = true
}

func (p *Pulsatio) SetInterval(interval int) {
	if p._interval != interval {
		p._interval = interval
		p.data["interval"] = interval
		p.update()
	}
}

func (p *Pulsatio) errorHandler(e error) error {
	if cb, ok := p.on["error"]; ok {
	    cb(e.Error())
	}
	return e
}

func (p *Pulsatio) SetCallback(e string, f func(string)) (error) {
	p.on[e] = f
	return nil
}

func (p *Pulsatio) SetData(k string, v string) {
	if value, ok := p.data[k]; ok && value != v {
		p.data[k] = v
		p.update()
	}
}

func (p *Pulsatio) GetData(k string) string {
	if v, ok := p.data[k]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func (p *Pulsatio) ClearData(k string) {
	if _, ok := p.data[k]; ok {
		delete(p.data, k)
	}
}

func (p *Pulsatio) Register() (string, error) {
	json, _ := sjson.Set("", "id", p.id)
	for k, v := range p.data {
		json, _ = sjson.Set(json, k, v)
	}
	resp, err := p.doRequest("POST", json)
	if err != nil {
		return resp, p.errorHandler(err)
	}
	if cb, ok := p.on["connection"]; ok {
	    cb(resp)
	}
	if resp != "" {
		p._connected = true
	}
	p._update = false
	return resp, nil
}

func (p *Pulsatio) SendHeartBeat() (string, error) {
	json, _ := sjson.Set("", "id", "1")
	resp, err := p.doRequest("PUT", json)
	if err != nil {
		return resp, p.errorHandler(err)
	}
	if resp != "" {
		if msg_id := gjson.Get(resp, "_message_id"); msg_id.Exists() {
			message_id := msg_id.String()
			if message_id != p._message_id {
				p._message_id = message_id
				if cb, ok := p.on["heartbeat"]; ok {
				    cb(resp)
				}
				return resp, nil
			}
		}
	}
	return resp, nil

}

func (p *Pulsatio) Start() {
	go func() {
		for p._active {
			if p._connected && !p._update {
				p.SendHeartBeat()
			} else {
				p.Register()
			}
			time.Sleep(time.Duration(p._interval) * time.Millisecond)
		}
	}()
}

func (p *Pulsatio) Stop() {
	p._active = false
	p._connected = false
}

func (p *Pulsatio) doRequest(method string, data string) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(p._interval) * time.Millisecond,
	}

	url := p.url + "/nodes"
	if method == "PUT" {
		url += "/" + p.id
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", p.errorHandler(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return "", p.errorHandler(err)
	}

	if resp.StatusCode >= 300 {
		p._connected = false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", p.errorHandler(err)
	}
	defer resp.Body.Close()

	return string(body), nil
}
