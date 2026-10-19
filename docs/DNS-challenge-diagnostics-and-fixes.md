# Let's Encrypt DNS-01 challenge: diagnostics and fixes (deSEC / Cloudflare)

This guide is the write-up of a real debugging session and covers three concrete
artifacts:

1. A **quick config + manual recovery** path (the immediate unblock).
2. A **backport note** confirming the serialization/persisted-account fix that
   already ships in unstable.
3. A **defensive stale-TXT cleanup** patch that guards against the deSEC
   provider's read-modify-write race (the underlying root cause of persistent
   "recursive nameservers ... did not return the expected TXT record" errors).

---

## 1. The problem being solved

Users on **Cosmos v0.23.x (master / 0.22.x)** reported DNS-01 renewals that
never complete no matter how high the propagation wait is set:

```
resolver: one or more domains had a problem:
  [*.ineltec-t4.asera.cc: dns01: time limit exceeded:
    last error: recursive nameservers:
      NS 8.8.8.8:53 returned NXDOMAIN for _acme-challenge.ineltec-t4.asera.cc.]
```

Raising `DNSChallengePropagationWait` to **1200 seconds** did *not* help. That
rules out classic propagation timing (`deSEC` publishes in seconds).

A later run exposed the real symptom:

```
recursive nameservers: NS 1.1.1.1:53 did not return the expected TXT record
  [fqdn: _acme-challenge.ineltec-t4.asera.cc.,
   value: uMG4tV6PpC0w2YwoGaWiZhttOpe6DdCswIgqoHpcYIA]:   <- expected (new) token
   YmNg_ks2ZkraVqY7UE82ektgApWW4b7IF1GaqU-hbAM,          <- STALE token
   zwTEdwSyOGF1sbaG7nvNxxTOc8fYaV6uMrONGA8g19Q           <- STALE token
```

The `_acme-challenge` record **exists** and carries **multiple obsolete tokens**.
lego's own freshly-written token is not the one being served. This is a
**stale-token pile-up**, not a propagation failure.

## 2. Why it happens

**lego v5.3.1 `dns01` pre-check** (`challenge/dns01/dns_challenge_precheck.go`) calls in order:

1. `resolveCNAME` at the recursive NS
2. **recursive** NS propagation check (must see *exactly the new token*)
3. **authoritative** NS propagation check (from SOA lookup)

If any step sees a different TXT value than the one just written, it retries and
eventually returns "did not return the expected TXT record".

**The deSEC provider (`go-acme/lego/providers/dns/desec`)** uses a
**read-modify-write** on the whole RRSet:

- `Present()`: `GET` the TXT RRSet, `append(newValue)`, `PATCH` the set.
- `CleanUp()`: `GET` the TXT RRSet, `remove(only its own value)`, `PATCH` the set.

When two runs touch the same `_acme-challenge` name concurrently (a scheduled
renew while a config-save-triggered renew is also running — possible on master
because **master has no lock** around `DoLetsEncrypt`), the read-modify-writes
clobber each other:

- Run A reads `[old]`, writes `[old, newA]`.
- Run B reads `[old, newA]`, writes `[old, newA, newB]`.
- Run A's `CleanUp` reads `[old, newA, newB]`, removes only `newA` → leaves
  `[old, newB]` — `old` and `newB` are stale for the next run.

The result self-perpetuates: each failed run leaves its token behind.

**Cosmos 0.23.4/master** has **no serialization**: `DoLetsEncrypt()` acquires no
mutex, and every run registers a fresh Let's Encrypt account (no persisted
account). Overlapping runs are therefore easy to trigger, and the account churn
also explains the `503 rateLimited` errors seen on other setups.

## 3. Quick recovery (do this first)

### a. Clear the stale TXT once

Delete the `_acme-challenge.ineltec-t4.asera.cc` TXT RRSet entirely (or remove
the stale values) in the deSEC dashboard or via the API:

```bash
# delete the whole TXT RRSet at _acme-challenge.ineltec-t4.asera.cc
curl -X DELETE -H "Authorization: Token $DESEC_TOKEN" \
  "https://desec.io/api/v1/domains/ineltec-t4.asera.cc/rrsets/_acme-challenge/TXT/"
```

If the responsible zone is the parent `asera.cc` instead, use that zone name.
Confirm which one with:

```bash
curl -s -H "Authorization: Token $DESEC_TOKEN" \
  "https://desec.io/api/v1/domains/?owns_qname=_acme-challenge.ineltec-t4.asera.cc"
```

The response's first entry is the zone lego will write to.

### b. Use sane resolvers and a sane wait

In `cosmos.config.json` for the zone:

```json
{
  "DNSChallengeProvider": "desec",
  "DNSChallengeResolvers": "8.8.8.8,1.1.1.1",
  "DNSChallengePropagationWait": 120,
  "DisablePropagationChecks": false,
  "DNSChallengeConfig": { "DESEC_TOKEN": "..." }
}
```

Do **not** put deSEC's *authoritative* nameservers (`ns1/ns2.desec.org`) into
`DNSChallengeResolvers` — lego would then query an authoritative NS as a
"recursive" resolver, which misreports propagation.

### c. Run exactly one renewal with nothing else touching config

On master (0.23.x), a config save or route change during a renewal can start a
second LE run (race). Do renewals at a quiet moment, then wait — don't change
anything until it finishes.

## 4. Backport / upgrade note

Both **`azukaar/Cosmos-Server` and `aseracorp/Cosmos-Server` `unstable`
branches already include** the fixes that prevent this class of bug:

- `dnsChallengeEnvLock sync.Mutex` in `DoLetsEncrypt` (serializes whole DNS-01
  runs, held for the full function).
- Persisted ACME account (`acme_account.go`) — no more new LE account per run
  (fixes `503 rateLimited`).

**Upgrading to unstable is the recommended fix.** The patch in this PR is a
*defense in depth* for deployments that stay on stable/master, plus a
documented recovery procedure.

## 5. Defensive stale-TXT cleanup (this PR)

In `src/utils/certificates.go`, still fully inside the existing
`dnsChallengeEnvLock`, we add a `cleanupStaleChallengeTokens` helper that runs
*after* a successful LE order completes and removes any `_acme-challenge.*` TXT
values that were **not** written by this run.

Rationale:

- It is safe: only touches `_acme-challenge` names, never user records.
- It runs under the same lock, so it can't race another `DoLetsEncrypt`.
- It turns a permanently-poisoned RRSet back into a clean one after every
  success, breaking the self-perpetuating pile-up.

The helper reuses the zone's configured `DESEC_TOKEN` and talks to the deSEC API
directly (stdlib only, no new dependency), keeping a single credential source.
Only `_acme-challenge` TXT records are removed — never user records.

> **Upstream**: the *proper* fix belongs in `go-acme/lego`'s deSEC provider:
> make `CleanUp` remove only its own token and switch `Present`/`CleanUp` to a
> compare-and-set so concurrent runs cannot clobber each other. A Cosmos-side
> PR cannot modify a Go module dependency, which is why this defense-in-depth
> lands here and the real fix is tracked upstream.

## 6. Verification

After applying the recovery steps (or this patch + upgrade):

```bash
# while a renewal runs, confirm the token is served and then cleaned up
dig +short TXT _acme-challenge.ineltec-t4.asera.cc @8.8.8.8
```

- During a run you should see exactly **one** `_acme-challenge` token.
- After success it should disappear.
- If you ever again see **multiple** tokens, you've hit the pile-up; clear it as
  in section 3a.