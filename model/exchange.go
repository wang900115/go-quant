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

package model

type ExchangeId byte

const (
	BINANCE ExchangeId = iota
	COINBASE
	OKX
)

type ExchangeName string

const (
	BINANCE_NAME  ExchangeName = "Binance"
	COINBASE_NAME ExchangeName = "Coinbase"
	OKX_NAME      ExchangeName = "OKX"
)

type Exchange struct {
	ID   ExchangeId
	Name ExchangeName
}

var ExchangeMap = map[ExchangeId]Exchange{
	BINANCE:  {ID: BINANCE, Name: BINANCE_NAME},
	COINBASE: {ID: COINBASE, Name: COINBASE_NAME},
	OKX:      {ID: OKX, Name: OKX_NAME},
}

func GetExchange(id ExchangeId) Exchange {
	ex, _ := ExchangeMap[id]
	return ex
}
