# Sproochen

Sproochen is a client-first spaced repetition learning application, specifically designed for language learning.

## Architecture

Functionality is accessed via a web client. The client is designed for both browser and progressive web app use. 
All data is stored on the client and can be optionally synced via a sync server. This has the side-effect of making the client work fully offline.

## Modes

The primary functionality is flashcards.
Flashcards should have the native language word and the target language word, defintion, example sentence, and audio.

Flashcards are in sets. Sets can contain other sets. Sets within a set can be toggled on/off. This supports functionality of having a "Language" superset divided in sets of new terms that can be enabled as learning progresses.

After seeing a card it is ranked: easy (<1s), good(1-2s), hard (2-3s), unknown (4-7s). There is an optional timer to encourage when a card is unknown.

The cards may be studied front, back, or shuffled orientation.
Progress is tied to each side of the card.

## User/accounts

User register with OIDC provider to sync with server.

## Import

Need a way to bulk import decks.

## Data

### User state

Contains learning progress. Scoped to deck.

Owned by one person, sync offline/multi device.

### Decks

Owned by one person, namespaced by ID, i.e. braden/Luxembourgish. 

Deck -> Cards
Deck -> Sets

#### Sharing decks

Can sync another person's deck (read-only).

Can fork to your own namespace. Can pull in changes from upstream deck.

### Audio / attachments

Audio (opus encoded) and other attachments are stored on sync server indexed by SHA256 sum.
When opening a deck, all attachments are downloaded to device and stored on OPFS.
New files are uploaded via separate HTTP endpoint than normal sync.

