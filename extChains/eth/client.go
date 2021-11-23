package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/sirupsen/logrus"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/sentry/field"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	defaultDialTimeout  = 10 * time.Second
	defaultCallTimeout  = 30 * time.Second
	defaultBlockTimeout = 120 * time.Second
)

type Config struct {
	CallTimeout  time.Duration
	DialTimeout  time.Duration
	BlockTimeout time.Duration
	Id           uint64
	Urls         []string
}

func (c *Config) SetDefault() {
	c.CallTimeout = defaultCallTimeout
	c.DialTimeout = defaultDialTimeout
	c.BlockTimeout = defaultBlockTimeout
}

type Client interface {
	bind.ContractBackend
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	WaitForBlockCompletion(txHash common.Hash) (int, *types.Receipt)
	WaitTransaction(txHash common.Hash) (*types.Receipt, error)
	CallOpt(privateKey *ecdsa.PrivateKey) (*bind.TransactOpts, error)
	AddWatcher(contract ContractWatcher)
	RemoveWatcher(contract ContractWatcher)
	Close()
}

type client struct {
	ctx                context.Context
	cancel             context.CancelFunc
	ethClient          *ethclient.Client
	watchers           map[string]ClientWatcher
	chainId            *big.Int
	callTimeout        time.Duration
	dialTimeout        time.Duration
	blockTimeout       time.Duration
	mx                 *sync.Mutex
	wg                 *sync.WaitGroup
	connected          bool
	recreated          bool
	subLoopInitialized bool
	subLoopMx          *sync.Mutex
	reSubWatchers      chan struct{}
	currentUrl         string
	urls               []string
}

func NewClient(ctx context.Context, cfg *Config) (*client, error) {
	if len(cfg.Urls) <= 0 {
		return nil, ErrClientUrlsEmpty
	}
	c := &client{
		watchers:     make(map[string]ClientWatcher),
		chainId:      new(big.Int).SetUint64(cfg.Id),
		callTimeout:  defaultCallTimeout,
		dialTimeout:  defaultDialTimeout,
		blockTimeout: defaultBlockTimeout,
		mx:           new(sync.Mutex),
		wg:           new(sync.WaitGroup),
		subLoopMx:    new(sync.Mutex),
		currentUrl:   "",
		urls:         cfg.Urls,
	}

	if cfg.CallTimeout > 0 {
		c.callTimeout = cfg.CallTimeout
	}

	if cfg.DialTimeout > 0 {
		c.dialTimeout = cfg.DialTimeout
	}

	if cfg.BlockTimeout > 0 {
		c.blockTimeout = cfg.BlockTimeout
	}
	c.ctx, c.cancel = context.WithCancel(ctx)

	return c, nil
}

func (c *client) watcherKey(contract ContractWatcher) string {
	var b strings.Builder
	b.WriteString(contract.Address().Hex())
	b.WriteString(":")
	b.WriteString(contract.Name())
	return b.String()
}

func (c *client) subLoop() {
	c.subLoopMx.Lock()
	if c.subLoopInitialized {
		c.subLoopMx.Unlock()
		return
	} else {
		c.reSubWatchers = make(chan struct{}, 1)
		c.subLoopInitialized = true
		c.subLoopMx.Unlock()
	}
	c.wg.Add(1)
	go func() {
		logrus.Infof("start subscription loop on chain %s", c.chainId.String())
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		var reconnectOnTimer bool
		rMx := new(sync.Mutex)
		for {
			select {
			case <-c.reSubWatchers:
				logrus.Infof("resubscribe watchers connected: %v, recreated: %v", c.connected, c.recreated)
				c.mx.Lock()
				if c.connected && c.recreated {
					logrus.Infof("set reconnect on timer false %s", c.chainId.String())
					c.mx.Unlock()
					rMx.Lock()
					reconnectOnTimer = false
					rMx.Unlock()
					for _, watcher := range c.getAllWatchers() {
						if err := watcher.Resubscribe(); err != nil {
							logrus.Infof("set reconnect on timer true %s", c.chainId.String())
							reconnectOnTimer = true
							break
						}
					}
				} else if c.connected && !c.recreated {
					logrus.Infof("set reconnect on timer false %s", c.chainId.String())
					c.mx.Unlock()
					rMx.Lock()
					reconnectOnTimer = false
					rMx.Unlock()
					for _, watcher := range c.getUnsubscribedWatchers() {
						if err := watcher.Resubscribe(); err != nil {
							reconnectOnTimer = true
							logrus.Infof("set reconnect on timer true %s", c.chainId.String())
							break
						}
					}
				} else if !c.connected {
					logrus.Infof("set reconnect on timer true %s", c.chainId.String())
					c.mx.Unlock()
					rMx.Lock()
					reconnectOnTimer = true
					rMx.Unlock()
				} else {
					c.mx.Unlock()
				}
			case <-c.ctx.Done():
				logrus.Infof("stop subLoop on context canceled %s", c.chainId.String())
				break
			case <-t.C:
				rMx.Lock()
				if reconnectOnTimer {
					logrus.Infof("reconnect on timer %s", c.chainId.String())
					rMx.Unlock()
					if _, err := c.getClient(); err == nil {
						rMx.Lock()
						reconnectOnTimer = false
						rMx.Unlock()
					}
				} else {
					rMx.Unlock()
				}
			default:
			}
		}
	}()
}

