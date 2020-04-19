// Copyright 2020 Geraldo Augusto Massahud Rodrigues dos Santos
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// GoAwait is a simple module for asynchronous waiting.
//
// Use goawait when you need to wait for asynchronous tasks to complete before continuing normal
// execution. It is very useful for waiting on integration and end to end tests.
//
// To use it just create a spec and call one of it's UntilTrue methods
//
//     func receivedMessage() bool { ... }
//
//     goawait.AtMost(10 * time.Second).
//         RetryingEvery(200 * time.Millisecond).
//         UntilTrue(receivedMessage)
//
// GoAwait is based on Java's Awaitility's DSL: https://github.com/awaitility/awaitility
// The polling functions were based on Bill Kennedy's **retryTimeout** concurrency example at
// https://github.com/ardanlabs/gotraining/blob/0728ec842fbde65115e1a0a255b62b4a93d4c6a8/topics/go/concurrency/channels/example1/example1.go#L290
package goawait

import (
	"context"
	"time"
)

// DefaultRetryTime: 1 seconds
var defaultRetryTime = 1 * time.Second

// Await is the GoAwait specification
type Await struct {
	ctx       context.Context
	maxWait   time.Duration
	retryTime time.Duration
}

// AtMost creates a new Await with a specified timeout and default retry time of 1 second
func AtMost(maxWait time.Duration) Await {
	return Await{maxWait: maxWait, retryTime: defaultRetryTime}
}

// WithCtx sets a parent context for await Await. This context can cancel the await when Done()
func WithCtx(ctx context.Context) Await {
	return Await{ctx: ctx}
}

// AtMost configures the maximul await time of the spec
func (await Await) AtMost(maxWait time.Duration) Await {
	await.maxWait = maxWait
	return await
}

// RetryingEvery configures the Await retryTime
func (await Await) RetryingEvery(retryTime time.Duration) Await {
	await.retryTime = retryTime
	return await
}

// UntilTrue executes the polling function until the poll function returns true, or a timeout occurs
// It returns a TimeoutError on timeout.
func (await Await) UntilTrue(poll func(ctx context.Context) bool) error {
	timeoutCtx, cancel := createTimeoutContext(await)
	defer cancel()
	return UntilTrue(timeoutCtx, await.retryTime, poll)
}

func createTimeoutContext(await Await) (context.Context, context.CancelFunc) {
	var timeoutCtx context.Context
	var cancel context.CancelFunc
	if await.ctx != nil {
		if await.maxWait < 0 {
			return await.ctx, func() {}
		}
		timeoutCtx, cancel = context.WithTimeout(await.ctx, await.maxWait)

	} else {
		timeoutCtx, cancel = context.WithTimeout(context.Background(), await.maxWait)
	}
	return timeoutCtx, cancel
}

// UntilNoError executes the polling function until it does not return an error.
// It returns a TimeoutError on timeout.
func (await Await) UntilNoError(poll func(ctx context.Context) error) error {
	var timeoutCtx context.Context
	var cancel context.CancelFunc
	if await.ctx != nil {
		timeoutCtx, cancel = context.WithTimeout(await.ctx, await.maxWait)
	} else {
		timeoutCtx, cancel = context.WithTimeout(context.Background(), await.maxWait)
	}
	defer cancel()
	return UntilNoError(timeoutCtx, await.retryTime, poll)
}
