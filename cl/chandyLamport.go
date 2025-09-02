package main

import (
	"fmt"
	"time"
)

// Generic message type
type Message struct {
	From int
	Data string
}

// Marker type for snapshot
type Marker struct {
	Initiator int
}

// Process structure
type Process struct {
	id            int
	state         string
	recorded      bool
	recordedState string
	channelState  map[int][]Message // channel -> messages recorded
	incoming      map[int]chan interface{}
	outgoing      map[int]chan interface{}
}

// Sending normal message
func (p *Process) sendMessage(to int, ch chan interface{}, data string) { //sending message to teh particular channel
	ch <- Message{From: p.id, Data: data}
	fmt.Printf("P%d sends '%s' to P%d\n", p.id, data, to)
}

// Sending marker message on all outgoing channels
func (p *Process) sendMarker(initiator int) {
	for _, ch := range p.outgoing {
		ch <- Marker{Initiator: initiator}
	}
	fmt.Printf("P%d sends marker on all outgoing channels\n", p.id)
}

// Handle incoming messages
func (p *Process) handleMessages() {
	for { // infinite loop

		//traverse all teh channels incomimng to the process
		for from, ch := range p.incoming {

			// handel logic based on message type
			select {
			case msg := <-ch:
				switch m := msg.(type) {
				case Message:
					fmt.Printf("P%d receives '%s' from P%d\n", p.id, m.Data, m.From)
					if p.recorded {
						// Record as in-transit if snapshot ongoing 
						p.channelState[from] = append(p.channelState[from], m)
					} else {
						// Update normal state
						p.state = fmt.Sprintf("%s|%s", p.state, m.Data)
					}

				case Marker:
					if !p.recorded {
						// First marker → record state
						p.recorded = true
						p.recordedState = p.state
						fmt.Printf("P%d records state: '%s'\n", p.id, p.recordedState)

						// Initialize empty channel states
						for src := range p.incoming {
							p.channelState[src] = []Message{}
						}

						// Forward marker
						p.sendMarker(m.Initiator)
					} else {
						// Already recorded → close channel recording
						fmt.Printf("P%d receives marker from P%d, channel state: %v\n",
							p.id, from, p.channelState[from])
					}
				}
			default:
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	// Channels (fully connected triangle)
	c12 := make(chan interface{}, 10)
	c13 := make(chan interface{}, 10)
	c21 := make(chan interface{}, 10)
	c23 := make(chan interface{}, 10)
	c31 := make(chan interface{}, 10)
	c32 := make(chan interface{}, 10)

	// Processes
	p1 := &Process{id: 1, state: "init1", incoming: map[int]chan interface{}{2: c21, 3: c31}, outgoing: map[int]chan interface{}{2: c12, 3: c13}, channelState: map[int][]Message{}}
	p2 := &Process{id: 2, state: "init2", incoming: map[int]chan interface{}{1: c12, 3: c32}, outgoing: map[int]chan interface{}{1: c21, 3: c23}, channelState: map[int][]Message{}}
	p3 := &Process{id: 3, state: "init3", incoming: map[int]chan interface{}{1: c13, 2: c23}, outgoing: map[int]chan interface{}{1: c31, 2: c32}, channelState: map[int][]Message{}}

	// Start handlers
	go p1.handleMessages()
	go p2.handleMessages()
	go p3.handleMessages()

	// Simulate normal message passing
	go func() {
		time.Sleep(50 * time.Millisecond)
		p1.sendMessage(2, c12, "A")
		time.Sleep(50 * time.Millisecond)
		p2.sendMessage(3, c23, "B")
		time.Sleep(50 * time.Millisecond)
		p3.sendMessage(1, c31, "C")
		time.Sleep(50 * time.Millisecond)
		p2.sendMessage(1, c21, "D")
	}()

	// Initiate snapshot from P1
	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("\n--- P1 initiates snapshot ---")
		p1.recorded = true
		p1.recordedState = p1.state
		for src := range p1.incoming {
			p1.channelState[src] = []Message{}
		}
		p1.sendMarker(p1.id)
	}()

	// Let simulation run
	time.Sleep(1 * time.Second)

	// Print final snapshot
	fmt.Println("\n--- Snapshot Results ---")
	for _, p := range []*Process{p1, p2, p3} {
		fmt.Printf("P%d state: '%s', channel states: %v\n", p.id, p.recordedState, p.channelState)
	}
}




// Algorithm:

// Process p:
//     state_p               // local state
//     recorded             // boolean: whether snapshot started
//     recordedState_p      // snapshot of local state
//     channelState[c_in]   // map: incoming channel -> list of in-transit messages


// Upon normal message m received from process q on channel c_in:

//     if recorded == false:
//         // Snapshot not started yet
//         state_p ← update(state_p, m)      // update local state normally
//     else:
//         // Snapshot started
//         if marker_received[c_in] == false:
//             // Channel c_in is still “open” for recording in-transit messages
//             append m to channelState[c_in]
//         else:
//             // Channel already closed, update local state normally
//             state_p ← update(state_p, m)


// Upon receiving Marker on channel c_in:

//     if recorded == false:
//         // First marker for this process → start snapshot
//         recorded ← true
//         recordedState_p ← state_p

//         // Initialize channel buffers for all incoming channels
//         for each channel c:
//             channelState[c] ← empty list
//             marker_received[c] ← false

//         // Mark channel from which this marker arrived as “closed”
//         marker_received[c_in] ← true

//         // Send marker to all outgoing channels
//         for each outgoing channel c_out:
//             send Marker to c_out

//     else:
//         // Snapshot already started
//         marker_received[c_in] ← true
//         // stop recording this channel; channelState[c_in] is now final


// Snapshot completion condition:

//     if recorded == true AND marker_received[c] == true for all incoming channels c:
//         Snapshot_p = (recordedState_p, channelState[c1], channelState[c2], ...)






// Assumptions for chandy Lamport Algorithm:
// 1. Reliable Channels: The communication channels between processes are reliable, meaning messages are neither lost nor duplicated or corrupted.
// 2. FIFO Channels: Messages sent from one process to another are received in the order they were sent.
// 3. No Global Clock: There is no global clock or synchronized time among processes; processes run asynchronously. each process operates based on its local clock.
// 4. Communication graph must be connected(should be able to reach each other directly or indirectly / the channels should be bidirectional).
// 5. Processes can atomically record their local state.