module github.com/danhunsaker/openapi2siege

go 1.19

require (
	github.com/juju/persistent-cookiejar v1.0.0
	github.com/pb33f/libopenapi v0.6.3
	github.com/urfave/cli/v2 v2.25.0
	golang.org/x/exp v0.0.0-20230315142452-642cacee5cc0
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/dprotaso/go-yit v0.0.0-20220510233725-9ba8df137936 // indirect
	github.com/juju/go4 v0.0.0-20160222163258-40d72ab9641a // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	gopkg.in/errgo.v1 v1.0.1 // indirect
	gopkg.in/retry.v1 v1.0.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/urfave/cli/v2 => github.com/danhunsaker/urfave-cli/v2 v2.0.0-20230325004445-ee8fbaa01564
