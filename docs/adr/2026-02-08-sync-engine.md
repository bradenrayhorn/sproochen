# Local first sync engines

Up for consideration:

CRDT libraries:

- Automerge ⭐️
- YJS
- Loro

Full local-first solution:
- LiveStore
- Jazz

Databases that sync:
- ElectricSQL
- PouchDB
- Turso

Out of consideration:
- Triplit (promising option, but no longer actively maintained)
- Zero (not local first; meant for client <-> server syncing only)
- Instant (a hosted platform)

**Winner:** Automerge will be trialed as the sync core. Reasons include:
- Stable release available & actively maintained
- Team behind it is committed to local-first research and ideas
- Because it is the sync core; the rest can be customized for this application (also a downside as it requires more work)
- Loading entire document into memory should be fine for this project as it is text

## CRDT Libraries

### Automerge ⭐️

Syncs documents. Documents must be full loaded in memory to use. Events are not materialized to a
local database.

Stable major release available.

Authentication/server is mostly left to end user. This is really just the core library.

(*) Note, updates are being actively developed for a local-first authentication scheme.

### YJS

One of the originally CRDT libraries, mostly focused on text editing.

Authentication/server is mostly left to the end user, it's mostly the sync engine, like Automerge.

### Loro

Newer CRDT library that seems to have more features than Automerge, but less mature.

Not many syncing and server examples.


## Full solution

### LiveStore

True event log that materializes state into a local SQLite database.
Offers some sync providers, including ElectricSQL.

Relatively new and missing some features like compaction, but under active development.

Keep an eye on future releases.

### Jazz

Relatively new but has a full suite of features, including authentication, for syncing and handling conflicts between users and devices.
It seems like it may be difficult to break out of prescribed path if it doesn't meet the exact needs of the project.
It's a little more opaque than other solutions.

Maybe the most fully featured out of the box solution.

## Databases

### ElectricSQL

Not a conflict resolution engine.

Syncs data one way from remote Postgres to local client.
Does not handle writes.

Would need some other mechanism on top of ElectricSQL to handle writes.

### PouchDB

Older and more mature offering. Syncs directly to CouchDB on the server.

Not frequently released, but to be expected given its maturity.

No real auth strategy. Need to create "one database per user" for authentication.

### Turso

Currently a cloud platform but working on offering a core database as open source, currently in beta.

Database supports E2EE.

