// package main

// import (
// 	"fmt"
// 	"sync"
// 	"time"
// )

// type Color int

// const (
// 	WHITE Color = iota
// 	RED
// )

// // LYMessage piggybacks the sender's color.
// type LYMessage struct {
// 	IsRed   bool
// 	Content int
// }

// // processLY simulates a node using the Lai-Yang algorithm.
// func processLY(id int, balance int, incoming map[int]chan LYMessage, outgoing map[int]chan LYMessage, wg *sync.WaitGroup) {
// 	defer wg.Done()

// 	color := WHITE
// 	localState := -1
// 	channelStates := make(map[int][]LYMessage)
// 	heardFrom := make(map[int]bool)

// 	// Channel to trigger the snapshot from main
// 	snapshotTrigger := make(chan bool, 1)

// 	// Goroutine to simulate some application work
// 	go func() {
// 		if id == 0 {
// 			time.Sleep(1 * time.Second)
// 			snapshotTrigger <- true
// 		}
// 		for i := 1; i <= 2; i++ {
// 			time.Sleep(time.Duration(500+id*100) * time.Millisecond)
// 			amount := 10 * i
// 			for destID, ch := range outgoing {
// 				fmt.Printf("P%d (%s) -> P%d: $%d\n", id, colorToString(color), destID, amount)
// 				ch <- LYMessage{IsRed: (color == RED), Content: amount}
// 			}
// 		}
// 	}()

// 	// define handleMessage *before* use
// 	handleMessage := func(fromID int, msg LYMessage) {
// 		if color == WHITE && msg.IsRed {
// 			// --- A white process receives a red message ---
// 			fmt.Printf("P%d turned RED by message from P%d\n", id, fromID)
// 			color = RED
// 			localState = balance // Record local state
// 			// Initialize channel states for recording
// 			for chanID := range incoming {
// 				channelStates[chanID] = []LYMessage{}
// 			}
// 			heardFrom[fromID] = true // This channel's state is empty
// 		} else {
// 			// Normal processing
// 			balance += msg.Content
// 			if color == RED && !msg.IsRed {
// 				// A red process records any white messages it receives
// 				channelStates[fromID] = append(channelStates[fromID], msg)
// 			}
// 			// If we are red, we mark that we've heard from this channel
// 			if color == RED {
// 				heardFrom[fromID] = true
// 			}
// 		}
// 	}

// 	for {
// 		select {
// 		case <-snapshotTrigger:
// 			// --- Initiator becomes red ---
// 			if color == WHITE {
// 				fmt.Printf("\n--- SNAPSHOT INITIATED BY P%d ---\n", id)
// 				color = RED
// 				localState = balance // Record local state
// 				// Initialize channel states for recording
// 				for chanID := range incoming {
// 					channelStates[chanID] = []LYMessage{}
// 				}
// 			}

// 		case msg := <-incoming[0]:
// 			handleMessage(0, msg)
// 		case msg := <-incoming[1]:
// 			handleMessage(1, msg)
// 		case msg := <-incoming[2]:
// 			handleMessage(2, msg)
// 		}

// 		// Termination condition
// 		if color == RED && len(heardFrom) == len(incoming) {
// 			fmt.Printf("\n--- SNAPSHOT RESULT FOR P%d ---\n", id)
// 			fmt.Printf("  Local State (Balance): $%d\n", localState)
// 			for chanID, messages := range channelStates {
// 				fmt.Printf("  Incoming Channel P%d State: %v\n", chanID, messages)
// 			}
// 			fmt.Println("--------------------------------")
// 			return
// 		}
// 	}
// }

// func colorToString(c Color) string {
// 	if c == RED {
// 		return "RED"
// 	}
// 	return "WHITE"
// }

// func main() {
// 	fmt.Println("\n\n===== Running Lai-Yang Algorithm =====")

// 	var wg sync.WaitGroup
// 	const N = 3 // Number of processes

// 	channels := make([][]chan LYMessage, N)
// 	for i := 0; i < N; i++ {
// 		channels[i] = make([]chan LYMessage, N)
// 		for j := 0; j < N; j++ {
// 			if i != j {
// 				channels[i][j] = make(chan LYMessage, 10)
// 			}
// 		}
// 	}

// 	wg.Add(N)