func (c *client) getAllWatchers() []ClientWatcher {
	c.mx.Lock()
	defer c.mx.Unlock()
	watchers := make([]ClientWatcher, 0, len(c.watchers))
	for _, watcher := range c.watchers {
		watchers = append(watchers, watcher)
	}
	return watchers
}

func (c *client) getUnsubscribedWatchers() []ClientWatcher {
	c.mx.Lock()
	defer c.mx.Unlock()
	watchers := make([]ClientWatcher, 0, len(c.watchers))
	for _, watcher := range c.watchers {
		if !watcher.IsSubscribed() {
			watchers = append(watchers, watcher)
		}
	}
	return watchers
}

func (c *client) AddWatcher(contract ContractWatcher) {
	logrus.Infof("add watcher for %s on chain %s", contract.Name(), c.chainId.String())
	c.mx.Lock()
	key := c.watcherKey(contract)
	if _, exists := c.watchers[key]; !exists {
		watcher := NewClientWatcher(c.ctx, c, contract)
		c.watchers[key] = watcher
		c.mx.Unlock()
		c.subLoop()
		if err := watcher.Subscribe(); err != nil {
			logrus.Error(fmt.Errorf("watcher subscribe error: %w", err))
		}
	} else {
		logrus.Infof("added watcher for %s on chain %s", contract.Name(), c.chainId.String())
		c.mx.Unlock()
	}
}

func (c *client) RemoveWatcher(contract ContractWatcher) {
	logrus.Infof("remove watcher for %s on chain %s", contract.Name(), c.chainId.String())
	key := c.watcherKey(contract)
	c.mx.Lock()
	defer c.mx.Unlock()
	if watcher, exists := c.watchers[key]; exists {
		watcher.Close()
		delete(c.watchers, key)
	}
}

func (c *client) closeWatchers() {
	for key, watcher := range c.watchers {
		watcher.Close()
		delete(c.watchers, key)
	}
}

func (c *client) Close() {
	c.mx.Lock()
	c.closeWatchers()
	c.mx.Unlock()
	c.cancel()
	c.wg.Wait()
	c.subLoopMx.Lock()
	defer c.subLoopMx.Unlock()
	if c.subLoopInitialized {
		close(c.reSubWatchers)
	}
}

func (c *client) dial(ctx context.Context, url string) (*ethclient.Client, error) {
	dialCtx, cancel := context.WithTimeout(ctx, c.dialTimeout)
	defer cancel()
	return ethclient.DialContext(dialCtx, url)
}

func (c *client) getClient() (client *ethclient.Client, err error) {
	c.mx.Lock()
	if c.ethClient == nil {
		for _, url := range c.urls {
			c.connected = false
			if c.currentUrl != "" && len(c.urls) > 1 && url == c.currentUrl {
				continue
			} else if client, err = c.dial(c.ctx, url); err != nil {
				err = fmt.Errorf("can not connect to chain rpc on error: %w", err)
				logrus.WithFields(logrus.Fields{
					field.CainId: c.chainId,
					field.EthUrl: url,
				}).Error(err)
				continue
			} else if chainId, err := c.getChainId(client, c.ctx); err != nil {
				err = fmt.Errorf("get network chainID [%s] error: %w", c.chainId.String(), err)
				logrus.WithFields(logrus.Fields{
					field.CainId: c.chainId,
				}).Error(err)
				continue
			} else if chainId.Cmp(c.chainId) != 0 {
				err = fmt.Errorf("client chainID [%s] not match to network: %w", c.chainId.String(), err)
				logrus.WithFields(logrus.Fields{
					field.CainId:         c.chainId,
					field.NetworkChainId: chainId,
					field.EthUrl:         url,
				}).Error(err)
				continue
			} else {
				c.ethClient = client
				c.connected = true
				c.recreated = true
				break
			}
		}

		if c.connected {
			defer c.mx.Unlock()
			c.SendResubscribeSignal()
			return c.ethClient, nil
		} else {
			defer c.mx.Unlock()

			return nil, fmt.Errorf("connection to all rpc url for chain [%s]  failed", c.chainId.String())
		}
	} else {
		if chainId, err := c.getChainId(c.ethClient, c.ctx); err != nil {
			logrus.WithFields(logrus.Fields{
				field.CainId: c.chainId,
			}).Error(fmt.Errorf("get network chainID [%s] error: %w", c.chainId.String(), err))
			c.ethClient = nil
			c.connected = false

			c.mx.Unlock()
			return c.getClient()
		} else if chainId.Cmp(c.chainId) != 0 {
			logrus.WithFields(logrus.Fields{
				field.CainId:         c.chainId,
				field.NetworkChainId: chainId,
				field.EthUrl:         c.currentUrl,
			}).Error(fmt.Errorf("client chainID [%s] not match to network: %w", c.chainId.String(), err))
			c.ethClient = nil
			c.connected = false
			c.mx.Unlock()

			return c.getClient()
		} else {
			c.connected = true
			c.recreated = false
			c.mx.Unlock()

			return c.ethClient, nil
		}
	}
}

func (c *client) SendResubscribeSignal() {
	c.subLoopMx.Lock()
	if c.subLoopInitialized {
		c.reSubWatchers <- struct{}{}
	}
	c.subLoopMx.Unlock()
}

func (c *client) getChainId(client *ethclient.Client, ctx context.Context) (*big.Int, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	return client.ChainID(callCtx)
}
