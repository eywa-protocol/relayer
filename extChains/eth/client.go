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
	defaultDialTimeout = 5 * time.Second
	defaultCallTimeout = 3 * time.Second
)

type Config struct {
	CallTimeout time.Duration
	DialTimeout time.Duration
	Id          uint
	Urls        []string
}

type Client interface {
	bind.ContractBackend
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
	WaitForBlockCompletion(txHash common.Hash) (int, *types.Receipt)
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
		watchers:    make(map[string]ClientWatcher),
		chainId:     big.NewInt(int64(cfg.Id)),
		callTimeout: defaultCallTimeout,
		dialTimeout: defaultDialTimeout,
		mx:          new(sync.Mutex),
		wg:          new(sync.WaitGroup),
		subLoopMx:   new(sync.Mutex),
		currentUrl:  "",
		urls:        cfg.Urls,
	}

	if cfg.CallTimeout > 0 {
		c.callTimeout = cfg.CallTimeout
	}

	if cfg.DialTimeout > 0 {
		c.dialTimeout = cfg.DialTimeout
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
		for {
			select {
			case <-c.reSubWatchers:
				c.mx.Lock()
				if c.connected && c.recreated {
					c.mx.Unlock()
					for _, watcher := range c.watchers {
						watcher.Resubscribe()
					}
				} else {
					c.mx.Unlock()
				}
			case <-c.ctx.Done():
				break
			}
		}
	}()
}

func (c *client) AddWatcher(contract ContractWatcher) {
	c.mx.Lock()
	key := c.watcherKey(contract)
	if _, exists := c.watchers[key]; !exists {
		watcher := NewClientWatcher(c.ctx, c, contract)
		c.watchers[key] = watcher
		c.mx.Unlock()
		if err := watcher.Subscribe(); err != nil {

			logrus.Error(fmt.Errorf("watcher subscribe error: %w", err))
		}
		c.subLoop()
	} else {
		c.mx.Unlock()
	}
}

func (c *client) RemoveWatcher(contract ContractWatcher) {
	c.mx.Lock()
	defer c.mx.Unlock()
	key := c.watcherKey(contract)
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
	c.recreated = false
	if c.ethClient == nil {
		for _, url := range c.urls {
			if c.currentUrl != "" && len(c.urls) > 1 && url == c.currentUrl {
				continue
			} else if client, err = c.dial(c.ctx, url); err != nil {
				logrus.WithFields(logrus.Fields{
					field.CainId: c.chainId,
					field.EthUrl: url,
				}).Error(fmt.Errorf("can not connect to chain rpc on error: %w", err))
				continue
			} else if chainId, err := c.getChainId(client, c.ctx); err != nil {
				logrus.WithFields(logrus.Fields{
					field.CainId: c.chainId,
				}).Error(fmt.Errorf("get network chainID [%s] error: %w", c.chainId.String(), err))
				continue
			} else if chainId.Cmp(c.chainId) != 0 {
				logrus.WithFields(logrus.Fields{
					field.CainId:         c.chainId,
					field.NetworkChainId: chainId,
					field.EthUrl:         url,
				}).Error(fmt.Errorf("client chainID [%s] not match to network: %w", c.chainId.String(), err))
				continue
			} else {
				c.ethClient = client
				c.connected = true
				c.recreated = true
				c.subLoopMx.Lock()
				if c.subLoopInitialized {
					c.reSubWatchers <- struct{}{}
				}
				c.subLoopMx.Unlock()
				c.mx.Unlock()

				return c.ethClient, nil
			}
		}
		c.connected = false

		c.mx.Unlock()
		return nil, fmt.Errorf("connection to all rpc url for chain [%s]  failed", c.chainId.String())
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

			c.mx.Unlock()
			return c.ethClient, nil
		}
	}
}

func (c *client) getChainId(client *ethclient.Client, ctx context.Context) (*big.Int, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.callTimeout)
	defer cancel()
	return client.ChainID(callCtx)
}