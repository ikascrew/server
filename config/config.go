package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

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

	// ikasbox 同居モード(-ikasbox)。実行時フラグであり
	// work file には保存しない
	Ikasbox   bool   `json:"-"`
	IkasboxDB string `json:"-"`
}

type Content struct {
	ContentID int
	Name      string
	Path      string
	Type      string
	Params    string
}

type Default struct {
	Type string
	Name string
}

var gConf *Config

// gRPC ハンドラと同居 ikasbox の作成 API が並行に触るため、
// gConf の差し替え(Reload)と参照は排他する
var gMu sync.RWMutex

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

// Set はワークファイルから設定を読み込む(ikasbox には接続しない)。
// ikasbox 同居モード(Ikasbox オプション)のときだけは work file が
// 無くても既定値で起動を許す(初回は UI の create で作成するため)
func Set(opts ...Option) error {

	conf := defaultConfig()

	buf, readErr := ioutil.ReadFile(WorkPath())
	if readErr == nil {
		err := json.Unmarshal(buf, conf)
		if err != nil {
			return xerrors.Errorf("work file unmarshal: %w", err)
		}
	}

	for _, opt := range opts {
		err := opt(conf)
		if err != nil {
			return xerrors.Errorf("option error: %w", err)
		}
	}

	if readErr != nil {
		if !conf.Ikasbox {
			return xerrors.Errorf("read work file(%s). run \"create\" first: %w", WorkPath(), readErr)
		}
		fmt.Printf("work file not found(%s): starting empty. create it from the ikasbox UI\n", WorkPath())
	}

	gMu.Lock()
	gConf = conf
	gMu.Unlock()

	return nil
}

// Reload は work file を読み直して設定を差し替える(ホットリロード)。
// 実行時のみのフラグ(Ikasbox 等)は現在の設定から引き継ぐ。
// 呼び出し側はスナップショット(Get の返り値)を使い続けられるため、
// 再生中の動画には影響せず、以後の参照から新しい設定が効く
func Reload() error {

	conf := defaultConfig()

	buf, err := ioutil.ReadFile(WorkPath())
	if err != nil {
		return xerrors.Errorf("read work file(%s): %w", WorkPath(), err)
	}

	err = json.Unmarshal(buf, conf)
	if err != nil {
		return xerrors.Errorf("work file unmarshal: %w", err)
	}

	gMu.Lock()
	if gConf != nil {
		conf.Ikasbox = gConf.Ikasbox
		conf.IkasboxDB = gConf.IkasboxDB
		conf.Headless = gConf.Headless
		conf.Verbose = gConf.Verbose
	}
	gConf = conf
	gMu.Unlock()

	return nil
}

func Get() *Config {
	gMu.RLock()
	defer gMu.RUnlock()
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

	if resp.StatusCode != http.StatusOK {
		return xerrors.Errorf("ikasbox response %s: project[%d] not found?", resp.Status, p)
	}

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
		con.Type = elm.Type
		con.Params = elm.Params

		conf.Contents[elm.ID] = &con

		//fmt.Printf("%d=[%s][%s]\n", elm.ID, elm.Name, elm.Path)
	}

	return nil
}