// 	for i := 0; i < N; i++ {
// 		incoming := make(map[int]chan LYMessage)
// 		outgoing := make(map[int]chan LYMessage)
// 		for j := 0; j < N; j++ {
// 			if i != j {
// 				incoming[j] = channels[j][i]
// 				outgoing[j] = channels[i][j]
// 			}
// 		}
// 		go processLY(i, 1000, incoming, outgoing, &wg)
// 	}

// 	wg.Wait()
// 	fmt.Println("\nAll processes finished. Snapshot complete.")
// }

package main

import (
	"fmt"
	"time"
)

type Color string

const (
	White Color = "white"
	Red   Color = "red"
)

type LYMessage struct {
	From  int
	Data  string
	Color Color // piggybacked color of sender at send time
}

type Process struct {
	id           int
	color        Color
	state        string
	recorded     bool
	recordedState string

	incoming    map[int]chan interface{} // key = source process id
	outgoing    map[int]chan interface{} // key = dest process id

	// Lai-Yang bookkeeping:
	// store white messages received after turning red on a per-channel basis
	inTransit map[int][]LYMessage
	// whether this process has seen a red message on incoming channel from src
	redSeen map[int]bool
}

// send a normal (colored) message using the sender's current color
func (p *Process) send(to int, ch chan interface{}, data string) {
	msg := LYMessage{From: p.id, Data: data, Color: p.color}
	ch <- msg
	fmt.Printf("P%d (color=%s) sends '%s' to P%d\n", p.id, p.color, data, to)
}

// action when a process turns red (first time)
func (p *Process) turnRed() {
	if p.color == Red {
		return
	}
	p.color = Red
	p.recorded = true
	p.recordedState = p.state
	// initialize maps
	p.inTransit = make(map[int][]LYMessage)
	p.redSeen = make(map[int]bool)
	fmt.Printf("P%d turns RED and records state: '%s'\n", p.id, p.recordedState)
	// after turning red, future sends are in red (piggyback color)
}

