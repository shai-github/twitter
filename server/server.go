package server

import (
	"encoding/json"
	"log"
	"proj1/feed"
	"proj1/queue"
	"sync"
)

type Config struct {
	Encoder        *json.Encoder
	Decoder        *json.Decoder
	Mode           string
	ConsumersCount int
}

type WorkerContext struct {
	queue     queue.Queue
	feed      feed.Feed
	wg        sync.WaitGroup
	condition *sync.Cond
	mu        sync.Mutex
	done      bool
}

// Represents the json objects for response with success and id attributes
type Response struct {
	Success bool
	Id      int32
}

// Run starts up the twitter server based on the configuration
// information provided and only returns when the server is fully
// shutdown.
func Run(cfg Config) {
	thisFeed := feed.NewFeed()
	if cfg.Mode == "s" {
		ctx := WorkerContext{feed: thisFeed}
		for {
			requests := &queue.Request{}
			if err := cfg.Decoder.Decode(&requests); err != nil {
				log.Println(err)
				return
			}
			if requests.Command == "DONE" {
				break
			}
			process(&ctx, &cfg, requests)
		}
	} else {
		ctx := WorkerContext{}
		ctx.condition = sync.NewCond(&ctx.mu)
		ctx.queue = queue.MakeQueue()
		ctx.feed = thisFeed
		ctx.done = false
		ctx.wg.Add(cfg.ConsumersCount)

		for i := 0; i < cfg.ConsumersCount; i++ {
			go consumer(&ctx, &cfg)
		}

		producer(&ctx, &cfg)
		ctx.wg.Wait()
	}
}

// spawns goroutines that process the request and send a requisite response
func consumer(ctx *WorkerContext, cfg *Config) {
	for {
		ctx.mu.Lock()
		for ctx.queue.Emp() {
			if ctx.done {
				ctx.mu.Unlock()
				ctx.wg.Done()
				return
			} else {
				ctx.condition.Wait()
			}
		}
		ctx.mu.Unlock()
		deqItem := ctx.queue.Deq()
		if deqItem != nil {
			process(ctx, cfg, deqItem)
		}
	}
}

// the main gorourtine that collects series of tasks and place in the queue
func producer(ctx *WorkerContext, cfg *Config) {
	for {
		requests := &queue.Request{}
		if err := cfg.Decoder.Decode(&requests); err != nil {
			log.Println(err)
		}
		if requests.Command == "DONE" {
			ctx.mu.Lock()
			ctx.done = true
			ctx.condition.Broadcast()
			ctx.mu.Unlock()
			break
		}
		ctx.queue.Enq(requests)
		ctx.condition.Signal()
	}
}

// processes the appropriate request based on the input command
func process(ctx *WorkerContext, cfg *Config, req *queue.Request) {
	thisCommand := req.Command

	switch thisCommand {
	case "ADD":
		ctx.feed.Add(req.Body, req.Timestamp)
		res := Response{Success: true, Id: req.Id}
		cfg.Encoder.Encode(res)
	case "REMOVE":
		res := Response{Success: ctx.feed.Remove(req.Timestamp), Id: req.Id}
		cfg.Encoder.Encode(res)
	case "CONTAINS":
		res := Response{Success: ctx.feed.Contains(req.Timestamp), Id: req.Id}
		cfg.Encoder.Encode(res)
	case "FEED":
		res := ctx.feed.Output()
		res.Id = int32(req.Id)
		cfg.Encoder.Encode(&res)
	}
}
