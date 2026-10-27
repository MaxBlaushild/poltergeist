# Travel calendar setup

Travel Angels includes an ongoing personal calendar, individual trip permissions, guest browsing and email verification, invitation claiming, trip hosts, participation requests, guest subscriptions, and email broadcasts. Each calendar stop opens a trip page with a cover image, attendance, lodging, and flights.

**Database.** Apply the normal repository migration workflow to the intended environment before starting the updated API. Migration `000473_travel_calendars` adds the calendar, stop, access, collaboration, and participation tables. Migration `000474_travel_calendar_guests` adds verified recipient identities, invitation/subscription records, and the notification outbox. Migration `000475_travel_trip_pages` adds host roles, capacity and roster settings, reserved places, revision checks, an audit log, and protected trip media. The server does not automatically modify the schema. Both up and down migrations are included; rolling down removes feature data.

**API configuration.** The same calendar routes are registered in the standalone Travel Angels service and the aggregated core API. Configure the process running those routes:

| Variable | Purpose |
| --- | --- |
| `TRAVEL_ANGELS_WEB_URL` | Public HTTP(S) address of the hosted Flutter app, without a fragment. Used for calendar, stop, and preference links. |
| `TRAVEL_ANGELS_EMAIL_FROM` | Verified email sender address; falls back to `EMAIL_FROM_ADDRESS`. |
| `TWILIO_ACCOUNT_SID` | Account used by the existing email client. |
| `TWILIO_AUTH_TOKEN` | Credential used by the existing email client. |

Missing mail configuration produces an actionable unavailable response for verification/invitation/broadcast actions rather than reporting an email as sent. Calendar creation and permission-controlled link browsing remain available. The shared email client accepts an injected HTTP client; Travel Angels sets a 30-second request timeout.

Configured API processes drain the durable notification outbox every 10 seconds. Deliveries are rechecked against current stop permissions and subscriptions before sending. Known failed deliveries can be retried through the UI. A delivery with an uncertain provider result remains in `sending` for review, because blindly resending could duplicate a message. Successful deliveries are not automatically resent.

**Web and mobile build configuration.** The Flutter API URL retains its existing default. Override it when targeting another environment:

```sh
flutter build web \
  --dart-define=TRAVEL_ANGELS_API_URL=https://your-core-api.example \
  --dart-define=TRAVEL_ANGELS_WEB_URL=https://your-travel-app.example
```

The web URL can be inferred from the browser origin for browser-created links; provide the build define for native app sharing. The configured backend also returns share URLs. Use the aggregated core API for split-origin development: its CORS policy accepts `X-Travel-Guest-Token` alongside account authorization. Standalone Travel Angels retains its existing same-origin/CORS behavior.

Shared URLs use fragments, so ordinary static hosting does not need a separate rewrite rule for each calendar:

- `/#/calendar/<share-token>` opens a filtered calendar.
- `/#/stops/<share-token>` opens its authorized trip page directly.
- `/#/preferences/<management-token>` manages notifications without an app account.

Path-based and `travelangels://` equivalents are also understood by the app. Shared calendar and stop routes bypass the ordinary account login gate. Restricted recipients verify their email for limited guest access. Joining then uses the existing phone-based account login/registration and explicitly associates the verified invitation; the UI returns to review instead of submitting automatically.

**Calendar behavior.** New calendars are private. Trips inherit the owning calendar's audience unless set to private, specific people, or anyone with the link. Trip overrides replace the calendar audience. Displaying a hosted trip on another calendar preserves its owning policy. Calendar editors retain their existing rights on legacy stops; newly created trips use explicit, accepted full-host roles. Only a calendar's owner can create new trips in it.

**Trip hosts and attendance.** New trips list their creator as an attending host. Hosts can edit and publish content, manage the audience, approve attendance, and invite or remove non-owner hosts. Invited hosts must accept; accepting as an organizer does not reserve a place. Ownership transfer and deletion belong to the owner. Existing stops keep their legacy roles until the owner reviews and confirms conversion; collaborators are never silently upgraded. Attendance limits count distinct confirmed people and places reserved during date reconfirmation. Optional waitlists require manual host approval. Guests can browse authorized details, but must claim an account and explicitly submit their attendance choice. Roster consent is required before sharing a participant's identity with other attendees.

**Trip images and logistics.** Cover uploads accept JPEG and PNG up to 8 MB and 20 million pixels. The API re-encodes images, strips metadata, and stores the protected bytes in PostgreSQL; it needs no new storage service or public media bucket. Image requests recheck trip access. Include this media table in normal database backup and storage planning. Lodging and flight records live in the trip's draft and published content. Their audiences can be narrower than the trip's audience; exact lodging addresses and arrival instructions use separate restrictions. Host notes remain private to hosts. Publishing content and applying immediate permission changes both check the current trip revision, so a stale editor cannot overwrite a host's newer changes.

**Trip destinations.** Migration `000477_travel_destinations` adds a canonical destination catalog, indexed search terms, and private per-user selection history. It seeds 249 countries and territories, so country autocomplete works without an external lookup. The authenticated `GET /travel-angels/destinations?q=...` endpoint searches names and aliases and ranks the caller's recently/frequently selected places. New places can be resolved through the existing Google Maps client (`GOOGLE_MAPS_API_KEY`); provider failures leave the local catalog available. Country relationships use structured provider data, and ambiguous regions do not imply a time zone.

Trips accept an ordered `destinations` array of up to 12 canonical IDs or custom names. IDs resolve to server-controlled snapshots stored independently in draft and published content. Custom names stay local to the trip and never enter the shared catalog. The legacy `destination` summary remains available to older clients, and existing free-text destinations are preserved as one location rather than split or guessed. Unrelated edits from older clients preserve structured locations. Changing selected places or their order requires the same attendance reconfirmation as other material trip changes. Place metadata and theme fields provide a foundation for later destination styling; destination pages, automatic styling, and per-location dates are not part of this release.

Stop content is drafted before publication, while confirmed sharing changes apply immediately. Access changes preview affected participation. Revoking a person's last access path removes their active participation and pauses updates. Link rotation checks the actual links previously used by participants/subscribers; it does not grant access through an alternative link they never received. Dates remain local `YYYY-MM-DD` values. A destination IANA time zone, defaulting explicitly to UTC, determines when a stop has ended.

**Verification.** From the Flutter package, run `flutter test` and `flutter build web --no-pub`. From `go/travel-angels`, run `go test ./...`. Database integration tests are included in `internal/travelcalendar`; set `TC_TEST_DATABASE_URL` to a disposable PostgreSQL database to run them. Each integration test creates a unique schema, loads the actual migrations, uses a fake mail sender, and removes that schema afterward. The supplied database role needs schema-creation permission. Coverage includes publication, permission replacement, account claiming, partial dates, reconfirmation, capacity, host roles, revocation/rotation, protected media, logistics visibility, delivery retries, and migration rollback/reapplication.

**Release.** Apply repository migrations through 477 before rolling out the API that serves Travel Angels (the aggregated core service or standalone Travel Angels service, depending on the environment), then release the rebuilt Flutter web/iOS clients. Existing trip links and calendar references remain valid. No new service or API credential is required; destination lookup uses the existing Google Maps configuration.

This change does not provision web hosting, change production configuration, run production migrations, or send invitations. Calendar access, email delivery, and existing phone authentication must be configured in the environment where the feature is deployed.