// per-channel message handling
// func (p *Process) handle() {
// 	for {
// 		for src, ch := range p.incoming {
// 			select {
// 			case raw := <-ch:
// 				switch m := raw.(type) {
// 				case LYMessage:
// 					// If sender sent as red and receiver is white -> receiver turns red
// 					if m.Color == Red && p.color == White {
// 						// Receiving a red message from src triggers turnRed
// 						// Mark that this incoming channel has seen red (because this message is red)
// 						p.turnRed()
// 						p.redSeen[src] = true
// 						fmt.Printf("P%d receives RED message from P%d on channel %d (closing that channel's recording)\n", p.id, m.From, src)
// 						// Note: a red message itself is not recorded in inTransit (it's a red message)
// 						// Any messages received on other incoming channels afterwards that are white should be recorded.
// 					} else {
// 						// Normal delivery behavior:
// 						if p.recorded {
// 							// If process already turned red but this incoming channel has NOT yet seen a red message,
// 							// and the message itself is white, then this message is part of the channel's in-transit (per Lai-Yang)
// 							if m.Color == White && !p.redSeen[src] {
// 								p.inTransit[src] = append(p.inTransit[src], m)
// 								fmt.Printf("P%d (red) records in-transit message on channel %d: {from:%d '%s' color=%s}\n",
// 									p.id, src, m.From, m.Data, m.Color)
// 							} else {
// 								// Otherwise (msg is red or channel already closed), treat as normal delivery
// 								p.state = fmt.Sprintf("%s|%s", p.state, m.Data)
// 								fmt.Printf("P%d (red) applies message from P%d: '%s' -> state now '%s'\n", p.id, m.From, m.Data, p.state)
// 							}
// 						} else {
// 							// If not recorded yet, normal processing: update state
// 							p.state = fmt.Sprintf("%s|%s", p.state, m.Data)
// 							fmt.Printf("P%d (white) applies message from P%d: '%s' -> state now '%s'\n", p.id, m.From, m.Data, p.state)
// 						}
// 					}
// 				default:
// 					// ignore unknown types
// 				}
// 			default:
// 			}
// 		}
// 		time.Sleep(10 * time.Millisecond)
// 	}
// }
func (p *Process) handle() {
	for {
		for src, ch := range p.incoming {
			select {
			case raw := <-ch:
				switch m := raw.(type) {
				case LYMessage:
					switch {
					// Case 1: red message arrives at a white process
					case m.Color == Red && p.color == White:
						p.turnRed()
						p.redSeen[src] = true
						fmt.Printf("P%d receives RED message from P%d on channel %d (closing that channel's recording)\n",
							p.id, m.From, src)

					// Case 2: process is already red
					case p.recorded:
						if m.Color == White && !p.redSeen[src] {
							// white message received after turning red → in-transit
							p.inTransit[src] = append(p.inTransit[src], m)
							fmt.Printf("P%d (red) records in-transit message on channel %d: {from:%d '%s' color=%s}\n",
								p.id, src, m.From, m.Data, m.Color)
						} else {
							// apply normally
							p.state = fmt.Sprintf("%s|%s", p.state, m.Data)
							fmt.Printf("P%d (red) applies message from P%d: '%s' -> state now '%s'\n",
								p.id, m.From, m.Data, p.state)
						}

					// Case 3: process is white and receives a white message
					default:
						p.state = fmt.Sprintf("%s|%s", p.state, m.Data)
						fmt.Printf("P%d (white) applies message from P%d: '%s' -> state now '%s'\n",
							p.id, m.From, m.Data, p.state)
					}

				default:
					// ignore unknown types
				}
			default:
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}


func main() {
	// create channels (fully connected triangle)
	c12 := make(chan interface{}, 10) // from 1->2
	c13 := make(chan interface{}, 10) // 1->3
	c21 := make(chan interface{}, 10)
	c23 := make(chan interface{}, 10)
	c31 := make(chan interface{}, 10)
	c32 := make(chan interface{}, 10)

	// construct processes
	p1 := &Process{
		id:        1,
		color:     White,
		state:     "init1",
		incoming:  map[int]chan interface{}{2: c21, 3: c31},
		outgoing:  map[int]chan interface{}{2: c12, 3: c13},
	}
	p2 := &Process{
		id:        2,
		color:     White,
		state:     "init2",
		incoming:  map[int]chan interface{}{1: c12, 3: c32},
		outgoing:  map[int]chan interface{}{1: c21, 3: c23},
	}
	p3 := &Process{
		id:        3,
		color:     White,
		state:     "init3",
		incoming:  map[int]chan interface{}{1: c13, 2: c23},
		outgoing:  map[int]chan interface{}{1: c31, 2: c32},
	}

	// start handlers
	go p1.handle()
	go p2.handle()
	go p3.handle()

	// Simulate message sends so we can get some white in-transit messages.
	// Step: send a few white messages, then make p1 turn red (initiate), then send more messages.
	go func() {
		time.Sleep(50 * time.Millisecond)
		p1.send(2, c12, "A") // white (p1 is white now)
		time.Sleep(40 * time.Millisecond)
		p2.send(3, c23, "B") // white
		time.Sleep(40 * time.Millisecond)
		p3.send(1, c31, "C") // white
		// Now trigger snapshot at p1
		time.Sleep(40 * time.Millisecond)
		fmt.Println("\n--- P1 initiates Lai-Yang snapshot (turns red) ---")
		p1.turnRed()
		// After turning red, p1 will send red messages
		// sending a white message from p2 to p1 that may be in-transit depending on timing
		time.Sleep(20 * time.Millisecond)
		p2.send(1, c21, "D") // p2 still white → this D may be in-transit as white if p1 already red
		time.Sleep(10 * time.Millisecond)
		// now p2 may receive a red message from someone and turn red, etc., but we also let processes continue
		// Now p1 sends a red message
		p1.send(2, c12, "E") // red (piggybacked)
		time.Sleep(20 * time.Millisecond)
		p3.send(2, c32, "F") // maybe white or red depending on p3 status
	}()

	// Allow time for protocol to run
	time.Sleep(1 * time.Second)

	// Print final Lai-Yang snapshot summary
	fmt.Println("\n--- Lai-Yang Snapshot Results (recorded states & in-transit white messages) ---")
	for _, p := range []*Process{p1, p2, p3} {
		if p.recorded {
			fmt.Printf("P%d recordedState: '%s'\n", p.id, p.recordedState)
			fmt.Printf("  in-transit messages (per incoming channel):\n")
			for src, msgs := range p.inTransit {
				fmt.Printf("    from P%d: %v\n", src, msgs)
			}
		} else {
			fmt.Printf("P%d did NOT record (still white)\n", p.id)
		}
	}
}
