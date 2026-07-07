module github.com/ikascrew/server

go 1.25.0

require (
	github.com/ikascrew/core v0.0.0-20210324041206-fb346c8e5c80
	github.com/ikascrew/ikasbox v0.0.0-20210324033018-91003da54aed
	github.com/ikascrew/pb v0.0.0-20200229215417-95f0a80962e7
	github.com/ikascrew/plugin v0.0.0-20200715234203-87c9c5b19416
	github.com/shirou/gopsutil v2.20.6+incompatible
	gocv.io/x/gocv v0.38.0
	golang.org/x/net v0.23.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/grpc v1.30.0
)

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/monochromegane/argen v0.0.0-20150711140148-c37112f9dc50 // indirect
	github.com/monochromegane/goban v0.0.0-20141019070712-284a52313eb5 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20200710124503-20a17af7bd0e // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/ikascrew/core => ../core

replace github.com/ikascrew/plugin => ../plugin

replace github.com/ikascrew/ikasbox => ../ikasbox
