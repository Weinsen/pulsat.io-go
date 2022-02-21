package pulsatio_client

import (
	"encoding/json"
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
	on map[string]func([]byte)
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
	p.on = map[string]func([]byte){}
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
	    cb([]byte(e.Error()))
	}
	return e
}

func (p *Pulsatio) SetCallback(e string, f func([]byte)) (error) {
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

func (p *Pulsatio) Register() ([]byte, error) {
	
	p.data["id"] = p.id
	req, err := json.Marshal(&p.data)
	if err != nil {
		return []byte{}, p.errorHandler(err)
	}

	body, err := p.doRequest("POST", req)
	if err != nil {
		return body, p.errorHandler(err)
	}
	if cb, ok := p.on["connection"]; ok {
	    cb(body)
	}
	if len(body) > 0 {
		p._connected = true
	}
	p._update = false
	return body, nil
}

func (p *Pulsatio) SendHeartBeat() ([]byte, error) {

	req_data := struct {
		id string `json: "id"`
	}{
		id: p.id,
	}

	req, err := json.Marshal(&req_data)
	if err != nil {
		return []byte{}, p.errorHandler(err)
	}

	body, err := p.doRequest("PUT", req)
	if err != nil {
		return body, p.errorHandler(err)
	}
	if len(body) > 0 {

		type Message struct {
			id string `json: "_message_id"`
		}

		msg := Message{}

		err := json.Unmarshal(body, &msg)
		if err != nil {
			return []byte{}, p.errorHandler(err)
		}

		if msg.id != p._message_id {
			p._message_id = msg.id
			if cb, ok := p.on["heartbeat"]; ok {
			    cb(body)
			}
			return body, nil
		}
	}
	return body, nil

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

func (p *Pulsatio) doRequest(method string, data []byte) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Duration(p._interval) * time.Millisecond,
	}

	url := p.url + "/nodes"
	if method == "PUT" {
		url += "/" + p.id
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return []byte{}, p.errorHandler(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, p.errorHandler(err)
	}

	if resp.StatusCode >= 300 {
		p._connected = false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, p.errorHandler(err)
	}
	defer resp.Body.Close()

	return body, nil
}
