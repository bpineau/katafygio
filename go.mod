module github.com/bpineau/katafygio

go 1.12

require (
	cloud.google.com/go v0.0.0-20160913182117-3b1ae45394a2
	github.com/Azure/go-autorest v9.10.0+incompatible
	github.com/cpuguy83/go-md2man v0.0.0-20180619205630-691ee98543af
	github.com/davecgh/go-spew v1.1.0
	github.com/dgrijalva/jwt-go v0.0.0-20160705203006-01aeca54ebda
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v0.0.0-20170330071051-c0656edd0d9e
	github.com/golang/glog v0.0.0-20141105023935-44145f04b68c
	github.com/golang/protobuf v0.0.0-20171021043952-1643683e1b54
	github.com/google/gofuzz v0.0.0-20161122191042-44d81051d367
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	github.com/gophercloud/gophercloud v0.0.0-20180210024343-6da026c32e2d
	github.com/hashicorp/golang-lru v0.0.0-20160207214719-a0d98a5f2880
	github.com/hashicorp/hcl v0.0.0-20180404174102-ef8a98b0bbce
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/imdario/mergo v0.0.0-20141206190957-6633656539c1
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/json-iterator/go v0.0.0-20171212105241-13f86432b882
	github.com/magiconair/properties v1.8.0
	github.com/mitchellh/mapstructure v0.0.0-20180511142126-bb74f1db0675
	github.com/pelletier/go-toml v1.2.0
	github.com/russross/blackfriday v0.0.0-20180526075726-670777b536d3
	github.com/shurcooL/sanitized_anchor_name v0.0.0-20170918181015-86672fcb3f95
	github.com/sirupsen/logrus v1.0.5
	github.com/spf13/afero v1.1.1
	github.com/spf13/cast v1.2.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/jwalterweatherman v0.0.0-20180109140146-7c0cea34c8ec
	github.com/spf13/pflag v1.0.1
	github.com/spf13/viper v1.0.2
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/crypto v0.0.0-20170825220121-81e90905daef
	golang.org/x/net v0.0.0-20170809000501-1c05540f6879
	golang.org/x/oauth2 v0.0.0-20170412232759-a6bd8cefa181
	golang.org/x/sys v0.0.0-20171031081856-95c657629925
	golang.org/x/text v0.0.0-20170810154203-b19bf474d317
	golang.org/x/time v0.0.0-20161028155119-f51c12702a4d
	google.golang.org/appengine v0.0.0-20160301025000-12d5545dc1cf
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/inf.v0 v0.9.0
	gopkg.in/yaml.v2 v2.0.0-20170721113624-670d4cfef054
	k8s.io/api v0.0.0-20180308224125-73d903622b73
	k8s.io/apimachinery v0.0.0-20180228050457-302974c03f7e
	k8s.io/client-go v7.0.0+incompatible
)
