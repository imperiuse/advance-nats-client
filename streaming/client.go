package streaming

import (
	"context"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/imperiuse/advance-nats-client/logger"
	nc "github.com/imperiuse/advance-nats-client/nats"
	"github.com/imperiuse/advance-nats-client/serializable"
)

type AdvanceNatsClient interface {
	// NATS @see -> nc.SimpleNatsClientI
	Ping(context.Context, nc.Subj) (bool, error)
	PongHandler(nc.Subj) (*nc.Subscription, error)
	PongQueueHandler(nc.Subj, nc.QueueGroup) (*nc.Subscription, error)
	Request(context.Context, Subj, Serializable, Serializable) error
	ReplyHandler(Subj, Serializable, nc.Handler) (*nc.Subscription, error)
	// NATS Streaming
	PublishSync(Subj, Serializable) error
	PublishAsync(Subj, Serializable, AckHandler) (GUID, error)
	DefaultAckHandler() AckHandler
	Subscribe(Subj, Serializable, Handler, ...SubscriptionOption) (Subscription, error)
	QueueSubscribe(Subj, Qgroup, Serializable, Handler, ...SubscriptionOption) (Subscription, error)
	// General for both NATS and NATS Streaming
	UseCustomLogger(logger.Logger)
	NatsConn() *nats.Conn
	Nats() nc.SimpleNatsClientI
	Close() error
}

type (
	client struct {
		log logger.Logger
		sc  PureNatsStunConnI    // StunConnI equals stan.Conn
		nc  nc.SimpleNatsClientI // Simple Nats client (from another package of this library =) )
	}

	URL = string

	Option = stan.Option
)

//go:generate mockery --name=PureNatsStunConnI
type (
	// StunConnI represents a connection to the NATS Streaming subsystem. It can Publish and
	// Subscribe to messages within the NATS Streaming cluster.
	// The connection is safe to use in multiple Go routines concurrently.
	PureNatsStunConnI interface {
		Publish(Subj, []byte) error
		PublishAsync(Subj, []byte, AckHandler) (string, error)
		Subscribe(Subj, MsgHandler, ...SubscriptionOption) (Subscription, error)
		QueueSubscribe(Subj, Qgroup, MsgHandler, ...SubscriptionOption) (Subscription, error)
		Close() error
	}

	Handler = func(*stan.Msg, Serializable)

	Subj   = string
	Qgroup = string

	Msg                = stan.Msg
	MsgHandler         = stan.MsgHandler // func(msg *Msg)
	Subscription       = stan.Subscription
	SubscriptionOption = stan.SubscriptionOption
	AckHandler         = stan.AckHandler // func(string, error)

	GUID = string // id send msg from Nats Streaming

	Serializable = serializable.Serializable
)

const EmptyGUID = ""

var (
	DefaultDSN       = []URL{nats.DefaultURL}
	DefaultClusterID = "test-cluster"
	testDSN          = []URL{"nats://127.0.0.1:4223"}
	EmptyHandler     = func(*Msg, Serializable) {}
)

var (
	ErrEmptyNatsConn   = errors.New("empty pure nats connection from advance nats client")
	ErrEmptyNatsClient = errors.New("empty advance nats client")
)

// New - Create Nats Streaming client with instance of Advance Nats client
func New(clusterID string, clientID string, nc nc.SimpleNatsClientI, options ...Option) (*client, error) {
	if nc != nil {
		if nc.NatsConn() == nil {
			return nil, ErrEmptyNatsConn
		}
		options = append(options, stan.NatsConn(nc.NatsConn()))
	}

	c, err := NewOnlyStreaming(clusterID, clientID, nil, options...)
	if err != nil || c == nil {
		return nil, err
	}

	c.nc = nc

	return c, nil
}

func NewOnlyStreaming(clusterID string, clientID string, dsn []URL, options ...Option) (*client, error) {
	c := NewDefaultClient()

	// Default settings for internal NATS client
	if len(options) == 0 {
		options = c.defaultNatsStreamingOptions()
	}

	// DSN for NATS connection, e.g. "nats://127.0.0.1:4222" stan.DefaultNatsURL
	if dsn != nil {
		options = append(options, stan.NatsURL(strings.Join(dsn, ", ")))
	}

	sc, err := stan.Connect(clusterID, clientID, options...)
	if err != nil {
		return nil, fmt.Errorf("can't crate nats-streaming conn. %w", err)
	}

	c.sc = sc

	return c, nil
}

