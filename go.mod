module github.com/ikascrew/server

go 1.16

require (
	github.com/ikascrew/core v0.0.0-20210324041206-fb346c8e5c80
	github.com/ikascrew/ikasbox v0.0.0-20210324033018-91003da54aed
	github.com/ikascrew/pb v0.0.0-20200229215417-95f0a80962e7
	//github.com/ikascrew/plugin v0.0.0-20211230065329-b13846bf31f9 // indirect
	github.com/ikascrew/plugin v0.0.0-20200715234203-87c9c5b19416
	github.com/shirou/gopsutil v2.20.6+incompatible
	github.com/stretchr/testify v1.7.0 // indirect
	gocv.io/x/gocv v0.29.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20200710124503-20a17af7bd0e // indirect
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/ikascrew/plugin => /home/secondarykey/GoApp/ikascrew/plugin
