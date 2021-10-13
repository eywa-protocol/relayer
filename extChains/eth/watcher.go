package eth

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
	"sync"
)


type ContractWatcher interface {
	Abi() abi.ABI
	Name() string
	Address() common.Address
	Query() [][]interface{}
	NewEventPointer() interface{}
	SetEventRaw(interface{}, types.Log)
	OnEvent(event interface{}) error
	OnSubscribe(chainId *big.Int)
	OnResubscribe(chainId *big.Int)
}

type ClientWatcher interface {
	Subscribe() error
	Resubscribe() error
	Close()
}

func NewClientWatcher(ctx context.Context, client *client, contract ContractWatcher) *clientWatcher {
	w := &clientWatcher{
		client:   client,
		cache:    NewLogCache(),
		wg:       new(sync.WaitGroup),
		mx:       new(sync.Mutex),
		contract: contract,
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	return w
}

type clientWatcher struct {
	ctx      context.Context
	cancel   context.CancelFunc
	client   *client
	cache    *LogCache
	wg       *sync.WaitGroup
	mx       *sync.Mutex
	contract ContractWatcher
	sub      event.Subscription
}

func (cw *clientWatcher) unpackLog(out interface{}, log types.Log) error {
	if len(log.Data) > 0 {
		if err := cw.contract.Abi().UnpackIntoInterface(out, cw.contract.Name(), log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range cw.contract.Abi().Events[cw.contract.Name()].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}

func (cw *clientWatcher) watchEvents(ch chan<- interface{}) (event.Subscription, error) {

	if logChan, sub, err := cw.watchLogs(); err != nil {

		return nil, err
	} else {
		return event.NewSubscription(func(quit <-chan struct{}) error {
			defer sub.Unsubscribe()
			for {
				select {
				case log := <-logChan:
					eventPointer := cw.contract.NewEventPointer()
					if err := cw.unpackLog(eventPointer, log); err != nil {
						return err
					}
					cw.contract.SetEventRaw(eventPointer, log)
					select {
					case ch <- eventPointer:
					case err := <-sub.Err():

						return err
					case <-quit:

						return nil
					}

				case err := <-sub.Err():

					return err
				case <-quit:

					return nil
				}
			}
		}), nil
	}
}

func (cw *clientWatcher) watchLogs() (chan types.Log, event.Subscription, error) {

	// Append the event selector to the query parameters and construct the topic set
	query := append([][]interface{}{{cw.contract.Abi().Events[cw.contract.Name()].ID}}, cw.contract.Query()...)

	logs := make(chan types.Log, 128)

	if topics, err := abi.MakeTopics(query...); err != nil {

		return nil, nil, err
	} else {

		filterQuery := ethereum.FilterQuery{
			Addresses: []common.Address{cw.contract.Address()},
			Topics:    topics,
		}

		if blockNumber := cw.cache.BlockNumber(); blockNumber > 0 {

			filterQuery.FromBlock = new(big.Int).SetUint64(blockNumber)
		}

		if sub, err := cw.client.SubscribeFilterLogs(cw.ctx, filterQuery, logs); err != nil {

			return nil, nil, err
		} else {

			return logs, sub, nil
		}
	}
}

func (cw *clientWatcher) Subscribe(ch chan<- interface{}) error {
    var err
    if cw.sub,err = cw.watchEvents(ch);err!= nil {

    	return err
    }else {
    	return err
	}
}

func (cw *clientWatcher) Resubscribe() error {
	panic("implement me")
}

func (cw *clientWatcher) Close() {
	cw.cancel()
}