func NewDefaultClient() *client {
	return &client{
		log: logger.Log,
	}
}

func (c *client) defaultNatsStreamingOptions() []Option {
	return []Option{
		stan.Pings(stan.DefaultPingInterval, stan.DefaultPingMaxOut), // todo, maybe should increase, very hard
		stan.ConnectWait(stan.DefaultConnectWait),
		stan.MaxPubAcksInflight(stan.DefaultMaxPubAcksInflight),
		stan.PubAckWait(stan.DefaultAckWait),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			c.log.Error("Connection lost, reason: %v", zap.Error(reason))
		}),
	}
}

// Wrapper for Nats Simple Client

// Ping - under the hood wrapper for nc.Ping
func (c *client) Ping(ctx context.Context, subj nc.Subj) (bool, error) {
	if c.nc == nil {
		return false, ErrEmptyNatsClient
	}
	return c.nc.Ping(ctx, subj)
}

// PongHandler - under the hood wrapper for nc.PongHandler
func (c *client) PongHandler(subj nc.Subj) (*nc.Subscription, error) {
	if c.nc == nil {
		return nil, ErrEmptyNatsClient
	}
	return c.nc.PongHandler(subj)
}

// PongQueueHandler - under the hood wrapper for nc.PongQueueHandler
func (c *client) PongQueueHandler(subj nc.Subj, qgroup nc.QueueGroup) (*nc.Subscription, error) {
	if c.nc == nil {
		return nil, ErrEmptyNatsClient
	}
	return c.nc.PongQueueHandler(subj, qgroup)
}

// Request under the hood used simple NATS connect and simple Request - Reply semantic with at most once guarantee (! not at least !)
func (c *client) Request(ctx context.Context, subj Subj, requestData Serializable, replyData Serializable) error {
	if c.nc == nil {
		return ErrEmptyNatsClient
	}
	return c.nc.Request(ctx, subj, requestData, replyData)
}

// ReplyHandler under the hood used simple Advance NATS client, Reply semantic with at most once (! not at least once guarantee !)
func (c *client) ReplyHandler(subj Subj, awaitData Serializable, msgHandler nc.Handler) (*nc.Subscription, error) {
	if c.nc == nil {
		return nil, ErrEmptyNatsClient
	}
	return c.nc.ReplyHandler(subj, awaitData, msgHandler)
}

func (c *client) UseCustomLogger(log logger.Logger) {
	c.log = log
	if c.nc != nil {
		c.nc.UseCustomLogger(c.log)
	}
}

// PublishSync will publish to the NATS Streaming cluster and wait for an ACK.
func (c *client) PublishSync(subj Subj, data Serializable) error {
	c.log.Debug("[PublishSync]",
		zap.String("subj", subj),
	)
	b, err := data.Marshal()
	if err != nil {
		c.log.Error("[PublishSync] Marshall",
			zap.String("subj", subj),
			zap.Error(err),
		)
		return err
	}
	return c.sc.Publish(subj, b)
}

// PublishAsync will publish to the cluster and asynchronously process
// the ACK or error state. It will return the GUID for the message being sent.
func (c *client) PublishAsync(subj Subj, data Serializable, ah AckHandler) (GUID, error) {
	c.log.Debug("[PublishAsync]",
		zap.String("subj", subj),
		zap.Any("data", data),
	)
	b, err := data.Marshal()
	if err != nil {
		c.log.Error("[PublishAsync] Marshall",
			zap.String("subj", subj),
			zap.Error(err))
		return EmptyGUID, err
	}

	if ah == nil {
		c.log.Debug("[PublishAsync] AckHandler does not set. Use DefaultAckHandler",
			zap.String("subj", subj),
		)
		ah = c.DefaultAckHandler()
	}

	return c.sc.PublishAsync(subj, b, ah)
}

// DefaultAckHandler - return default ack func with logging problem's, !please better use own ack handler func!
func (c *client) DefaultAckHandler() AckHandler {
	return func(ackUID string, err error) {
		if err != nil {
			c.log.Error("Warning: error publishing msg", zap.Error(err), zap.String("msg_id", ackUID))
		} else {
			c.log.Debug("Received ack for msg", zap.String("msg_id", ackUID))
		}
	}
}

