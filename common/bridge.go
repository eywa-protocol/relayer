package common

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	//"github.com/gorilla/mux"
	"github.com/linkpoolio/bridges"
)

type DBridge struct {
	CLient *ethclient.Client
	Name   string
	Url    string
	Lambda bool
	Func   func(*bridges.Helper) (*Output, error)
}

type Output struct {
	ChainId  string       `json:"chainId"`
	BlockNum string       `json:"blockNum"`
	TxHash   string       `json:"txhash"`
	Data2    bridges.JSON `json:"Data2"`
}

func NewDBridge(ethClient *ethclient.Client, name, url string, f func(*bridges.Helper) (*Output, error)) (a *DBridge, err error) {
	a = &DBridge{}
	a.CLient = ethClient
	a.Name = name
	a.Url = url
	a.Func = f
	return

}

func (ap *DBridge) Opts() *bridges.Opts {
	return &bridges.Opts{
		Name:   ap.Name,
		Lambda: ap.Lambda,
		Path:   "/" + ap.Url,
	}
}

func (ap *DBridge) Run(helper *bridges.Helper) (interface{}, error) {

	res, err := ap.Func(helper)

	if err != nil {
		return nil, err
	}
	logrus.Print(res)
	return res, nil

}
