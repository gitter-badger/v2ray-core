package core

import (
	"github.com/v2ray/v2ray-core/common/log"
	v2net "github.com/v2ray/v2ray-core/common/net"
)

var (
	inboundFactories  = make(map[string]InboundConnectionHandlerFactory)
	outboundFactories = make(map[string]OutboundConnectionHandlerFactory)
)

func RegisterInboundConnectionHandlerFactory(name string, factory InboundConnectionHandlerFactory) error {
	// TODO check name
	inboundFactories[name] = factory
	return nil
}

func RegisterOutboundConnectionHandlerFactory(name string, factory OutboundConnectionHandlerFactory) error {
	// TODO check name
	outboundFactories[name] = factory
	return nil
}

// Point is an single server in V2Ray system.
type Point struct {
	port       uint16
	ichFactory InboundConnectionHandlerFactory
	ichConfig  []byte
	ochFactory OutboundConnectionHandlerFactory
	ochConfig  []byte
}

// NewPoint returns a new Point server based on given configuration.
// The server is not started at this point.
func NewPoint(config PointConfig) (*Point, error) {
	var vpoint = new(Point)
	vpoint.port = config.Port()

	ichFactory, ok := inboundFactories[config.InboundConfig().Protocol()]
	if !ok {
		panic(log.Error("Unknown inbound connection handler factory %s", config.InboundConfig().Protocol()))
	}
	vpoint.ichFactory = ichFactory
	vpoint.ichConfig = config.InboundConfig().Content()

	ochFactory, ok := outboundFactories[config.OutboundConfig().Protocol()]
	if !ok {
		panic(log.Error("Unknown outbound connection handler factory %s", config.OutboundConfig().Protocol))
	}

	vpoint.ochFactory = ochFactory
	vpoint.ochConfig = config.OutboundConfig().Content()

	return vpoint, nil
}

type InboundConnectionHandlerFactory interface {
	Create(vp *Point, config []byte) (InboundConnectionHandler, error)
}

type InboundConnectionHandler interface {
	Listen(port uint16) error
}

type OutboundConnectionHandlerFactory interface {
	Initialize(config []byte) error
	Create(VP *Point, firstPacket v2net.Packet) (OutboundConnectionHandler, error)
}

type OutboundConnectionHandler interface {
	Start(ray OutboundRay) error
}

// Start starts the Point server, and return any error during the process.
// In the case of any errors, the state of the server is unpredicatable.
func (vp *Point) Start() error {
	if vp.port <= 0 {
		return log.Error("Invalid port %d", vp.port)
	}

	vp.ochFactory.Initialize(vp.ochConfig)

	inboundConnectionHandler, err := vp.ichFactory.Create(vp, vp.ichConfig)
	if err != nil {
		return err
	}
	err = inboundConnectionHandler.Listen(vp.port)
	return nil
}

func (p *Point) DispatchToOutbound(packet v2net.Packet) InboundRay {
	ray := NewRay()
	// TODO: handle error
	och, _ := p.ochFactory.Create(p, packet)
	_ = och.Start(ray)
	return ray
}

func (p *Point) DispatchToInbound(packet v2net.Packet) {
	return
}
