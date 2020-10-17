package server

import (
	"fmt"

	"github.com/ikascrew/core"

	cd "github.com/ikascrew/plugin/countdown"
	file "github.com/ikascrew/plugin/file"
	img "github.com/ikascrew/plugin/image"
	term "github.com/ikascrew/plugin/terminal"

	"golang.org/x/xerrors"
)

var NotFoundError = fmt.Errorf("NotFound Video Type")

func Get(t string, params ...string) (core.Video, error) {

	var v core.Video
	var err error

	switch t {
	case "file":
		v, err = file.New(params...)
	case "img":
		v, err = img.New(params...)
	case "cd":
		v, err = cd.New(params...)
	case "terminal":
		v, err = term.New(params...)
	}

	if err != nil {
		return nil, xerrors.Errorf("video new[%s]: %w", t, err)
	}

	if v == nil {
		return nil, NotFoundError
	}
	return v, nil
}
