# Travel Angels: travel calendars, sharing, and participation

Status: initial calendar feature implemented in this workspace. See [setup and verification](CALENDAR_SETUP.md) for configuration, migration, and validation details. The defaults below describe the implemented first version.

**Objective.** Give each user an ongoing travel calendar where they can publish plans, control who sees each stop, broadcast relevant updates, and coordinate which portions friends will join. Receiving and browsing authorized plans must work without signing into a Travel Angels account or installing an app. Recording interest or joining requires a claimed account.

**Confirmed direction.** Each user has a general calendar rather than creating separate journeys. Individual stops have manageable viewing permissions. Friends who have not signed up must be supported throughout receiving and browsing plans.

**Initial release defaults.** One primary calendar per user. New calendars default to private; stops inherit their owning calendar's audience unless explicitly overridden. Joining requires host approval. The first delivery channel is email, alongside shareable browser links.

1. **Give every user an ongoing travel calendar.**

   A claimed account has a primary calendar containing its travel stops, with no required trip container, overall date range, or end date. Calendar presentation includes month navigation and a chronological agenda, destination/date filters, and past/current/upcoming plans. Support at least a full year spanning calendar years. Unscheduled or approximate plans appear in a clearly labeled agenda section rather than implying exact calendar dates.

   Each stop has a title, destination, date range or clearly labeled approximate/TBD dates, description, tentative/confirmed/cancelled planning status, viewing permissions, and whether friends can request to join. Owners can add, edit, and cancel stops as plans evolve. Overlaps and gaps are allowed. Travel dates are local destination dates so time-zone conversion does not shift the intended days. Joining coordinates attendance; it does not reserve accommodation or transportation.

2. **Support collaboration while keeping personal calendars.**

   Owners can grant calendar-wide editing access or invite a co-host for an individual stop. Calendar editors can see and manage every stop owned by that calendar, including private stops; this scope must be explicit when granting access. References to stops owned by another calendar retain their existing grants and confer no additional viewing or management access. Stop co-hosts can manage only their assigned stops. Managing a stop includes editing, publishing, broadcasting, and approving participation. Viewing alone grants none of these powers. The calendar owner manages calendar-wide access and can separately authorize an editor or co-host to manage a stop's viewing permissions.

   A shared stop can appear on both travelers' calendars after they accept co-hosting, with one owning calendar, one set of details, one viewing policy, and one participation list. Showing it on a second calendar does not copy it or expand its audience. Inherited permissions come from its owning calendar. The owner can remove editors/co-hosts; matching dates or destinations never automatically associate two independently created stops. Approved participants can see joined stops in their own calendar view as references, with the same permission checks.

3. **Manage calendar defaults and individual stop visibility.**

   Calendar owners set a default viewing audience: private, specific people, or anyone with the calendar share link. Specific people may be existing accounts or invited email/phone contacts without accounts. Each stop has the following options:

   | Stop setting | Who can browse it |
   | --- | --- |
   | Inherit calendar | The owning calendar's current default audience. |
   | Private | The owner, owning calendar's editors, and assigned stop co-hosts only. |
   | Specific people | Explicitly selected accounts/verified invited contacts, plus its managers. |
   | Anyone with the link | Anyone opening its share link or the calendar's shared view, plus its managers. |

   An explicit stop audience replaces the calendar's default viewing audience; it does not add to it. A broadly shared calendar cannot expose a private or restricted stop. Adding a calendar viewer does not grant access to stops with explicit restrictions. Changing the calendar default affects inherited stops, including future stops, but leaves explicit overrides unchanged. Management access is separate from viewer access and is displayed separately.

   Owners and authorized permission managers can add/remove viewers, change visibility, and preview the effective audience before confirming a change. The interface explains which stops are affected by a calendar-wide change. Confirmed permission changes on published content take effect immediately, independently of draft content edits. Draft stops remain visible only to their managers regardless of their intended published audience. Permission to view, subscription to updates, and permission to participate remain separate decisions.

4. **Separate drafting, publishing, and broadcasting.**

   New stops begin as drafts. Managers can preview the stop as an anonymous visitor or an intended recipient. Draft revisions do not replace published details until explicitly published. Publishing updates authorized calendar views; sending a general broadcast is a separate action with a preview of the message and eligible audience. Managers can unpublish a stop or close it to new participation. Material changes affecting existing participants have the notification behavior described below.

5. **Make guest browsing respect permissions without requiring signup.**

   Provide stable calendar and stop links that open in a mobile or desktop browser without mandatory app installation. Anyone-with-link content is browsable anonymously. For specific-people sharing, a recipient without an account can verify control of the invited email/phone using a fresh verification step and receive limited guest viewing access. This does not create or claim an app account or require signing into Travel Angels. Existing users can use their authenticated identity when it matches an authorized recipient.

   A calendar displays only published stops the current viewer may access, including when browsing someone else's co-hosted stops. Apply identical permissions to agenda/month views, direct links, filters, counts, previews, and responses. Hidden stops expose no title, destination, dates, or busy placeholder. Unauthorized direct links show a generic unavailable/access-required state without revealing whether a stop exists. An expired app session still permits anonymous content; restricted content can recover through guest verification without requiring signup.

   A forwarded anyone-with-link URL intentionally permits browsing. Forwarding a restricted invitation does not prove identity: the new viewer must verify an authorized contact or use their own authorized account. Guest access never permits claiming another recipient, editing, or submitting participation. If the app is installed, opening it is optional and preserves the destination.

