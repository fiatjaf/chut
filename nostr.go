package main

import (
	"context"
	"log"

	"github.com/nbd-wtf/go-nostr"
	sdk "github.com/nbd-wtf/nostr-sdk"
	cache_memory "github.com/nbd-wtf/nostr-sdk/cache/memory"
)

const secretKey = "01"

var sys = sdk.System{
	Pool:             nostr.NewSimplePool(context.Background()),
	RelaysCache:      cache_memory.New32[[]sdk.Relay](1000),
	MetadataCache:    cache_memory.New32[sdk.ProfileMetadata](1000),
	FollowsCache:     cache_memory.New32[[]sdk.Follow](1),
	RelayListRelays:  []string{"wss://purplepag.es", "wss://relay.nostr.band"},
	FollowListRelays: []string{"wss://public.relaying.io", "wss://relay.nostr.band"},
	MetadataRelays:   []string{"wss://nostr-pub.wellorder.net", "wss://purplepag.es", "wss://relay.nostr.band"},
}

func sendChatMessage(m model, text string) {
	evt := nostr.Event{
		Kind: 9,
		Tags: nostr.Tags{
			nostr.Tag{"h", m.groupId},
		},
		CreatedAt: nostr.Now(),
		Content:   text,
	}
	if err := evt.Sign(secretKey); err != nil {
		panic(err)
	}

	relay, err := sys.Pool.EnsureRelay(m.relayUrl)
	if err != nil {
		log.Printf("failed to connect to relay '%s': %s\n", m.relayUrl, err)
		return
	}

	if status, err := relay.Publish(m.ctx, evt); err != nil || status == nostr.PublishStatusFailed {
		log.Println("failed to publish message: ", err)
		return
	}

	program.Send(evt)
}

func subscribeToMessages(m model) {
	relay, err := sys.Pool.EnsureRelay(m.relayUrl)
	if err != nil {
		log.Printf("failed to connect to relay '%s': %s\n", m.relayUrl, err)
		return
	}

	sub, err := relay.Subscribe(m.ctx, nostr.Filters{
		{
			Kinds: []int{9},
			Tags: nostr.TagMap{
				"h": []string{m.groupId},
			},
			Limit: 500,
		},
	})
	if err != nil {
		log.Printf("failed to subscribe to group %s at '%s': %s\n", m.groupId, m.relayUrl, err)
		return
	}

	{
		stored := make([]*nostr.Event, 0, 500)
		for {
			select {
			case evt := <-sub.Events:
				// send stored messages in a big batch first
				stored = append(stored, evt)
			case <-sub.EndOfStoredEvents:
				// reverse and send
				n := len(stored)
				messages := messagesbox{events: make([]nostr.Event, n)}
				for i, evt := range stored {
					messages.events[n-1-i] = *evt
					messages.total += len(evt.Content)
				}
				program.Send(messages)
				goto continuous
			}
		}
	}

continuous:
	// after we got an eose we will just send messages as they come one by one
	for evt := range sub.Events {
		program.Send(evt)
	}
}
