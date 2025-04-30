
---

# Byzantine Behavior Simulation in Bamboo

This project extends the [Bamboo](https://github.com/gitferry/bamboo) consensus framework to simulate and benchmark various **Byzantine faults** in HotStuff-based BFT protocols. It also aims to abstract Byzantine behaviors for use in other blockchain systems and developer networks.

---

## Goals

- Define and implement **Byzantine behaviors** inside the Bamboo framework.
- Benchmark the **impact on TPS and Safety** under Byzantine faults.
- Create a **reusable fault injection package** applicable to other BFT environments.
- Design for eventual **cross-chain abstraction** and applicability (e.g., Ethereum, Solana, Cosmos SDK, Sui).

---

## Current Status

### Defined Byzantine Behaviors in Bamboo

- `silence`: Node does not send any messages.
- `crash`: Node process is forcefully terminated.
- `drop`: Node drops specific messages.
- `delay`: Artificial delays are added to messages.

Defined in `config/byzantine.go`, behaviors can be enabled per node via configuration.

---

## Extensible Faults: To Be Added

- `equivocation`: Sending conflicting messages in the same round.
- `invalid-signature`: Sending messages with forged or wrong signatures.
- `fake-commit`: Broadcasting commit messages for nonexistent blocks.
- `view-change-abuse`: Spamming or spoofing view change requests.
- `adaptive behavior`: Faults triggered under specific consensus conditions (e.g., when node becomes leader).
- `message-mutation`: Corrupting messages with arbitrary salts or bit-flipping.

---

## Cross-Protocol Applicability

| Fault Type        | Compatible Protocols             | Notes                                           |
|-------------------|----------------------------------|-------------------------------------------------|
| silence           | PBFT, HotStuff, CosmosBFT        | Universally applicable                          |
| crash             | All (simulation required)         | Best used in testnets or local environments     |
| drop              | All                              | Can simulate network failure                    |
| equivocation      | PBFT, HotStuff                   | Key Byzantine property                          |
| fake-commit       | HotStuff variants                | May not apply to all protocols                  |
| view-change-abuse | View-change based BFTs           | Not applicable to protocols without view-change |
| mutation          | Protocols with message signing   | Can trigger signature validation failures       |

---

## Benchmark Abstraction

Byzantine behaviors can be defined per node using a simple JSON/YAML format:

```json
{
  "byzantine_behavior": {
    "node_id": "n3",
    "trigger": {
      "block_height": 10
    },
    "actions": [
      "crash",
      "fake_commit"
    ]
  }
}
```

This allows the benchmark tool to inject faults **automatically** when specific internal conditions are met (e.g., block height, view number, or leader status).

---

## DevNet Extension Plan

| Chain           | Injection Strategy     | Feasibility       | Notes                            |
|-----------------|------------------------|-------------------|----------------------------------|
| Ethereum (Geth) | Validator code patch   | ✅ DevNet only     | Requires slashing-safe testnet  |
| Solana          | Local test-validator   | ⚠️ Limited         | No message tampering without core change |
| Cosmos SDK      | Validator patch        | ✅ Testnet          | Easier to modify consensus logic |
| Sui             | Move testnet programs  | ⚠️ Limited         | Messages aren't accessible at network level |

In most cases, direct validator modification or simulated attacker nodes are necessary. Public testnets with slashing must be avoided.

---

## `byzinject`: Reusable Fault Injection Package

We aim to extract the logic into a reusable package:
- `injector.go`: Core logic to execute Byzantine actions.
- `config.go`: Load YAML/JSON behavior configuration.
- `logger.go`: Capture all fault actions and state changes.
- No external proxy or trigger service — internal condition-based automation.

---

## Coordinated Byzantine Attacks (Future Work)

- Multiple nodes can share the same behavior and **synchronize faults** (e.g., all fake-commit at the same round).
- Useful to simulate **colluding Byzantine replicas** or **majority faults** in theoretical edge cases.

---

## Example Scenario: Block Height Trigger

1. **Trigger**: Block height reaches 10.
2. **Action**: Node crashes (simulated process kill).
3. **Effect**: Benchmark tool tracks TPS drop and any violations of Safety properties.
4. **Log**: All actions are timestamped and recorded for post-analysis.

---