6. **Make sharing and revocation clear.**

   Anyone-with-link sharing is unlisted, not restricted to named invitees. Shared calendars/stops are excluded from app-wide discovery and request exclusion from search indexing. Anonymous link previews contain only information approved for anonymous sharing; restricted links use generic previews. Guest views never expose recipient/participant lists, contact details, private travel documents, booking references, or unrelated account information.

   Owners and authorized permission managers can disable/replace share links and revoke specific recipients' viewing access within their scope. Old links and previously issued guest access must not bypass the current permissions. If someone still qualifies through another explicitly enabled access path, such as anyone-with-link sharing, the interface must explain that they retain access until that path is restricted too. Previously delivered messages and downloaded information cannot be recalled.

   When a sharing change removes someone's effective viewing permission, remove that stop from their calendar views, pause its subscriptions, and end future messages containing stop details. Active interest/requests/attendance become Removed and no longer count toward participation. Send only a minimal notice that access/participation ended, without restricted stop details. Preview which participation will end before a manager confirms the permission change. Restoring access does not automatically restore an RSVP or a paused subscription; the person explicitly opts back in. Expiring guest/app sessions do not revoke underlying grants or participation; unpublishing follows requirement 11.

7. **Invite people who have no account.**

   Managers can add recipients by the chosen delivery contact, with an optional name, and send an initial invitation within existing access permissions. Only calendar owners can grant calendar-default access; owners or authorized stop permission managers can grant access to selected stops. Sending an invitation cannot bypass these authorities or implicitly broaden access. The invitation explains the scope and links to the appropriate view; restricted invitations carry no stop details before contact verification.

   A pending recipient record represents an invitation awaiting a claim; it is not an authenticated account or participant. Managers can see invitation, following, claim, and delivery states without exposing them publicly. Repeated invitations from either host reuse the same recipient identity and scope. Inviting, granting access, subscribing, claiming, and joining are distinct actions; none automatically creates a friendship in My Circle.

8. **Let guests follow updates without claiming an account.**

   Guests can opt into updates for a calendar or selected stops after verifying their delivery contact. Contact verification alone does not claim an account. Following a calendar includes newly published stops only when the follower has viewing permission. Following selected stops excludes unrelated additions. Subscribing never grants access to otherwise hidden content. An initial invitation does not silently opt someone into recurring broadcasts.

   Guests can change preferences or unsubscribe through a scoped management link without logging in. Following expresses notification preference, not an RSVP. Subscription-management links do not grant restricted viewing, app account access, or participation permissions. If access is revoked, the behavior in requirement 6 applies.

9. **Let friends choose exactly what they want to join.**

   Guests can select one or multiple visible stops and, for fixed-date stops, the whole stop or a date subset within its range. Offer “Interested” for tentative interest and “Request to join” for host review. Both require a claimed, authenticated account before being recorded or shown to managers. Guests can make selections first, but see clearly that nothing has been submitted. TBD-date stops support interest; join requests require a concrete date range. Closed, cancelled, past, or unauthorized stops do not accept new requests.

   Participation states are Interested, Requested, Joined, Needs reconfirmation, Declined, Withdrawn, Removed, and Cancelled. Host approval changes Requested to Joined; claiming alone never does. Friends can revise dates or withdraw; changes to approved dates require renewed approval. Managers can approve, decline, or remove participants with a notification. Declined or withdrawn friends can request again while eligible; removed friends require a manager to restore eligibility first. Each person has one current participation record per stop, regardless of which calendar they found it on. Friends can manage their own participation without seeing everyone else's personal information.

10. **Define account claiming precisely.**

    Claiming means verifying ownership of the invited contact, creating or signing into the appropriate Travel Angels account, and associating pending invitations, viewing grants, and follow preferences with that account. Existing accounts must be reused rather than duplicated. If an invited contact already belongs to another account, require sign-in or recovery for that account; never silently transfer the contact or merge accounts. Account merging is outside the initial scope. Someone arriving from an anonymous share link can create/sign into an account and request to join an eligible stop without a prior invitation.

    Travel Angels currently authenticates by phone verification. Email invitations additionally require verified ownership of the invited email before attaching its recipient record; phone login alone does not establish email ownership. Guest viewing verification is not itself account claiming. The final authentication method is a design decision, but verified ownership is mandatory. A forwarded link, typed name, or unverified contact match cannot claim someone else's invitation.

    Preserve the calendar context, selected stops, dates, and intended action through claiming and recoverable verification failures. Request only identity/display-name information needed to participate; other profile completion can wait. Return to a review of selections and require explicit submission. Revalidate current permissions, dates, and eligibility before accepting it; preserve valid selections when another stop becomes unavailable. Replace expired verification links/codes without losing valid selections. Opening a message or automatic link scanning must not claim an account or submit participation.

