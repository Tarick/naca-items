module github.com/Tarick/naca-items

go 1.15

replace github.com/Tarick/naca-items => ./

require (
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/jackc/pgx/v4 v4.9.0
	github.com/nsqio/go-nsq v1.0.8
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	go.uber.org/zap v1.16.0
)
