module github.com/QOSGroup/qos

// go: no requirements found in Gopkg.lock

require (
	github.com/QOSGroup/kepler v0.6.0
	github.com/QOSGroup/qbase v0.2.3-0.20190927065041-32eb90018d34
	github.com/go-kit/kit v0.8.0
	github.com/gorilla/mux v1.7.3
	github.com/magiconair/properties v1.8.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/rakyll/statik v0.1.6

	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.1
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.2
	github.com/tendermint/tm-db v0.1.1
)

replace github.com/QOSGroup/qbase v0.2.3-0.20190927065041-32eb90018d34 => ../qbase