11. **Broadcast only information each recipient may see.**

    Broadcasts contain a host-written message or change summary and links to current authorized plans. Managers can target eligible calendar followers or followers of selected stops. Each recipient must have both the relevant subscription and current viewing permission. Recheck permissions before delivery, including queued retries; omit inaccessible stops rather than sending one unrestricted calendar digest to everyone. Deduplicate updates for a shared stop even if someone follows both hosts. Stop-specific host text uses that stop's audience. Free-form text common to multiple stops may be sent only to recipients authorized for all included stops; otherwise managers prepare separate messages for the different audiences.

    Published date/destination changes and cancellations affecting requested/joined participants generate participation notices. Managers receive new-request and withdrawal notifications; friends receive approval, decline, removal, and cancellation notices. Every notice respects current access. Revocation uses the minimal notice in requirement 6. Distinguish general updates from participation notices in notification preferences and explain how withdrawing ends participation notices.

    Material changes retain prior selections and move Requested/Joined participation to Needs reconfirmation; these people are not counted as confirmed for the revised plan. Friends review and resubmit for approval. Cancellation sets affected participation to Cancelled and prevents joining; reopening never restores it automatically. Unpublishing notifies affected participants without exposing unpublished details and requires reconfirmation after republication. Notification links support authorized guest browsing; participation actions still require a claimed account.

12. **Handle access and delivery reliably.**

    Show whether broadcasts are pending, sent, or failed and support retrying failures without duplicating successful deliveries. Sending/opening a message is not invitation acceptance. Prevent duplicate identities, subscriptions, and participation from repeated clicks or invitations. Enforce access on the server as well as in the interface: unauthorized visitors cannot enumerate hidden stops/contacts, record participation, edit plans, or manage permissions. Invitation sending and verification need abuse controls. Calendar views and verification flows must work on small screens with keyboard navigation and accessible labels.

**Acceptance scenarios for the first release.**

- A user adds stops to an ongoing calendar spanning multiple years without creating a journey. Friends browse permitted stops in month and agenda views.
- An entirely new friend opens an anyone-with-link calendar/stop in a logged-out mobile browser without a mandatory login screen. Private and specific-people stops remain hidden.
- An invited friend verifies their contact and browses a restricted stop without claiming an account. An unauthorized person receiving the forwarded link cannot see it.
- A stop's specific-people override stays restricted when its calendar is broadly shared or its defaults change. A private stop is absent from views, counts, filters, and previews for ordinary viewers.
- One co-hosted stop appears on both travelers' calendars with consistent details, permissions, and participation. A viewer of the second calendar does not gain access through its defaults.
- A friend follows permitted stops, receives only relevant authorized updates, and unsubscribes without claiming an account. Following a calendar does not expose newly added private stops.
- A friend selects two stops and partial dates, claims/signs into one account, retains valid selections, and submits requests. Joining occurs only after approval.
- Access is revoked while a message is queued or a friend is claiming. No restricted details are delivered afterward; revoked selections cannot be submitted, and other valid selections remain.
- Revoking a participant's last viewing permission ends their participation with a generic notice. Restoring access does not silently restore attendance or subscriptions.
- Draft content stays unpublished. Permission managers can narrow access immediately without publishing unrelated draft edits. Ordinary viewers and unclaimed guests cannot change permissions or RSVP.
- Material plan changes require reconfirmation, and retrying a failed broadcast does not resend successful deliveries or duplicate shared-stop updates.

**Initial release scope.** One primary calendar per user, month/agenda views, stop management, calendar defaults and stop-level permissions, co-hosting, guest viewing, share links, one broadcast channel, guest subscriptions, invitation claiming, participation requests/approvals, and change notices. Multiple custom calendars per user, Google Calendar/Gmail integration, external calendar sync, booking, payments, cost splitting, group chat, maps, waitlists/capacity management, and automated itinerary imports are deferred. The calendar analogy describes the product organization; it does not require an external calendar integration.

**Success signals.** Measure authorized shared-page visits, claim completion/abandonment, submitted requests, approved participation, delivery failures, and unsubscribe rates. Do not require identifying anonymous visitors or treat email opens as proof a plan was viewed. Friends should discover relevant dates before signup and coordinate joining them while owners retain control over each stop's audience.

**Existing app context.** The current app has a global [login gate](../lib/widgets/home_widget.dart), [phone-based account access](../lib/services/auth_service.dart), and [friend invitations requiring an existing user](../lib/services/friend_service.dart). The [Travel Angels backend](../../../go/travel-angels/internal/server/server.go) protects existing friend/document routes. This feature adds permission-aware calendar browsing and invitation claiming; existing friendship invitations are not stop participation.
