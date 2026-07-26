package server

import (
	"log"

	"github.com/ikascrew/ikasbox"
	ibconfig "github.com/ikascrew/ikasbox/config"
	"github.com/ikascrew/ikasbox/handler/api"
	"github.com/ikascrew/plugin/video/output"
	"github.com/ikascrew/server/config"

	"golang.org/x/xerrors"
)

// startIkasbox は ikasbox(HTTP :5555)を同一プロセス内で同居起動し、
// React UI から server の work file を作成できるエンドポイントを
// 登録する。work file 形式の知識を ikasbox 側へ持ち込まないため、
// エンドポイントの実装と登録は server 側から行う。
// 同居モード中は単体の ikasbox を並行起動しないこと(SQLite の排他)
func startIkasbox(db string) {

	// ルーティングは排他していないため handler.Listen より先に登録する
	api.AddEndpoint("v1/server/status", newServerStatus)
	api.AddEndpoint("v1/server/create", newServerCreate)

	go func() {
		err := ikasbox.Start(
			ibconfig.Argument([]string{"start"}),
			ibconfig.Path(db),
		)
		if err != nil {
			log.Printf("ikasbox error: %+v", err)
		}
	}()
}

// serverStatus は同居モードの存在確認と現在の状態を返す。
// UI はこの API が成功したときだけ「Server Create」ボタンを表示する
type serverStatus struct{}

func newServerStatus() api.Parameter {
	return &serverStatus{}
}

type serverStatusReturn struct {
	Success   bool `json:"success"`
	ProjectID int  `json:"projectId"`
	Playing   bool `json:"playing"`
}

func (r *serverStatusReturn) IsSuccess() bool { return r.Success }

func (p *serverStatus) Processing() (api.Return, error) {

	rtn := serverStatusReturn{Success: true}

	if conf := config.Get(); conf != nil {
		rtn.ProjectID = conf.ProjectID
	}
	if server != nil && server.window != nil {
		rtn.Playing = server.window.isPlaying()
	}

	return &rtn, nil
}

// serverCreate は指定プロジェクトで work file を作り直し、
// 走行中の設定へホットリロードする
type serverCreate struct {
	ProjectID int `json:"projectId"`
}

func newServerCreate() api.Parameter {
	return &serverCreate{}
}

type serverCreateReturn struct {
	Success   bool `json:"success"`
	ProjectID int  `json:"projectId"`
}

func (r *serverCreateReturn) IsSuccess() bool { return r.Success }

func (p *serverCreate) Processing() (api.Return, error) {

	if p.ProjectID <= 0 {
		return nil, xerrors.New("projectId is required")
	}

	if server == nil || server.window == nil {
		return nil, xerrors.New("server is not ready")
	}

	// 本番中(フルスクリーン)はマッピングを凍結する
	// (フルスクリーン中は ESC / Ctrl+C を無効化する既存方針と同じ)
	if server.window.isPlaying() {
		return nil, xerrors.New("now playing (fullscreen): create is disabled")
	}

	// 同一プロセス内の ikasbox(:5555)へのループバックで、
	// CLI の create と完全に同じ経路・同じデータを通す
	err := config.Create(p.ProjectID)
	if err != nil {
		return nil, xerrors.Errorf("create work file: %w", err)
	}

	err = config.Reload()
	if err != nil {
		return nil, xerrors.Errorf("reload config: %w", err)
	}

	// 生成型プラグインの描画解像度も新プロジェクトへ追随させる。
	// 再生中の動画は自分のインスタンスを持つため影響しない
	conf := config.Get()
	output.Set(conf.Width, conf.Height)

	log.Printf("work file created and reloaded: project=%d (%dx%d)",
		p.ProjectID, conf.Width, conf.Height)

	return &serverCreateReturn{Success: true, ProjectID: p.ProjectID}, nil
}
