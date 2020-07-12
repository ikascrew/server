package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ikascrew/ikasbox/handler"

	"golang.org/x/xerrors"
)

type AppConfig struct {
	Width   int
	Height  int
	Default Default

	Contents map[int]*Content
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

var conf *AppConfig

func init() {
	conf = nil
}

func Get() *AppConfig {
	return conf
}

const boxIP = "localhost"
const boxPort = "5555"

func Load(p int) error {

	url := fmt.Sprintf("http://"+boxIP+":"+boxPort+"/project/content/list/%d", p)
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

	app := AppConfig{
		Width:   res.Project.Width,
		Height:  res.Project.Height,
		Default: def,
	}

	app.Contents = make(map[int]*Content)

	for _, elm := range res.Contents {
		con := Content{}
		con.Name = elm.Name
		con.Path = elm.Path
		con.ContentID = elm.ContentID

		app.Contents[elm.ID] = &con
	}

	conf = &app

	return nil
}