// Subscribe - func for subscribe on any subject, if no options received - default for all options append stan.SetManualAckMode()
func (c *client) Subscribe(subj Subj, awaitData Serializable, handler Handler, opt ...SubscriptionOption) (Subscription, error) {
	opt = append(opt, stan.SetManualAckMode()) // NB! stan.StartWithLastReceived()) - начинает с последнего уже доставленного! с виду кажется дубль
	msgHandler := func(msg *Msg) {
		if err := msg.Ack(); err != nil { // manual fast ack
			c.log.Error("[Subscribe] msg.Ack() problem",
				zap.Any("msg", msg),
				zap.String("subj", subj),
				zap.Error(err))
			return
		}

		if msg.Redelivered {
			c.log.Warn("[Subscribe] Redelivered msg received",
				zap.Any("msg", msg),
				zap.String("subj", subj))
			//return // TODO. THINK HERE. WHAT WE NEED TO DO?
		}

		if msg == nil {
			c.log.Warn("[Subscribe] Msg is nil",
				zap.Any("msg", msg),
				zap.String("subj", subj))
			return
		}

		if err := awaitData.Unmarshal(msg.Data); err != nil {
			c.log.Error("[Subscribe] Unmarshal",
				zap.Error(err),
				zap.Any("msg", msg),
				zap.String("subj", subj))
			return
		}

		handler(msg, awaitData)
	}
	c.log.Debug("[Subscribe]", zap.String("subj", subj))
	return c.sc.Subscribe(subj, msgHandler, opt...)
}

// QueueSubscribe will perform a queue subscription with the given options to the cluster.
// If no option is specified, DefaultSubscriptionOptions are used. The default start
// position is to receive new messages only (messages published after the subscription is
// registered in the cluster).
func (c *client) QueueSubscribe(subj Subj, qgroup Qgroup, awaitData Serializable, handler Handler, opt ...SubscriptionOption) (Subscription, error) {
	opt = append(opt, stan.SetManualAckMode()) // NB! stan.StartWithLastReceived()) - начинает с последнего уже доставленного! с виду кажется дубль
	msgHandler := func(msg *Msg) {
		if err := msg.Ack(); err != nil { // manual fast ack
			c.log.Error("[QueueSubscribe] msg.Ack() problem",
				zap.String("subj", subj),
				zap.String("qgroup", qgroup),
				zap.Any("msg", msg),
				zap.Error(err))
			return
		}

		if msg.Redelivered {
			c.log.Warn("[QueueSubscribe] Redelivered msg received",
				zap.Any("msg", msg),
				zap.String("subj", subj),
				zap.String("qgroup", qgroup))
			//return // TODO. THINK HERE. WHAT WE NEED TO DO?
		}

		if msg == nil {
			c.log.Warn("[QueueSubscribe] Msg is nil",
				zap.String("subj", subj),
				zap.String("qgroup", qgroup))
			return
		}

		if err := awaitData.Unmarshal(msg.Data); err != nil {
			c.log.Error("[QueueSubscribe] Unmarshal",
				zap.String("subj", subj),
				zap.String("qgroup", qgroup),
				zap.Any("msg", msg),
				zap.Error(err))
			return
		}

		handler(msg, awaitData)
	}

	c.log.Debug("[QueueSubscribe]", zap.String("subj", subj), zap.String("qgroup", qgroup))
	return c.sc.QueueSubscribe(subj, qgroup, msgHandler, opt...)
}

// Nats - return advance Nats client
func (c *client) Nats() nc.SimpleNatsClientI {
	return c.nc
}

// NatsConn - return pure nats conn pointer
func (c *client) NatsConn() *nats.Conn {
	if c.nc == nil {
		return nil
	}

	return c.nc.NatsConn()
}

// Close - close Nats streaming connection and NB! Also Close pure Nats Connection
func (c *client) Close() error {
	var err error
	// Note that you will be responsible for closing the NATS Connection after the streaming connection has been closed.
	defer func() {
		if c.nc != nil {
			err1 := c.nc.Close()
			if err == nil {
				err = err1
			} else {
				err = errors.Wrap(err, fmt.Sprint(err1))
			}
		}
	}()
	err = c.sc.Close()
	return err
}