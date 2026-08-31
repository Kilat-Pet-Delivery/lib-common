---
name: project_state
description: The resume point for this repo - current checkpoint (sha, environment, open units table, recommended next unit) at the top, earlier checkpoints below. Read first in every session; rewritten by /recap.
metadata:
  type: project
---

## 2026-08-31 state (resume here)

- **Repo:** `main` @ `e78f944` - Merge pull request #1 from Kilat-Pet-Delivery/add-mit-license
- **Environment:** dev-infra stack up (`./dev.ps1 up kilat`).
- **Open units**

| Unit / ticket | State | Blocked on | Note |
|---|---|---|---|
| KPD-64 + KPD-66 Kafka producer: topic auto-creation and the unguarded map write | In Review | review | PR #2. KPD-64 is a PARTIAL fix - the publish that triggers topic creation is still lost (KPD-67) |
| KPD-63 gofmt sweep | In Review | review | PR #3 |
| KPD-53 reusable Go CI workflow | In Progress | **a credential scope** | written and verified locally; push blocked until `gh auth refresh -h github.com -s workflow` |

- **Recommended next unit:** KPD-67 - pre-create topics from topics.json, the complete fix for the dropped first event.
- **Waiting on Luqman:** merge the open PRs above. Several are stacked, so order matters.

## Earlier checkpoints

(none - this layer was created 2026-08-31 under KPD-51)
