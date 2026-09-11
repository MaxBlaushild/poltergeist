# Travel calendar setup

Travel Angels now includes an ongoing personal calendar, individual stop permissions, guest browsing and email verification, invitation claiming, co-hosting, participation requests, guest subscriptions, and email broadcasts.

**Database.** Apply the normal repository migration workflow to the intended environment before starting the updated API. Migration `000473_travel_calendars` adds the calendar, stop, access, collaboration, and participation tables. Migration `000474_travel_calendar_guests` adds verified recipient identities, invitation/subscription records, and the notification outbox. The server does not automatically modify the schema. Both up and down migrations are included; rolling down removes feature data.

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
- `/#/stops/<share-token>` opens one authorized stop.
- `/#/preferences/<management-token>` manages notifications without an app account.

Path-based and `travelangels://` equivalents are also understood by the app. Shared calendar and stop routes bypass the ordinary account login gate. Restricted recipients verify their email for limited guest access. Joining then uses the existing phone-based account login/registration and explicitly associates the verified invitation; the UI returns to review instead of submitting automatically.

**Calendar behavior.** New calendars are private. Stops inherit the owning calendar's audience unless set to private, specific people, or anyone with the link. Stop overrides replace the calendar audience. Displaying a co-hosted stop on another calendar preserves its owning policy. Calendar editors can manage the calendar's owned stops; referenced stops keep their own grants.

Stop content is drafted before publication, while confirmed sharing changes apply immediately. Access changes preview affected participation. Revoking a person's last access path removes their active participation and pauses updates. Link rotation checks the actual links previously used by participants/subscribers; it does not grant access through an alternative link they never received. Dates remain local `YYYY-MM-DD` values. A destination IANA time zone, defaulting explicitly to UTC, determines when a stop has ended.

**Verification.** From the Flutter package, run `flutter test` and `flutter build web --no-pub`. From `go/travel-angels`, run `go test ./...`. Database integration tests are included in `internal/travelcalendar`; set `TC_TEST_DATABASE_URL` to a disposable PostgreSQL database to run them. Each integration test creates a unique schema, loads the actual migrations, uses a fake mail sender, and removes that schema afterward. The supplied database role needs schema-creation permission. The tests cover publication, permission replacement, account claiming, partial dates, reconfirmation, revocation/rotation, co-host references, delivery retries, and migration rollback/reapplication.

This change does not provision web hosting, change production configuration, run production migrations, or send invitations. Calendar access, email delivery, and existing phone authentication must be configured in the environment where the feature is deployed.
