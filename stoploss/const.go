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

package stoploss

const (
	TRIGGERED_REASON_FIXED_TRAILING_STOPLOSS      = "Trailing Stop Loss Triggered"
	TRIGGERED_REASON_FIXED_TRAILING_TAKEPROFIT    = "Trailing Take Profit Triggered"
	TRIGGERED_REASON_FIXED_PERCENTCILE_STOPLOSS   = "Fixed Percentile Stop Loss Triggered"
	TRIGGERED_REASON_FIXED_PERCENTCILE_TAKEPROFIT = "Fixed Percentile Take Profit Triggered"
	TRIGGERED_REASON_FIXED_ATR_STOPLOSS           = "ATR Based Stop Loss Triggered"
	TRIGGERED_REASON_FIXED_ATR_TAKEPROFIT         = "ATR Based Take Profit Triggered"
	TRIGGERED_REASON_FIXED_MA_STOPLOSS            = "Moving Average Stop Loss Triggered"
	TRIGGERED_REASON_FIXED_MA_TAKEPROFIT          = "Moving Average Take Profit Triggered"
)

const (
	TRIGGERED_REASON_DEBOUNCED_TRAILING_STOPLOSS      = "Trailing Stop Loss Triggered with Time Delay"
	TRIGGERED_REASON_DEBOUNCED_TRAILING_TAKEPROFIT    = "Trailing Take Profit Triggered with Time Delay"
	TRIGGERED_REASON_DEBOUNCED_PERCENTCILE_STOPLOSS   = "Fixed Percentile Stop Loss Triggered with Time Delay"
	TRIGGERED_REASON_DEBOUNCED_PERCENTCILE_TAKEPROFIT = "Fixed Percentile Take Profit Triggered with Time Delay"
	TRIGGERED_REASON_DEBOUNCED_ATR_STOPLOSS           = "ATR Based Stop Loss Triggered with Time Delay"
	TRIGGERED_REASON_DEBOUNCED_ATR_TAKEPROFIT         = "ATR Based Take Profit Triggered with Time Delay"
	TRIGGERED_REASON_DEBOUNCED_MA_STOPLOSS            = "Moving Average Stop Loss Triggered with Time Delay"
	TRIGGERED_REASON_DEBOUNCED_MA_TAKEPROFIT          = "Moving Average Take Profit Triggered with Time Delay"
)

const (
	TRIGGERED_REASON_HYBRID_RISK_REWARD_STOPLOSS   = "Hybrid Risk-Reward Stop Loss Triggered"
	TRIGGERED_REASON_HYBRID_RISK_REWARD_TAKEPROFIT = "Hybrid Risk-Reward Take Profit Triggered"
	TRIGGERED_REASON_STRUCTURE_SWING_STOPLOSS      = "Structure Swing Stop Loss Triggered"
	TRIGGERED_REASON_STRUCTURE_SWING_TAKEPROFIT    = "Structure Swing Take Profit Triggered"
)
