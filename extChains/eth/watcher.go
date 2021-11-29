package eth

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/sirupsen/logrus"
	"math/big"
	"sync"
)

type ContractWatcher interface {
	Abi() abi.ABI
	Name() string
	Address() common.Address
	Query() [][]interface{}
	NewEventPointer() interface{}
	SetEventRaw(eventPointer interface{}, log types.Log)
	OnEvent(eventPointer interface{}, srcChainId *big.Int)
}

type ClientWatcher interface {
	Subscribe() error
	Resubscribe() error
	IsSubscribed() bool
	Close()
}

func NewClientWatcher(ctx context.Context, client *client, contract ContractWatcher) *clientWatcher {
	w := &clientWatcher{
		client:       client,
		cache:        NewLogCache(),
		contract:     contract,
		subscribedMx: new(sync.Mutex),
		mx:           new(sync.Mutex),
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	return w
}

type clientWatcher struct {
	ctx          context.Context
	cancel       context.CancelFunc
	client       *client
	cache        *LogCache
	contract     ContractWatcher
	sub          event.Subscription
	subscribed   bool
	subscribedMx *sync.Mutex
	mx           *sync.Mutex
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
					logrus.Debugf("chain[%s] %s log recived: %s", w.client.chainId.String(), w.contract.Name(), log.TxHash.Hex())
					if !w.cache.Exists(log) {
						eventPointer := w.contract.NewEventPointer()
						if err := w.unpackLog(eventPointer, log); err != nil {
							return fmt.Errorf("can not unpack chain[%s] %s on error:%w", w.client.chainId.String(), w.contract.Name(), err)
						}
						w.contract.SetEventRaw(eventPointer, log)
						w.client.wg.Add(1)
						go func() {
							defer w.client.wg.Done()
							w.contract.OnEvent(eventPointer, w.client.chainId)
						}()
					} else {
						logrus.Debugf("chain[%s] %s skip already processed log: %s", w.client.chainId.String(), w.contract.Name(), log.TxHash.Hex())
					}
				case err := <-sub.Err():
					w.setSubscribed(false)
					if err != nil {
						logrus.Warnf("chain[%s] subscription to %s stoped on error %v", w.client.chainId.String(), w.contract.Name(), err)
						w.client.SendResubscribeSignal()
					} else {
						logrus.Infof("chain[%s] subscription to %s unsubscribed ", w.client.chainId.String(), w.contract.Name())
					}
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

func (w *clientWatcher) IsSubscribed() bool {
	w.subscribedMx.Lock()
	defer w.subscribedMx.Unlock()
	return w.subscribed
}

func (w *clientWatcher) setSubscribed(subscribed bool) {
	w.subscribedMx.Lock()
	defer w.subscribedMx.Unlock()
	w.subscribed = subscribed

}

func (w *clientWatcher) Subscribe() error {
	w.mx.Lock()
	defer w.mx.Unlock()
	var err error
	if w.sub, err = w.watchEvents(); err != nil {
		w.setSubscribed(false)
		logrus.Warnf("subscribe error: %v", err)
		return err
	} else {
		w.setSubscribed(true)
		logrus.Infof("subscribed to %s", w.contract.Name())
		return nil
	}
}

func (w *clientWatcher) Resubscribe() error {
	w.mx.Lock()
	defer w.mx.Unlock()
	var err error
	if w.sub, err = w.watchEvents(); err != nil {
		w.setSubscribed(false)
		err = fmt.Errorf("can not resubscribe to %s on chain %s on error: %v",
			w.contract.Name(), w.client.chainId.String(), err)
		logrus.Warn(err)
		return err
	} else {
		w.setSubscribed(true)
		logrus.Infof("resubscribed to %s on chain %s", w.contract.Name(), w.client.chainId.String())
		return nil
	}

	// w.sub = event.Resubscribe(5*time.Second, func(ctx context.Context) (event.Subscription, error) {
	// 	if sub, err := w.watchEvents(); err != nil {
	// 		logrus.Warnf("can not resubscribe to %s on chain %s on error: %v",
	// 			w.contract.Name(),w.client.chainId.String(), err)
	//
	// 		return nil, err
	// 	} else {
	// 		logrus.Infof("resubscribed to %s on chain %s", w.contract.Name(),w.client.chainId.String())
	//
	// 		return sub, err
	// 	}
	// })
}

func (w *clientWatcher) Close() {
	w.mx.Lock()
	defer w.mx.Unlock()

	if w.sub != nil {
		w.sub.Unsubscribe()
	}
	w.cancel()
}
