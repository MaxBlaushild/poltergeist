# Travel Angels Trip Pages PRD

**Status:** Implemented and verified locally October 24, 2026. Production migration and deployment pending; see [release instructions](CALENDAR_SETUP.md).

**Purpose.** Give each travel plan a compelling page that friends can open, understand, and join. The page combines the invitation, people attending, places to stay, and flights in one destination. Hosts control its content, attendance, and sharing.

**Product model.** Each existing calendar stop becomes a trip with a dedicated page. Users keep their ongoing personal calendars; creating a trip does not require a parent journey. A trip has one canonical record even when it appears on several hosts’ or attendees’ calendars. Grouping several stops into a larger itinerary is a separate future capability. This is the proposed interpretation of “trip” for this release.

This PRD extends the [travel calendar requirements](TRAVEL_PLAN_SHARING_REQUIREMENTS.md). Existing guest browsing, account claiming, partial-date participation, and stop visibility remain foundational. It proposes changes to host authority and attendee-list visibility, and brings capacity management into scope.

**Experience reference.** Partiful combines an expressive event poster with a name, date, location, and description, then exposes hosting and RSVP controls from the event page. Travel Angels should adopt that clear invitation structure with its own visual design. [Partiful event creation](https://help.partiful.com/en-us/articles/15525299-creating-your-first-partiful-event)

Partiful also supports accepted co-hosts who manage event settings, configurable guest-list privacy, and attendance limits with waitlists. Travel Angels adapts those patterns to multi-day travel and existing account-claim requirements; the defaults below are Travel Angels product decisions. [Co-hosts](https://help.partiful.com/en-us/articles/15525429-how-do-i-make-someone-a-cohost-or-become-a-cohost), [event settings](https://help.partiful.com/en-us/articles/15525301-what-features-are-available-to-change-in-my-event-settings), [capacity](https://help.partiful.com/en-us/articles/15525379-how-can-i-limit-group-size)

**Primary users and outcomes.**

- Hosts can publish an appealing invitation, coordinate their group, and keep travel information current without maintaining separate documents and message threads.
- Friends can understand the destination, dates, group, and logistics before deciding which dates to join.
- New friends can browse authorized pages without an account, then claim an account when they request attendance.
- Confirmed attendees can return to one page for the current lodging, flights, and their attendance status.

**Proposed release defaults.**

| Decision | Default |
| --- | --- |
| Trip organization | One page per existing calendar stop. |
| Admission | Host approval required. |
| Capacity | Unlimited unless a host sets a positive maximum; counts distinct people attending, including attending hosts. |
| Partial dates | Supported; each person consumes one place regardless of duration. |
| Waitlist | Optional, with host approval for promotion. |
| Attendee names | Visible to hosts and confirmed attendees. |
| Attendee count | Visible to authorized trip viewers; hosts may hide it. |
| Confirmed logistics | Visible to hosts and confirmed attendees by default. |
| Host powers | Every accepted trip host has full trip management permissions. |
| Plus-ones | Each attendee joins through their own claimed account. |

1. **Open a full trip page from every entry point.**

   Tapping a trip in the month view, agenda, shared calendar, or personal joined-plans view opens the same dedicated page. Shared links open that page directly in a browser, with optional app opening. Existing stop links remain valid and resolve to the corresponding trip. Back navigation restores the originating calendar’s month, filters, and scroll position.

   On mobile, use a vertically arranged page with a prominent hero and a persistent attendance action. On desktop, use the main content column for the invitation and logistics, with a compact attendance summary alongside it. The primary navigation within the page is Overview, People, Lodging, and Flights. These sections remain directly accessible as content grows.

   The initial view communicates the trip name, destination, dates or date uncertainty, host identities, planning status, and the viewer’s available action. The page has clear states for loading, unavailable access, cancelled trips, closed attendance, full capacity, and pending approval.

2. **Present an expressive invitation with structured essentials.**

   | Content | Requirement |
   | --- | --- |
   | Name | Required, up to 200 characters. |
   | Hero image | One host-uploaded image with crop and focal-point controls, replacement/removal, and an accessible description. A designed fallback supports trips without an image. |
   | Destination | Required readable place name; distinguish the destination from a private lodging address. |
   | Dates | Fixed date range, approximate range, or TBD; show destination time zone where relevant. |
   | Description | Paragraphs, basic lists, and links for the experience, expectations, and practical details. |
   | Hosts | Names and avatars of accepted hosts, with a clear host label. Pending host invitations are private. |
   | Status | Tentative, confirmed, or cancelled, separately from whether attendance is open. |
   | Attendance | Current viewer status, join action, and permitted attendee/count information. |

   Hosts can draft and preview changes before publication. New trips require a name and destination; the image and logistics may follow later. Existing trips without new fields render gracefully. Image processing must preserve useful crops on mobile and desktop, remove embedded location metadata, and apply the trip’s access restrictions to the image itself. Restricted trips use generic external link previews.

3. **Give hosts complete control within their trip.**

   The owning calendar’s owner is the trip owner and initial host. A person becomes an additional host only after accepting a named invitation using a claimed account. Sharing an ordinary trip link never grants hosting rights. All accepted hosts can edit the invitation, lodging, and flights; publish or unpublish; manage visibility and share links; invite and remove non-owner hosts; manage attendance and capacity; send updates; and cancel or reopen the trip.

   Ownership adds only continuity controls: transferring ownership and permanently deleting the trip. Other hosts cannot remove the owner. A host can leave, but an owner must transfer ownership to an accepted host before leaving. Ownership transfer preserves explicit trip permissions; if calendar inheritance would change the audience, show and confirm the affected audience before completing it. Public host identities and hosting invitations update immediately when roles change.

   | Role | Browse | Manage own attendance | Edit trip and logistics | Manage audience and attendance | Manage hosts |
   | --- | --- | --- | --- | --- | --- |
   | Owner | All trip content | Yes | Yes | Yes | Yes, including ownership transfer |
   | Accepted host | All trip content | Yes | Yes | Yes | Invite or remove non-owner hosts |
   | Confirmed attendee | Published content allowed for attendees | Yes | No | No | No |
   | Other authorized account | Published content allowed for viewers | Yes, when eligible | No | No | No |
   | Authorized guest without an account | Published content allowed for viewers | Claim required | No | No | No |

   Host access does not grant access to another person’s unrelated calendar, personal documents, or bookings outside the trip. A host’s management role and whether they are traveling are separate: someone can help organize without attending.

   **Change from the current model:** calendar editing permission alone will not authorize editing a trip. Explicit trip hosts are the people who manage trip information. Before an existing trip adopts this model, its owner reviews existing calendar editors and stop collaborators and confirms who should receive full host invitations. The conversion preview identifies whose legacy management ends and any participation affected by losing access. Activation ends those legacy rights; selected people gain full host authority only after accepting the new invitation. Until conversion is confirmed, the trip retains its existing collaboration model. Newly created trips use the new model immediately.

4. **Preserve browsing before signup and make joining clear.**

   Trip visibility continues to support inherited calendar access, private, specific people, and anyone with the link. Explicit trip permissions replace calendar defaults. A reference on a different calendar never broadens access. Authorized anonymous visitors can browse link-shared trips; invited contacts can verify their contact to browse restricted trips without claiming an app account.

   The main attendance action is “Request to join.” Friends can choose the entire trip or dates within its fixed range, with a secondary “Interested” action. Approximate or TBD trips support interest until dates are fixed. Guests can prepare a selection before claiming; the claim flow preserves the trip and dates, then returns to explicit review and submission. Following updates remains a separate action.

   Requested, confirmed, waitlisted, interested, declined, withdrawn, removed, and needs-reconfirmation states must be understandable to the person affected. “Confirmed” is the page label for the existing Joined state. Approval never implies a flight ticket or lodging reservation. Friends can withdraw or request changed dates, and hosts receive the relevant update. Participation is unique per person and trip across all entry points.

5. **Show who is coming with a separate privacy control.**

   The People section shows confirmed attendees with display names, avatars, and host badges. Hosts can filter the management view by attendance status and inspect selected dates. Other viewers never see pending invitations, declined requests, contact-verification status, or other people’s private messages.

   Hosts choose whether attendee names are visible to confirmed attendees or hosts only. The default is confirmed attendees. Guests and other unconfirmed viewers can see the total confirmed count when enabled, but not attendee identities. A separate setting hides the numeric count; “Open,” “Full,” or “Waitlist available” may still communicate whether someone can apply. Exact remaining-place numbers must not appear when counts are hidden.

   Before submitting an RSVP, explain who can see the person’s name and avatar. Existing attendees remain identifiable only to hosts until they acknowledge the new roster audience; their attendance can still contribute to an allowed aggregate count. A broader roster audience also requires acknowledgement from affected attendees before exposing their identities. Selected attendance dates are visible to that person and hosts by default; sharing them with fellow attendees requires the person’s opt-in. Emails, phone numbers, booking identifiers, and unrelated profiles never appear on the attendee roster. Removing attendance or viewing access removes access to attendee-only information immediately.

6. **Enforce an understandable maximum attendee count.**

   Hosts can leave capacity unlimited or set a positive whole-number maximum. The maximum counts distinct confirmed people across the entire trip, including hosts marked as attending. The owner of a newly created trip starts as attending; existing trips retain their recorded attendance during migration. Accepting a co-host invitation includes choosing whether that host is attending. Organizers who are not attending consume no place. One person cannot count twice through multiple roles or invitations. Attending hosts follow the same cancellation and reservation lifecycle as other attendees.

   Partial attendance still consumes one place. For example, a maximum of eight means eight distinct confirmed travelers across the trip, even if four attend the first week and four the second. It does not represent daily occupancy, beds, or flight inventory. Label the setting “Maximum attendees across the whole trip” and explain this behavior beside it. Daily capacity is outside this release.

   Interested people, first-time pending requests, invitations, and waitlisted people do not reserve places. Approving attendance or marking a host as attending rechecks capacity. Concurrent approvals cannot exceed the limit. A host must raise the maximum before approving someone beyond it; there is no hidden override. Lowering the maximum below occupied or reserved places is blocked until the host resolves the affected attendance.

   Previously confirmed attendees who need to reconfirm after a material change keep a reserved place, including after resubmitting while awaiting renewed host approval. This also applies when an approved attendee requests different dates. They retain attendee-only information while the reservation and trip access remain valid so they can review the revised plans. Approval converts that reservation back to confirmed attendance without consuming a second place. Withdrawal, decline, removal, cancellation, or explicit host release ends the reservation. Previously pending requests receive no reservation. Show confirmed attendance and reserved places separately. There is no automatic expiration of reserved places in this release.

   When full, show “Full” or “Join waitlist,” depending on the setting. Waitlisting requires a claimed account and explicit submission. Hosts see requests in order of submission but choose whom to approve when a place opens. No automatic promotion occurs. Pending requests already submitted when capacity is reached remain pending, with a clear capacity message; hosts can invite the person to opt into the waitlist. Disabling the waitlist stops new entries and preserves existing people for resolution. Withdrawal, removal, and access revocation release occupied or reserved places.

7. **Make lodging a first-class trip section.**

   Hosts can add, edit, reorder, and remove multiple lodging entries. Each entry includes a property name, accommodation type, general location, check-in and checkout dates or an explicit TBD state, optional local times and time zone, description, listing or booking link, and planning status: considering, selected, booked, or cancelled. Optional indicative cost includes a currency and a clear basis such as per night, per person, or total.

   A lodging entry distinguishes “recommended option” from “group stay.” Hosts can describe available rooms or beds and associate confirmed attendees with a planned stay. Named stay assignments are visible only to hosts and the person assigned. This records the group’s plan; it does not reserve inventory, charge anyone, or automatically change trip attendance. Lodging occupancy is separate from the trip’s attendee maximum. Checkout dates are exclusive when calculating nights.

   Confirmed group stays are visible to hosts and confirmed attendees by default. A host may share a recommendation with everyone who can view the trip. Exact street addresses and private arrival instructions belong in a separately restricted detail area, defaulting to confirmed attendees. Host notes remain host-only. Dates outside the trip range trigger a review warning and can be kept when intentional, such as arriving the night before.

   Viewers see only entries and fields they are permitted to read. Hosts see an “Add lodging” prompt when empty. Viewers with no visible entries see the same neutral empty state whether there are hidden entries or no entries at all; this behavior also applies to flights.

8. **Make flights a first-class trip section.**

   Hosts can add multiple flight itineraries, including connecting legs. Each leg supports airline, flight number, departure and arrival airports, local departure and arrival date/time, and each airport’s time zone. Itineraries include a label, notes, optional booking/search link, and status: suggested, planned, booked, or cancelled. Suggested routes may omit a specific flight or time and must display that uncertainty.

   Group flights and recommended flights are visually distinct. Hosts can associate confirmed travelers with a planned flight, but those assignments do not claim to verify a ticket purchase. Traveler associations are visible only to hosts and the named traveler in this release. The shared route and timing can still be shown to confirmed attendees. Participants cannot edit the shared flight record.

   Show departure and arrival in their local time zones, including date changes and overnight connections. Validate chronological order using the actual instants when times are known; an arrival with an earlier local clock time must not be rejected solely for that reason. Flights before or after the trip dates are allowed after a host reviews the discrepancy. Flight updates are manual; live status, price tracking, and airline synchronization are deferred.

9. **Keep logistics permissions distinct from general trip visibility.**

   Each lodging or flight entry can be shared with all authorized trip viewers, confirmed attendees, or hosts only. Entry access always requires current trip access, and the narrower field rules above still apply. Hosts preview the page as a visitor and as an attendee before publishing logistics.

   Publishing a trip never publishes the host’s existing document library. This release uses structured information and external links. Reservation confirmation numbers, ticket numbers, passport information, payment details, boarding passes, and private booking documents are outside these shared records. Private document linking and participant-managed personal itineraries require a separate design.

   Permissions apply to page sections, APIs, thumbnails, image URLs, notifications, and cached views. Counts and previews cannot reveal hidden records. Narrowing access takes effect immediately, including while another viewer has the page open or an update is queued. Previously downloaded content cannot be recalled.

10. **Keep changes predictable for hosts and attendees.**

    Invitation content and logistics retain draft and published versions. Permission changes, attendance decisions, and closing participation apply immediately and clearly indicate their effects. A draft hero or lodging edit must not delay urgent access revocation.

    Publishing changed trip dates, destination, or cancellation uses the existing participation-notification and reconfirmation rules. Lodging or flight changes notify the affected confirmed travelers according to their current access, without automatically resetting their trip RSVP. Cosmetic edits do not require reconfirmation. Share links always resolve to the current published version.

    Host actions record who changed content, permissions, hosts, capacity, and participation. Concurrent editing must surface a conflict instead of silently overwriting a newer change. Host removal immediately ends management authority; any separate viewer or attendee eligibility is evaluated and shown explicitly. Cancellation closes attendance and preserves the plan for authorized review. Reopening does not silently restore previous attendance or bookings.

11. **Make the page usable across devices and access states.**

    Mobile browser and app experiences support the same essential browse, claim, and attendance flows. Hosts can manage trips from mobile and desktop. The hero must not push essential dates and the join action beyond an excessive initial scroll. Images load progressively, maintain layout space, and use appropriately sized assets.

    Names, statuses, and actions remain readable over imagery and without relying on color. Support screen readers, keyboard navigation, large text, clear focus states, and reduced motion. Failed uploads preserve the prior published hero. Failed saves retain drafts and show a retry path. Failed claims preserve valid selections. Private data is cleared when identity or permissions change.

12. **Launch without breaking existing trips.**

    Existing titles, descriptions, dates, participation, grants, subscriptions, and links retain their meaning. Older trips start with the image fallback, unlimited capacity, no logistics entries, and the privacy defaults above. Attendee identities are not exposed to anonymous visitors during migration.

    Host-role conversion is a visible owner action because the new role includes permission management. Until conversion, current collaborators retain their existing rights rather than being silently reclassified. Release acceptance must cover both existing and newly created trips and the transition between the two management models.

**Acceptance scenarios.**

- A logged-out friend taps a calendar item and reaches its trip page with a hero, name, dates, destination, description, and hosts. A restricted trip requires contact verification but not account claiming to browse.
- The same trip opened from either host’s calendar or an old stop link has the same content, permissions, attendance, and logistics.
- A guest selects partial dates, claims an account, returns to review, and submits a request without losing the selection or becoming confirmed automatically.
- An accepted host edits the hero and flights, publishes the changes, changes viewing permissions, and manages attendance. An attendee attempting the same changes is denied.
- A non-attending organizer consumes no place. An attending host counts once. With one place remaining, two concurrent approvals produce only one confirmation.
- A trip capped at eight cannot confirm a ninth distinct attendee, even if that person selects dates that do not overlap the first eight.
- A confirmed person awaiting reconfirmation retains a reserved place through resubmission and renewed host review. A previously pending request receives no reservation. Hosts can see and explicitly release reservations.
- A full trip offers its optional waitlist, and a vacancy does not confirm anyone without a host decision. Disabling waitlisting preserves existing entries.
- Anonymous visitors can see an allowed count but not attendee identities. Confirmed attendees can see the roster under the default privacy setting; existing attendees’ names remain hidden until they acknowledge the new audience. Hiding the count also removes exact remaining-place numbers.
- A public lodging recommendation is visible to a guest while a confirmed stay’s address remains unavailable. A private booking document already in the app never appears automatically.
- A connecting flight crossing time zones and the date line shows correct local dates and an intelligible itinerary. Its named travelers remain private to hosts and the people named.
- Revoking access removes protected images, attendee information, and logistics from subsequent responses and queued messages. Attendee removal ends attendee-only access even if general trip viewing remains allowed.
- A host previews draft content while friends still see the published version. Cancellation and material changes produce the correct attendance states and notifications.
- An existing limited collaborator gains full host powers only through the explicit conversion and acceptance process; existing collaboration remains usable until then.

**Release scope.** Include canonical trip pages, hero uploads, structured essentials, host management, attendee privacy, capacity and host-managed waitlists, structured lodging and flights, guest claim continuity, and permission-aware updates. Defer nested multi-stop trips, daily capacity, unnamed plus-ones, automatic waitlist promotion, room allocation, booking and payments, cost splitting, live flight data, document imports, attendee-authored itineraries, chat, and social photo feeds.

**Success measures.** Track trip publication and sharing, authorized page visits, claim completion, request submission, approval and waitlist conversion, repeat attendee visits, and use of lodging/flight sections. Measure claim abandonment and host effort to publish and approve participants. Launch requires passing the access-control and capacity scenarios above, with no exposure of restricted attendee or booking information. Anonymous analytics must not require account creation or identify private travel itineraries.

**Approved product decisions.** A trip maps to one existing stop; the maximum counts unique people across the whole trip, including attending hosts; and calendar editors must explicitly become trip hosts. These decisions preserve the ongoing-calendar model.
