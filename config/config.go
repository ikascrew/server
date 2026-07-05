package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ikascrew/ikasbox/handler"

	"golang.org/x/xerrors"
)

const (
	workDir  = ".server"
	workFile = "config.json"
)

func WorkPath() string {
	return filepath.Join(workDir, workFile)
}

type Config struct {
	Port int

	DBIP   string
	DBPort int

	ProjectID int
	Width     int
	Height    int
	Default   Default
	Contents  map[int]*Content

	Headless bool
	Verbose  bool
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

// Create は ikasbox からプロジェクト情報を取得しワークファイルを作成する
func Create(p int, opts ...Option) error {

	conf := defaultConfig()
	for _, opt := range opts {
		err := opt(conf)
		if err != nil {
			return xerrors.Errorf("option error: %w", err)
		}
	}

	err := load(p, conf)
	if err != nil {
		return xerrors.Errorf("project[%d] load error: %w", p, err)
	}

	err = os.MkdirAll(workDir, 0755)
	if err != nil {
		return xerrors.Errorf("create work directory(%s): %w", workDir, err)
	}

	buf, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return xerrors.Errorf("json marshal: %w", err)
	}

	err = ioutil.WriteFile(WorkPath(), buf, 0644)
	if err != nil {
		return xerrors.Errorf("write work file(%s): %w", WorkPath(), err)
	}

	return nil
}

// Set はワークファイルから設定を読み込む(ikasbox には接続しない)
func Set(opts ...Option) error {

	conf := defaultConfig()

	buf, err := ioutil.ReadFile(WorkPath())
	if err != nil {
		return xerrors.Errorf("read work file(%s). run \"create\" first: %w", WorkPath(), err)
	}

	err = json.Unmarshal(buf, conf)
	if err != nil {
		return xerrors.Errorf("work file unmarshal: %w", err)
	}

	for _, opt := range opts {
		err := opt(conf)
		if err != nil {
			return xerrors.Errorf("option error: %w", err)
		}
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

	fmt.Println(url)
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

	conf.ProjectID = p
	conf.Width = res.Project.Width
	conf.Height = res.Project.Height
	conf.Default = def

	conf.Contents = make(map[int]*Content)

	for _, elm := range res.Contents {
		con := Content{}
		con.Name = elm.Name
		con.Path = elm.Path
		con.ContentID = elm.ID

		conf.Contents[elm.ID] = &con

		//fmt.Printf("%d=[%s][%s]\n", elm.ID, elm.Name, elm.Path)
	}

	return nil
}
