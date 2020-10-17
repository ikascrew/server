package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ikascrew/ikasbox/handler"

	"golang.org/x/xerrors"
)

type Config struct {
	Port int

	DBIP   string
	DBPort int

	ProjectID int
	Width     int
	Height    int
	Default   Default
	Contents  map[int]*Content
}

type Content struct {
	ContentID int
	Name      string
	Path      string
}

type Default struct {
	Type string
	Name string
}

var gConf *Config

func init() {
	gConf = nil
}

func Set(p int, opts ...Option) error {

	conf := defaultConfig()
	for _, opt := range opts {
		err := opt(conf)
		if err != nil {
		}
	}

	err := load(p, conf)
	if err != nil {
		return xerrors.Errorf("project[%d] load error: %w", p, err)
	}

	gConf = conf

	return nil
}

func Get() *Config {
	return gConf
}

func defaultConfig() *Config {
	c := Config{}
	c.Port = 55555
	c.DBIP = "localhost"
	c.DBPort = 5555
	return &c
}

func load(p int, conf *Config) error {

	url := fmt.Sprintf("http://%s:%d/project/content/list/%d", conf.DBIP, conf.DBPort, p)
	resp, err := http.Get(url)
	if err != nil {
		return xerrors.Errorf("http get: %w", err)
	}

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("read: %w", err)
	}

	res := handler.ProjectResponse{}

	err = json.Unmarshal(byteArray, &res)
	if err != nil {
		return xerrors.Errorf("json unmarshal: %w", err)
	}

	def := Default{
		Type: "terminal",
		Name: "blank",
	}

	conf.Width = res.Project.Width
	conf.Height = res.Project.Height
	conf.Default = def

	conf.Contents = make(map[int]*Content)

	for _, elm := range res.Contents {
		con := Content{}
		con.Name = elm.Name
		con.Path = elm.Path
		con.ContentID = elm.ContentID

		conf.Contents[elm.ID] = &con
	}

	return nil
}
