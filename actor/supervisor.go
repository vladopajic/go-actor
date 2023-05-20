package actor

import (
	"fmt"
	"sync"
)

type SupervisedWorker struct {
	Worker
	Inbox  <-chan Message
	Outbox chan<- Message
}

func (sw *SupervisedWorker) SetInbox(inbox <-chan Message) {
	sw.Inbox = inbox
}

func (sw *SupervisedWorker) SetOutbox(outbox chan<- Message) {
	sw.Outbox = outbox
}

type supervisedWorker interface {
	Worker
	GetID() string
	SetInbox(inbox <-chan Message)
	SetOutbox(outbox chan<- Message)
}

type Supervisor struct {
	actors   map[string]Actor
	outboxes map[string]Mailbox[Message]
	inboxes  map[string]Mailbox[Message]
	IDs      []string
	mu       sync.Mutex
}

type Message struct {
	From    string
	To      string
	Type    int
	Content string
}

func NewSupervisor() *Supervisor {
	supervisor := Supervisor{
		actors:   make(map[string]Actor),
		outboxes: make(map[string]Mailbox[Message]),
		inboxes:  make(map[string]Mailbox[Message]),
	}
	return &supervisor
}

func (s *Supervisor) RegisterWorker(worker supervisedWorker) {
	inbox := NewMailbox[Message]()
	outbox := NewMailbox[Message]()
	worker.SetInbox(inbox.ReceiveC())
	worker.SetOutbox(outbox.SendC())
	actor := New(worker)

	s.mu.Lock()
	s.actors[worker.GetID()] = Combine(actor, inbox, outbox)
	s.outboxes[worker.GetID()] = outbox
	s.inboxes[worker.GetID()] = inbox
	s.mu.Unlock()

	s.IDs = append(s.IDs, worker.GetID())
}

func (s *Supervisor) StartAll() {
	for i := range s.IDs {
		id := s.IDs[i]
		actor := s.actors[id]
		go func(id string) {
			// Execute actor's work
			fmt.Printf("Starting actor %s\n", id)
			actor.Start()
			// Listen for inbox
			for message := range s.outboxes[id].ReceiveC() {
				fmt.Printf("Supervisor received message of type %d from %s: %s\n", message.Type, message.From, message.Content)
				s.mu.Lock()
				inbox := s.inboxes[message.To]
				s.mu.Unlock()

				// Send message to worker
				inbox.SendC() <- message
			}
		}(id)
	}
}

func (s *Supervisor) StopAll() {
	for _, id := range s.IDs {
		s.mu.Lock()
		actor := s.actors[id]
		s.mu.Unlock()
		actor.Stop()
	}
}
