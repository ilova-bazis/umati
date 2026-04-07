package schema

type Status string
type Priority string
type Actor string
type EventType string

const (
	StatusDraft      Status = "draft"
	StatusPaused     Status = "paused"
	StatusReady      Status = "ready"
	StatusClaimed    Status = "claimed"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
	StatusDeleted    Status = "deleted"
)

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

const (
	ActorOpenCode Actor = "opencode"
	ActorCodex    Actor = "codex"
	ActorClaude   Actor = "claude"
	ActorHuman    Actor = "human"
)

const (
	EventCreated   EventType = "created"
	EventUpdated   EventType = "updated"
	EventClaimed   EventType = "claimed"
	EventReleased  EventType = "released"
	EventStarted   EventType = "started"
	EventPaused    EventType = "paused"
	EventCompleted EventType = "completed"
	EventCancelled EventType = "cancelled"
	EventDeleted   EventType = "deleted"
)

var activeStatuses = map[Status]struct{}{
	StatusDraft:      {},
	StatusPaused:     {},
	StatusReady:      {},
	StatusClaimed:    {},
	StatusInProgress: {},
	StatusDone:       {},
	StatusCancelled:  {},
}

var priorities = map[Priority]struct{}{
	PriorityLow:    {},
	PriorityMedium: {},
	PriorityHigh:   {},
	PriorityUrgent: {},
}

var actors = map[Actor]struct{}{
	ActorOpenCode: {},
	ActorCodex:    {},
	ActorClaude:   {},
	ActorHuman:    {},
}

var eventTypes = map[EventType]struct{}{
	EventCreated:   {},
	EventUpdated:   {},
	EventClaimed:   {},
	EventReleased:  {},
	EventStarted:   {},
	EventPaused:    {},
	EventCompleted: {},
	EventCancelled: {},
	EventDeleted:   {},
}

func IsValidActiveStatus(status Status) bool {
	_, ok := activeStatuses[status]
	return ok
}

func IsValidPriority(priority Priority) bool {
	_, ok := priorities[priority]
	return ok
}

func IsValidActor(actor Actor) bool {
	_, ok := actors[actor]
	return ok
}

func IsValidEventType(eventType EventType) bool {
	_, ok := eventTypes[eventType]
	return ok
}
