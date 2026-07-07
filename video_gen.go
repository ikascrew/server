package server

import (
	"github.com/ikascrew/core"
	"github.com/ikascrew/plugin/video"
)

var NotFoundError = video.NotFoundError

// Get は型名と JSON param から Video を生成する。
// 対応表と型名の正規化は plugin/video レジストリに集約されている
func Get(t string, param string) (core.Video, error) {
	return video.Get(t, param)
}
