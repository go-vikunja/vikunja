"""Cost-logging wrapper around the Anthropic API.

Every call made through `CostLogger.call(...)` is logged to a JSONL file
(timestamp, model, input/output tokens, computed cost) before the response
is returned. Later phases (triage agent, resolve agent) should route all
their Claude API calls through this wrapper so the $50 budget stays
visible from hour zero, per Phase 0 of the plan.

Usage:
    from agent.cost_logger import CostLogger

    logger = CostLogger()
    response = logger.call(
        model="claude-haiku-4-5",
        max_tokens=1024,
        messages=[{"role": "user", "content": "hello"}],
    )
    print(logger.total_cost())

Run directly for a smoke test that makes one real, cheap API call and
prints the logged entry:
    python3 agent/cost_logger.py
"""

from __future__ import annotations

import json
import os
import time
from dataclasses import dataclass, asdict
from pathlib import Path
from typing import Any

import anthropic

LOG_PATH = Path(__file__).parent / "logs" / "cost_log.jsonl"

# $ per million tokens (input, output). Source: platform.claude.com/docs/en/about-claude/pricing, Aug 2026.
# Matched by prefix against the model id passed to `call()` — update here if pricing changes.
PRICING_PER_MTOK: dict[str, tuple[float, float]] = {
    "claude-opus": (5.00, 25.00),
    "claude-sonnet": (2.00, 10.00),
    "claude-haiku": (1.00, 5.00),
}

DEFAULT_PRICING = (3.00, 15.00)  # conservative fallback for an unrecognized model id


def _price_for_model(model: str) -> tuple[float, float]:
    for prefix, prices in PRICING_PER_MTOK.items():
        if model.startswith(prefix):
            return prices
    return DEFAULT_PRICING


@dataclass
class CallRecord:
    timestamp: str
    model: str
    input_tokens: int
    output_tokens: int
    cache_creation_input_tokens: int
    cache_read_input_tokens: int
    cost_usd: float
    label: str | None = None


class CostLogger:
    """Wraps an Anthropic client and logs every `messages.create` call made through it."""

    def __init__(self, api_key: str | None = None, log_path: Path = LOG_PATH):
        self.client = anthropic.Anthropic(api_key=api_key or os.environ.get("ANTHROPIC_API_KEY"))
        self.log_path = log_path
        self.log_path.parent.mkdir(parents=True, exist_ok=True)

    def call(self, *, label: str | None = None, **kwargs: Any):
        """Makes a messages.create call and logs cost. kwargs are passed through unchanged."""
        response = self.client.messages.create(**kwargs)

        usage = response.usage
        input_price, output_price = _price_for_model(kwargs.get("model", ""))
        cache_creation = getattr(usage, "cache_creation_input_tokens", 0) or 0
        cache_read = getattr(usage, "cache_read_input_tokens", 0) or 0

        cost = (
            usage.input_tokens * input_price
            + usage.output_tokens * output_price
            + cache_creation * input_price * 1.25
            + cache_read * input_price * 0.1
        ) / 1_000_000

        record = CallRecord(
            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            model=kwargs.get("model", ""),
            input_tokens=usage.input_tokens,
            output_tokens=usage.output_tokens,
            cache_creation_input_tokens=cache_creation,
            cache_read_input_tokens=cache_read,
            cost_usd=round(cost, 6),
            label=label,
        )
        with self.log_path.open("a") as f:
            f.write(json.dumps(asdict(record)) + "\n")

        return response

    def total_cost(self) -> float:
        if not self.log_path.exists():
            return 0.0
        total = 0.0
        with self.log_path.open() as f:
            for line in f:
                if line.strip():
                    total += json.loads(line)["cost_usd"]
        return round(total, 6)

    def summary_by_label(self) -> dict[str, float]:
        totals: dict[str, float] = {}
        if not self.log_path.exists():
            return totals
        with self.log_path.open() as f:
            for line in f:
                if not line.strip():
                    continue
                rec = json.loads(line)
                key = rec.get("label") or "(unlabeled)"
                totals[key] = round(totals.get(key, 0.0) + rec["cost_usd"], 6)
        return totals


if __name__ == "__main__":
    logger = CostLogger()
    resp = logger.call(
        model="claude-haiku-4-5",
        max_tokens=64,
        messages=[{"role": "user", "content": "Reply with exactly: cost logger wired up."}],
        label="phase0-smoke-test",
    )
    print("Response:", resp.content[0].text)
    print("Logged to:", LOG_PATH)
    print("Total cost so far: $%.6f" % logger.total_cost())
