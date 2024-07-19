package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	serverutils "github.com/livekit/livekit-server/pkg/utils"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"
	"github.com/livekit/protocol/rpc"
	"github.com/livekit/protocol/utils"
	"github.com/livekit/psrpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	EnabledCacheTTL         = 1 * time.Minute
	RoomAgentTopic          = "room"
	PublisherAgentTopic     = "publisher"
	DefaultHandlerNamespace = ""

	CheckEnabledTimeout = 5 * time.Second
)

type Client interface {
	// LaunchJob starts a room or participant job on an agent.
	// it will launch a job once for each worker in each namespace
	LaunchJob(ctx context.Context, desc *JobRequest)
	Stop() error
}

type JobRequest struct {
	JobType livekit.JobType
	Room    *livekit.Room
	// only set for participant jobs
	Participant *livekit.ParticipantInfo
	Metadata    string
	AgentName   string
}

type agentClient struct {
	client rpc.AgentInternalClient

	mu sync.RWMutex

	// cache response to avoid constantly checking with controllers
	// cache is invalidated with AgentRegistered updates
	roomNamespaces      *serverutils.IncrementalDispatcher[string] // deprecated
	publisherNamespaces *serverutils.IncrementalDispatcher[string] // deprecated
	roomAgentNames      *serverutils.IncrementalDispatcher[string]
	publisherAgentNames *serverutils.IncrementalDispatcher[string]

	enabledExpiresAt time.Time

	workers *workerpool.WorkerPool

	invalidateSub psrpc.Subscription[*emptypb.Empty]
	subDone       chan struct{}
}

func NewAgentClient(bus psrpc.MessageBus) (Client, error) {
	client, err := rpc.NewAgentInternalClient(bus)
	if err != nil {
		return nil, err
	}

	c := &agentClient{
		client:  client,
		workers: workerpool.New(50),
		subDone: make(chan struct{}),
	}

	sub, err := c.client.SubscribeWorkerRegistered(context.Background(), DefaultHandlerNamespace)
	if err != nil {
		return nil, err
	}

	c.invalidateSub = sub

	go func() {
		// invalidate cache
		for range sub.Channel() {
			c.mu.Lock()
			c.roomNamespaces = nil
			c.publisherNamespaces = nil
			c.roomAgentNames = nil
			c.publisherAgentNames = nil
			c.mu.Unlock()
		}

		c.subDone <- struct{}{}
	}()

	return c, nil
}

func (c *agentClient) LaunchJob(ctx context.Context, desc *JobRequest) {
	jobTypeTopic := RoomAgentTopic
	if desc.JobType == livekit.JobType_JT_PUBLISHER {
		jobTypeTopic = PublisherAgentTopic
	}

	dispatcher := c.getDispatcher(desc.AgentName, desc.JobType)

	if dispatcher == nil {
		logger.Infow("not dispatching agent job since no worker is available", "agentName", desc.AgentName, "jobType", desc.JobType)
		return
	}

	dispatcher.ForEach(func(curNs string) {
		topic := GetAgentTopic(desc.AgentName, curNs)
		c.workers.Submit(func() {
			// The cached agent parameters do not provide the exact combination of available job type/agent name/namespace, so some of the JobRequest RPC may not trigger any worker
			_, err := c.client.JobRequest(context.Background(), topic, jobTypeTopic, &livekit.Job{
				Id:          utils.NewGuid(utils.AgentJobPrefix),
				Type:        desc.JobType,
				Room:        desc.Room,
				Participant: desc.Participant,
				Namespace:   curNs,
				AgentName:   desc.AgentName,
				Metadata:    desc.Metadata,
			})
			if err != nil {
				logger.Infow("failed to send job request", "error", err, "namespace", curNs, "jobType", desc.JobType)
			}
		})
	})
}

func (c *agentClient) getDispatcher(agName string, jobType livekit.JobType) *serverutils.IncrementalDispatcher[string] {
	c.mu.Lock()

	if time.Since(c.enabledExpiresAt) > EnabledCacheTTL || c.roomNamespaces == nil ||
		c.publisherNamespaces == nil || c.roomAgentNames == nil || c.publisherAgentNames == nil {
		c.enabledExpiresAt = time.Now()
		c.roomNamespaces = serverutils.NewIncrementalDispatcher[string]()
		c.publisherNamespaces = serverutils.NewIncrementalDispatcher[string]()
		c.roomAgentNames = serverutils.NewIncrementalDispatcher[string]()
		c.publisherAgentNames = serverutils.NewIncrementalDispatcher[string]()

		go c.checkEnabled(c.roomNamespaces, c.publisherNamespaces, c.roomAgentNames, c.publisherAgentNames)
	}

	target := c.roomNamespaces
	agentNames := c.roomAgentNames
	if jobType == livekit.JobType_JT_PUBLISHER {
		target = c.publisherNamespaces
		agentNames = c.publisherAgentNames
	}
	c.mu.Unlock()

	done := make(chan *serverutils.IncrementalDispatcher[string], 1)
	c.workers.Submit(func() {
		agentNames.ForEach(func(ag string) {
			if ag == agName {
				select {
				case done <- target:
				default:
				}
			}
		})
		select {
		case done <- nil:
		default:
		}
	})

	return <-done
}

func (c *agentClient) checkEnabled(roomNamespaces, publisherNamespaces, roomAgentNames, publisherAgentNames *serverutils.IncrementalDispatcher[string]) {
	defer roomNamespaces.Done()
	defer publisherNamespaces.Done()
	defer roomAgentNames.Done()
	defer publisherAgentNames.Done()

	resChan, err := c.client.CheckEnabled(context.Background(), &rpc.CheckEnabledRequest{}, psrpc.WithRequestTimeout(CheckEnabledTimeout))
	if err != nil {
		logger.Errorw("failed to check enabled", err)
		return
	}

	roomNSMap := make(map[string]bool)
	publisherNSMap := make(map[string]bool)
	roomAgMap := make(map[string]bool)
	publisherAgMap := make(map[string]bool)

	for r := range resChan {
		if r.Result.GetRoomEnabled() {
			for _, ns := range r.Result.GetNamespaces() {
				if _, ok := roomNSMap[ns]; !ok {
					roomNamespaces.Add(ns)
					roomNSMap[ns] = true
				}
			}
			for _, ag := range r.Result.GetAgentNames() {
				if _, ok := roomAgMap[ag]; !ok {
					roomAgentNames.Add(ag)
					roomAgMap[ag] = true
				}
			}
		}
		if r.Result.GetPublisherEnabled() {
			for _, ns := range r.Result.GetNamespaces() {
				if _, ok := publisherNSMap[ns]; !ok {
					publisherNamespaces.Add(ns)
					publisherNSMap[ns] = true
				}
			}
			for _, ag := range r.Result.GetAgentNames() {
				if _, ok := publisherAgMap[ag]; !ok {
					publisherAgentNames.Add(ag)
					publisherAgMap[ag] = true
				}
			}
		}
	}
}

func (c *agentClient) Stop() error {
	_ = c.invalidateSub.Close()
	<-c.subDone
	return nil
}

func GetAgentTopic(agentName, namespace string) string {
	if agentName == "" {
		// Backward compatibility
		return namespace
	} else if namespace == "" {
		// Forward compatibility once the namespace field is removed from the worker SDK
		return agentName
	} else {
		return fmt.Sprintf("%s_%s", agentName, namespace)
	}
}
