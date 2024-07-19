package routing

import (
	"runtime"
	"time"

	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/utils"
	"github.com/livekit/protocol/utils/guid"

	"github.com/livekit/livekit-server/pkg/config"
)

type LocalNode *livekit.Node

func NewLocalNode(conf *config.Config) (LocalNode, error) {
	nodeID := guid.New(utils.NodePrefix)
	if conf.RTC.NodeIP == "" {
		return nil, ErrIPNotSet
	}
	node := &livekit.Node{
		Id:      nodeID,
		Ip:      conf.RTC.NodeIP,
		NumCpus: uint32(runtime.NumCPU()),
		Region:  conf.Region,
		State:   livekit.NodeState_SERVING,
		Stats: &livekit.NodeStats{
			StartedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		},
	}

	return node, nil
}
