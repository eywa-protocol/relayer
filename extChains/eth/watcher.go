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
	"time"
)

type ContractWatcher interface {
	Abi() abi.ABI
	Name() string
	Address() common.Address
	Query() [][]interface{}
	NewEvent() interface{}
	SetEventRaw(eventPointer interface{}, log types.Log)
	OnEvent(event interface{})
}

type ClientWatcher interface {
	Subscribe() error
	Resubscribe()
	Close()
}

func NewClientWatcher(ctx context.Context, client *client, contract ContractWatcher) *clientWatcher {
	w := &clientWatcher{
		client:   client,
		cache:    NewLogCache(),
		contract: contract,
		mx:       new(sync.Mutex),
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	return w
}

type clientWatcher struct {
	ctx      context.Context
	cancel   context.CancelFunc
	client   *client
	cache    *LogCache
	contract ContractWatcher
	sub      event.Subscription
	mx       *sync.Mutex
}

func (w *clientWatcher) unpackLog(out interface{}, log types.Log) error {
	if len(log.Data) > 0 {
		if err := w.contract.Abi().UnpackIntoInterface(out, w.contract.Name(), log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range w.contract.Abi().Events[w.contract.Name()].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}

func (w *clientWatcher) watchEvents() (event.Subscription, error) {

	if logChan, sub, err := w.watchLogs(); err != nil {

		return nil, err
	} else {
		return event.NewSubscription(func(quit <-chan struct{}) error {
			defer sub.Unsubscribe()
			for {
				select {
				case log := <-logChan:
					if !w.cache.Exists(log) {
						contractEvent := w.contract.NewEvent()
						if err := w.unpackLog(&contractEvent, log); err != nil {
							return err
						}
						w.contract.SetEventRaw(&contractEvent, log)
						w.client.wg.Add(1)
						go func() {
							defer w.client.wg.Done()
							w.contract.OnEvent(contractEvent)
						}()
						select {

						case err := <-sub.Err():

							return err
						case <-quit:

							return nil
						}
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

func (w *clientWatcher) watchLogs() (chan types.Log, event.Subscription, error) {

	// Append the event selector to the query parameters and construct the topic set
	query := append([][]interface{}{{w.contract.Abi().Events[w.contract.Name()].ID}}, w.contract.Query()...)

	logs := make(chan types.Log, 128)

	if topics, err := abi.MakeTopics(query...); err != nil {

		return nil, nil, err
	} else {

		filterQuery := ethereum.FilterQuery{
			Addresses: []common.Address{w.contract.Address()},
			Topics:    topics,
		}

		if blockNumber := w.cache.BlockNumber(); blockNumber > 0 {

			filterQuery.FromBlock = new(big.Int).SetUint64(blockNumber)
		}

		if sub, err := w.client.SubscribeFilterLogs(w.ctx, filterQuery, logs); err != nil {

			return nil, nil, err
		} else {

			return logs, sub, nil
		}
	}
}

func (w *clientWatcher) Subscribe() error {
	w.mx.Lock()
	defer w.mx.Unlock()
	var err error
	if w.sub, err = w.watchEvents(); err != nil {

		return err
	} else {
		return nil
	}
}

func (w *clientWatcher) Resubscribe() {
	w.mx.Lock()
	defer w.mx.Unlock()

	w.sub = event.Resubscribe(3*time.Second, func(ctx context.Context) (event.Subscription, error) {
		return w.watchEvents()
	})
}

func (w *clientWatcher) Close() {
	w.mx.Lock()
	defer w.mx.Unlock()

	if w.sub != nil {
		w.sub.Unsubscribe()
	}
	w.cancel()
}
