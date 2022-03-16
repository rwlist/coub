package coub

import (
	"bufio"
	"bytes"
	"net/http"

	"gorm.io/gorm"
)

type KV struct {
	Key   string `gorm:"primarykey"`
	Value string
}

const headersKey = "http_headers"

type Cookies struct {
	db *gorm.DB
}

func NewCookies(db *gorm.DB) *Cookies {
	return &Cookies{
		db: db,
	}
}

func (c *Cookies) AutoMigrate() error {
	err := c.db.AutoMigrate(&KV{})
	if err != nil {
		return err
	}

	_, err = c.GetOrSet(headersKey, "")
	return err
}

func (c *Cookies) GetOrSet(key, defaultValue string) (string, error) {
	kv := KV{}
	err := c.db.Where("key = ?", key).First(&kv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			kv.Key = key
			kv.Value = defaultValue
			err = c.db.Create(&kv).Error
			if err != nil {
				return "", err
			}
			return defaultValue, nil
		}
		return "", err
	}
	return kv.Value, nil
}

func (c *Cookies) Get() (*State, error) {
	headers, err := c.GetOrSet(headersKey, "")
	if err != nil {
		return nil, err
	}

	return ParseState(headers)
}

func (c *Cookies) Update(req *http.Request, resp *http.Response) error {
	// TODO:
	return nil
}

type State struct {
	Request *http.Request
}

func ParseState(headers string) (*State, error) {
	b := bytes.NewBuffer([]byte(headers))
	b.WriteString("\n\n")
	req, err := http.ReadRequest(bufio.NewReader(b))
	if err != nil {
		return nil, err
	}

	return &State{
		Request: req,
	}, nil
}
