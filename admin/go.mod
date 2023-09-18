module github.com/MaxBlaushild/poltergeist/admin

go 1.18

replace (
	github.com/MaxBlaushild/poltergeist/pkg/auth => ../pkg/auth
	github.com/MaxBlaushild/poltergeist/pkg/db => ../pkg/db
	github.com/MaxBlaushild/poltergeist/pkg/deep_priest => ../pkg/deep_priest
	github.com/MaxBlaushild/poltergeist/pkg/email => ../pkg/email
	github.com/MaxBlaushild/poltergeist/pkg/encoding => ../pkg/encoding
	github.com/MaxBlaushild/poltergeist/pkg/models => ../pkg/models
	github.com/MaxBlaushild/poltergeist/pkg/slack => ../pkg/slack
	github.com/MaxBlaushild/poltergeist/pkg/texter => ../pkg/texter
	github.com/MaxBlaushild/poltergeist/pkg/twilio => ../pkg/twilio
	github.com/MaxBlaushild/poltergeist/pkg/util => ../pkg/util
)

require (
	github.com/GoAdminGroup/go-admin v1.2.24 // indirect
	github.com/GoAdminGroup/html v0.0.1 // indirect
	github.com/GoAdminGroup/themes v0.0.43 // indirect
	github.com/NebulousLabs/fastrand v0.0.0-20181203155948-6fb6489aac4e // indirect
	github.com/gobuffalo/logger v1.0.6 // indirect
	github.com/gobuffalo/packd v1.0.1 // indirect
	github.com/gobuffalo/packr/v2 v2.8.3 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/markbates/errx v1.1.0 // indirect
	github.com/markbates/oncer v1.0.0 // indirect
	github.com/markbates/safe v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.6 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	xorm.io/builder v0.3.7 // indirect
	xorm.io/xorm v1.0.2 // indirect
)
