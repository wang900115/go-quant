// Copyright (C) 2025 Quantive
//
// SPDX-License-Identifier: MIT OR AGPL-3.0-or-later
//
// This file is part of the Decision Engine project.
// You may choose to use this file under the terms of either
// the MIT License or the GNU Affero General Public License v3.0 or later.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the LICENSE files for more details.

package sys

import (
	"context"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

type Engine struct {
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	HealthCheck   time.Duration
	RetryInterval time.Duration
}

func NewEngine(retry time.Duration, health time.Duration) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{
		ctx:           ctx,
		cancel:        cancel,
		RetryInterval: retry,
		HealthCheck:   health,
	}
}

func (e *Engine) Go(fn func(ctx context.Context), recoverFunc func(r any)) {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()

		defer func() {
			if r := recover(); r != nil {
				if recoverFunc != nil {
					recoverFunc(r)
				} else {
					log.Printf("Recovered from panic in engine goroutine: %v", r)
				}
			}
		}()
		fn(e.ctx)
	}()
}

func (e *Engine) SafeGo(fn func(ctx context.Context), restartFunc func()) {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()

		for {
			done := make(chan struct{})
			panicCh := make(chan any, 1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						select {
						case panicCh <- r:
						default:
						}
					}
				}()
				fn(e.ctx)
				close(done)
			}()

			select {
			case <-done:
				return
			case r := <-panicCh:
				log.Printf("[Engine.SafeGo] goroutine panic recovered: %v\n%s", r, debug.Stack())
				if restartFunc != nil {
					restartFunc()
				}
			case <-e.ctx.Done():
				return
			}

			select {
			case <-e.ctx.Done():
				return
			case <-time.After(e.RetryInterval):
			}
		}
	}()
}

func (e *Engine) Stop() {
	e.cancel()
	e.wg.Wait()
}

func (e *Engine) Done() <-chan struct{} {
	return e.ctx.Done()
}
