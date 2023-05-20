package actor

import (
	"fmt"
	"testing"
	"time"
)

type testConsumeWorker struct {
	SupervisedWorker
	ID string
}

func (tw *testConsumeWorker) GetID() string {
	return tw.ID
}

func (tw *testConsumeWorker) DoWork(c Context) WorkerStatus {
	// For simplicity, just wait for a message and then end
	fmt.Print("Worker 2 waiting for message\n")
	for msg := range tw.Inbox {
		fmt.Printf("Worker 2 received message from %s: %s\n", msg.From, msg.Content)
		return WorkerEnd
	}

	return WorkerEnd
}

type testProduceWorker struct {
	SupervisedWorker
	ID string
}

func (tw *testProduceWorker) GetID() string {
	return tw.ID
}

func (tw *testProduceWorker) DoWork(c Context) WorkerStatus {
	// For simplicity, just send a message and then end
	fmt.Print("testProduceWorker sending message\n")
	tw.Outbox <- Message{
		From:    tw.ID,
		To:      "worker2",
		Type:    0,
		Content: "Hello from worker1",
	}
	return WorkerEnd

}

func TestSupervisor(t *testing.T) {
	// Test creating a new Supervisor.
	fmt.Printf("Creating supervisor\n")
	worker1 := testProduceWorker{ID: "worker1"}
	worker2 := testConsumeWorker{ID: "worker2"}
	supervisor := NewSupervisor()
	supervisor.RegisterWorker(&worker1)
	supervisor.RegisterWorker(&worker2)

	if _, ok := supervisor.actors["worker1"]; !ok {
		t.Errorf("Expected worker1 to be in supervisor.actors")
	}

	if _, ok := supervisor.actors["worker2"]; !ok {
		t.Errorf("Expected worker2 to be in supervisor.actors")
	}

	// Test starting and stopping all workers.
	supervisor.StartAll()

	// Allow some time for the goroutines to start.
	time.Sleep(1 * time.Second)

	for _, id := range supervisor.IDs {
		_, ok := supervisor.actors[id]
		if !ok {
			t.Errorf("Expected actor %s to be running", id)
		}
	}

	supervisor.StopAll()
}
