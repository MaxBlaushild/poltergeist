import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/user.dart';
import '../providers/auth_provider.dart';
import '../providers/party_provider.dart';
import '../services/admin_service.dart';
import '../services/poi_service.dart';
import '../models/zone.dart';
import '../models/zone_seed_job.dart';

class AdminScreen extends StatefulWidget {
  const AdminScreen({super.key});

  @override
  State<AdminScreen> createState() => _AdminScreenState();
}

class _AdminScreenState extends State<AdminScreen> {
  final _teamIdController = TextEditingController();
  final _pointOfInterestIdController = TextEditingController();
  final _quantityController = TextEditingController();
  final _partyInviteeIdController = TextEditingController();
  final _seedPlaceCountController = TextEditingController(text: '8');
  final _seedCharacterCountController = TextEditingController(text: '4');
  final _seedQuestCountController = TextEditingController(text: '4');
  final _seedInputEncounterCountController = TextEditingController(text: '0');
  final _seedOptionEncounterCountController = TextEditingController(text: '0');
  bool _unlockLoading = false;
  bool _captureLoading = false;
  bool _zonesLoading = false;
  bool _seedJobsLoading = false;
  bool _seedCreating = false;
  String? _error;
  String? _success;
  String? _partyError;
  String? _partySuccess;
  String? _seedError;
  String? _seedSuccess;
  List<Zone> _zones = [];
  Zone? _selectedZone;
  List<ZoneSeedJob> _seedJobs = [];
  final Set<String> _partyBusy = {};
  final Set<String> _seedBusy = {};

  @override
  void dispose() {
    _teamIdController.dispose();
    _pointOfInterestIdController.dispose();
    _quantityController.dispose();
    _partyInviteeIdController.dispose();
    _seedPlaceCountController.dispose();
    _seedCharacterCountController.dispose();
    _seedQuestCountController.dispose();
    _seedInputEncounterCountController.dispose();
    _seedOptionEncounterCountController.dispose();
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<PartyProvider>().refresh();
      _loadZones();
    });
  }

  User _stubUser(String id) {
    return User(
      id: id,
      phoneNumber: '',
      name: '',
      username: '',
      profilePictureUrl: '',
    );
  }

  Future<void> _runPartyAction(
    String key,
    Future<void> Function() action, {
    String? successMessage,
  }) async {
    if (_partyBusy.contains(key)) return;
    setState(() {
      _partyError = null;
      _partySuccess = null;
      _partyBusy.add(key);
    });
    try {
      await action();
      if (successMessage != null && mounted) {
        setState(() => _partySuccess = successMessage);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _partyError = e.toString());
      }
    } finally {
      if (mounted) {
        setState(() => _partyBusy.remove(key));
      }
    }
  }

  Future<void> _loadZones() async {
    if (_zonesLoading) return;
    setState(() {
      _zonesLoading = true;
      _seedError = null;
    });
    try {
      final zones = await context.read<PoiService>().getZones();
      zones.sort((a, b) => a.name.compareTo(b.name));
      if (!mounted) return;
      setState(() {
        _zones = zones;
        _selectedZone ??= zones.isNotEmpty ? zones.first : null;
      });
      await _loadSeedJobs();
    } catch (e) {
      if (!mounted) return;
      setState(() => _seedError = e.toString());
    } finally {
      if (mounted) {
        setState(() => _zonesLoading = false);
      }
    }
  }

  Future<void> _loadSeedJobs() async {
    if (_seedJobsLoading) return;
    if (_selectedZone == null) {
      setState(() => _seedJobs = []);
      return;
    }
    setState(() {
      _seedJobsLoading = true;
      _seedError = null;
    });
    try {
      final jobs = await context.read<AdminService>().getZoneSeedJobs(
        zoneId: _selectedZone!.id,
        limit: 25,
      );
      if (!mounted) return;
      setState(() => _seedJobs = jobs);
    } catch (e) {
      if (mounted) setState(() => _seedError = e.toString());
    } finally {
      if (mounted) setState(() => _seedJobsLoading = false);
    }
  }

  Future<void> _createSeedDraft() async {
    final zone = _selectedZone;
    if (zone == null || _seedCreating) return;
    final placeCount = int.tryParse(_seedPlaceCountController.text.trim());
    final characterCount = int.tryParse(
      _seedCharacterCountController.text.trim(),
    );
    final questCount = int.tryParse(_seedQuestCountController.text.trim());
    final inputEncounterCount = int.tryParse(
      _seedInputEncounterCountController.text.trim(),
    );
    final optionEncounterCount = int.tryParse(
      _seedOptionEncounterCountController.text.trim(),
    );
    if (placeCount == null ||
        characterCount == null ||
        questCount == null ||
        inputEncounterCount == null ||
        optionEncounterCount == null) {
      setState(() => _seedError = 'Counts must be integers.');
      return;
    }
    setState(() {
      _seedCreating = true;
      _seedError = null;
      _seedSuccess = null;
    });
    try {
      await context.read<AdminService>().seedZoneDraft(
        zoneId: zone.id,
        placeCount: placeCount,
        characterCount: characterCount,
        questCount: questCount,
        inputEncounterCount: inputEncounterCount,
        optionEncounterCount: optionEncounterCount,
      );
      if (!mounted) return;
      setState(() => _seedSuccess = 'Draft queued.');
      await _loadSeedJobs();
    } catch (e) {
      if (mounted) setState(() => _seedError = e.toString());
    } finally {
      if (mounted) setState(() => _seedCreating = false);
    }
  }

  Future<void> _approveSeedJob(ZoneSeedJob job) async {
    if (_seedBusy.contains(job.id)) return;
    setState(() {
      _seedBusy.add(job.id);
      _seedError = null;
      _seedSuccess = null;
    });
    try {
      await context.read<AdminService>().approveZoneSeedJob(job.id);
      if (!mounted) return;
      setState(() => _seedSuccess = 'Job approved and applying.');
      await _loadSeedJobs();
    } catch (e) {
      if (mounted) setState(() => _seedError = e.toString());
    } finally {
      if (mounted) setState(() => _seedBusy.remove(job.id));
    }
  }

  Future<void> _unlock() async {
    final teamId = _teamIdController.text.trim();
    final poiId = _pointOfInterestIdController.text.trim();
    if (teamId.isEmpty || poiId.isEmpty) return;
    setState(() {
      _error = null;
      _success = null;
      _unlockLoading = true;
    });
    try {
      await context.read<AdminService>().unlockPointOfInterestForTeam(
        teamId: teamId,
        pointOfInterestId: poiId,
      );
      if (mounted) {
        setState(() {
          _unlockLoading = false;
          _success = 'Unlocked successfully.';
        });
        _teamIdController.clear();
        _pointOfInterestIdController.clear();
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _unlockLoading = false;
          _error = e.toString();
        });
      }
    }
  }

  Future<void> _capture() async {
    final teamId = _teamIdController.text.trim();
    final poiId = _pointOfInterestIdController.text.trim();
    final q = _quantityController.text.trim();
    if (teamId.isEmpty || poiId.isEmpty || q.isEmpty) return;
    final tier = int.tryParse(q);
    if (tier == null) {
      setState(() => _error = 'Quantity must be an integer (tier).');
      return;
    }
    setState(() {
      _error = null;
      _success = null;
      _captureLoading = true;
    });
    try {
      await context.read<AdminService>().capturePointOfInterestForTeam(
        teamId: teamId,
        pointOfInterestId: poiId,
        tier: tier,
      );
      if (mounted) {
        setState(() {
          _captureLoading = false;
          _success = 'Capture successful.';
        });
        _teamIdController.clear();
        _pointOfInterestIdController.clear();
        _quantityController.clear();
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _captureLoading = false;
          _error = e.toString();
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 3,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Admin'),
          bottom: const TabBar(
            tabs: [
              Tab(text: 'Team & POI'),
              Tab(text: 'Parties'),
              Tab(text: 'Zone Seeding'),
            ],
          ),
        ),
        body: TabBarView(
          children: [
            _buildTeamPoiTab(context),
            _buildPartyTab(context),
            _buildZoneSeedTab(context),
          ],
        ),
      ),
    );
  }

  Widget _buildTeamPoiTab(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (_error != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Text(
                _error!,
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
            ),
          if (_success != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Text(
                _success!,
                style: TextStyle(color: Colors.green.shade700),
              ),
            ),
          Text(
            'Team & POI',
            style: Theme.of(
              context,
            ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
          ),
          const Divider(),
          const SizedBox(height: 8),
          TextField(
            controller: _teamIdController,
            decoration: const InputDecoration(
              labelText: 'Team ID',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _pointOfInterestIdController,
            decoration: const InputDecoration(
              labelText: 'Point of Interest ID',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 24),
          Text(
            'Unlock point for team',
            style: Theme.of(
              context,
            ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
          ),
          const Divider(),
          const SizedBox(height: 8),
          FilledButton(
            onPressed:
                _unlockLoading ||
                    _teamIdController.text.trim().isEmpty ||
                    _pointOfInterestIdController.text.trim().isEmpty
                ? null
                : _unlock,
            child: Text(_unlockLoading ? 'Unlocking…' : 'Unlock'),
          ),
          const SizedBox(height: 24),
          Text(
            'Capture for team',
            style: Theme.of(
              context,
            ).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
          ),
          const Divider(),
          const SizedBox(height: 8),
          TextField(
            controller: _quantityController,
            decoration: const InputDecoration(
              labelText: 'Quantity (tier)',
              border: OutlineInputBorder(),
            ),
            keyboardType: TextInputType.number,
          ),
          const SizedBox(height: 12),
          FilledButton(
            onPressed:
                _captureLoading ||
                    _teamIdController.text.trim().isEmpty ||
                    _pointOfInterestIdController.text.trim().isEmpty ||
                    _quantityController.text.trim().isEmpty
                ? null
                : _capture,
            child: Text(_captureLoading ? 'Capturing…' : 'Capture'),
          ),
        ],
      ),
    );
  }

  Widget _buildPartyTab(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    final partyProvider = context.watch<PartyProvider>();
    final party = partyProvider.party;
    final invites = partyProvider.partyInvites;
    final isLeader = party?.leaderId == auth.user?.id;
    final theme = Theme.of(context);

    if (auth.user == null) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Text('Log in to manage parties.'),
        ),
      );
    }

    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Text(
                'Party Tools',
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const Spacer(),
              OutlinedButton.icon(
                onPressed: _partyBusy.contains('refresh')
                    ? null
                    : () => _runPartyAction(
                        'refresh',
                        () => context.read<PartyProvider>().refresh(),
                        successMessage: 'Party data refreshed.',
                      ),
                icon: const Icon(Icons.refresh, size: 18),
                label: Text(
                  _partyBusy.contains('refresh') ? 'Refreshing…' : 'Refresh',
                ),
              ),
            ],
          ),
          const Divider(),
          if (_partyError != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Text(
                _partyError!,
                style: TextStyle(color: theme.colorScheme.error),
              ),
            ),
          if (_partySuccess != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Text(
                _partySuccess!,
                style: TextStyle(color: Colors.green.shade700),
              ),
            ),
          Text(
            'Current Party',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const Divider(),
          if (partyProvider.loading)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 16),
              child: Center(child: CircularProgressIndicator()),
            )
          else if (party == null)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 8),
              child: Text('No active party. Send an invite to create one.'),
            )
          else ...[
            Text('Party ID: ${party.id}'),
            const SizedBox(height: 8),
            Text(
              'Leader: ${party.leader.username.isNotEmpty ? party.leader.username : party.leader.name}',
            ),
            const SizedBox(height: 12),
            Text(
              'Members',
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 8),
            ...party.members.map((member) {
              final isMemberLeader = member.id == party.leaderId;
              return Container(
                margin: const EdgeInsets.only(bottom: 8),
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        member.username.isNotEmpty
                            ? member.username
                            : member.name.isNotEmpty
                            ? member.name
                            : member.id,
                      ),
                    ),
                    if (isMemberLeader)
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: theme.colorScheme.tertiary,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          'LEADER',
                          style: theme.textTheme.labelSmall?.copyWith(
                            fontWeight: FontWeight.w700,
                            color: theme.colorScheme.onTertiary,
                          ),
                        ),
                      )
                    else if (isLeader)
                      TextButton(
                        onPressed: _partyBusy.contains('promote:${member.id}')
                            ? null
                            : () => _runPartyAction(
                                'promote:${member.id}',
                                () => context.read<PartyProvider>().setLeader(
                                  member,
                                ),
                                successMessage: 'Leader updated.',
                              ),
                        child: Text(
                          _partyBusy.contains('promote:${member.id}')
                              ? 'Promoting…'
                              : 'Make leader',
                        ),
                      ),
                  ],
                ),
              );
            }),
            const SizedBox(height: 8),
            FilledButton.icon(
              onPressed: _partyBusy.contains('leave')
                  ? null
                  : () => _runPartyAction(
                      'leave',
                      () => context.read<PartyProvider>().leaveParty(),
                      successMessage: 'Left party.',
                    ),
              icon: const Icon(Icons.logout, size: 18),
              label: Text(
                _partyBusy.contains('leave') ? 'Leaving…' : 'Leave party',
              ),
              style: FilledButton.styleFrom(
                backgroundColor: theme.colorScheme.error,
                foregroundColor: theme.colorScheme.onError,
              ),
            ),
          ],
          const SizedBox(height: 24),
          Text(
            'Invite User',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const Divider(),
          TextField(
            controller: _partyInviteeIdController,
            decoration: const InputDecoration(
              labelText: 'Invitee user ID',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          FilledButton(
            onPressed:
                _partyBusy.contains('invite') ||
                    _partyInviteeIdController.text.trim().isEmpty
                ? null
                : () {
                    final id = _partyInviteeIdController.text.trim();
                    _runPartyAction(
                      'invite',
                      () => context.read<PartyProvider>().inviteToParty(
                        _stubUser(id),
                      ),
                      successMessage: 'Invite sent.',
                    ).then((_) {
                      if (mounted) _partyInviteeIdController.clear();
                    });
                  },
            child: Text(
              _partyBusy.contains('invite') ? 'Sending…' : 'Send invite',
            ),
          ),
          const SizedBox(height: 24),
          Text(
            'Party Invites',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const Divider(),
          if (invites.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 8),
              child: Text('No pending party invites.'),
            )
          else
            ...invites.map((invite) {
              final inviteeName = invite.invitee.username.isNotEmpty
                  ? invite.invitee.username
                  : invite.invitee.name.isNotEmpty
                  ? invite.invitee.name
                  : invite.inviteeId;
              final inviterName = invite.inviter.username.isNotEmpty
                  ? invite.inviter.username
                  : invite.inviter.name.isNotEmpty
                  ? invite.inviter.name
                  : invite.inviterId;
              final isInvitee = invite.inviteeId == auth.user?.id;
              return Container(
                margin: const EdgeInsets.only(bottom: 8),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Inviter: $inviterName'),
                          Text('Invitee: $inviteeName'),
                          Text(
                            'Invite ID: ${invite.id}',
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ),
                    ),
                    if (isInvitee) ...[
                      TextButton(
                        onPressed: _partyBusy.contains('accept:${invite.id}')
                            ? null
                            : () => _runPartyAction(
                                'accept:${invite.id}',
                                () => context
                                    .read<PartyProvider>()
                                    .acceptPartyInvite(invite.id),
                                successMessage: 'Invite accepted.',
                              ),
                        child: Text(
                          _partyBusy.contains('accept:${invite.id}')
                              ? 'Accepting…'
                              : 'Accept',
                        ),
                      ),
                      TextButton(
                        onPressed: _partyBusy.contains('reject:${invite.id}')
                            ? null
                            : () => _runPartyAction(
                                'reject:${invite.id}',
                                () => context
                                    .read<PartyProvider>()
                                    .rejectPartyInvite(invite.id),
                                successMessage: 'Invite rejected.',
                              ),
                        child: Text(
                          _partyBusy.contains('reject:${invite.id}')
                              ? 'Rejecting…'
                              : 'Reject',
                        ),
                      ),
                    ] else
                      Text(
                        'Pending',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                  ],
                ),
              );
            }),
        ],
      ),
    );
  }

  Widget _buildZoneSeedTab(BuildContext context) {
    final theme = Theme.of(context);
    final selectedZone = _selectedZone;

    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Text(
                'Zone Seeding',
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const Spacer(),
              OutlinedButton.icon(
                onPressed: _zonesLoading ? null : _loadZones,
                icon: const Icon(Icons.refresh, size: 18),
                label: Text(_zonesLoading ? 'Refreshing...' : 'Refresh zones'),
              ),
            ],
          ),
          const Divider(),
          if (_seedError != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Text(
                _seedError!,
                style: TextStyle(color: theme.colorScheme.error),
              ),
            ),
          if (_seedSuccess != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Text(
                _seedSuccess!,
                style: TextStyle(color: Colors.green.shade700),
              ),
            ),
          DropdownButtonFormField<Zone>(
            value: selectedZone,
            items: _zones
                .map(
                  (zone) =>
                      DropdownMenuItem(value: zone, child: Text(zone.name)),
                )
                .toList(),
            onChanged: (value) {
              if (value == null) return;
              setState(() => _selectedZone = value);
              _loadSeedJobs();
            },
            decoration: const InputDecoration(
              labelText: 'Zone',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 16),
          Text(
            'Draft settings',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const Divider(),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _seedPlaceCountController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Places',
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: TextField(
                  controller: _seedCharacterCountController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Characters',
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: TextField(
                  controller: _seedQuestCountController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Quests',
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _seedInputEncounterCountController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Input encounters',
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: TextField(
                  controller: _seedOptionEncounterCountController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    labelText: 'Option encounters',
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              FilledButton(
                onPressed: _seedCreating || selectedZone == null
                    ? null
                    : _createSeedDraft,
                child: Text(_seedCreating ? 'Queuing...' : 'Create draft'),
              ),
              const SizedBox(width: 12),
              OutlinedButton.icon(
                onPressed: _seedJobsLoading ? null : _loadSeedJobs,
                icon: const Icon(Icons.refresh, size: 18),
                label: Text(_seedJobsLoading ? 'Loading...' : 'Refresh drafts'),
              ),
            ],
          ),
          const SizedBox(height: 24),
          Text(
            'Draft jobs',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const Divider(),
          if (_seedJobsLoading)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 16),
              child: Center(child: CircularProgressIndicator()),
            )
          else if (_seedJobs.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 8),
              child: Text('No draft jobs yet.'),
            )
          else
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemBuilder: (context, index) {
                final job = _seedJobs[index];
                return _buildSeedJobCard(context, job);
              },
              separatorBuilder: (_, __) => const SizedBox(height: 12),
              itemCount: _seedJobs.length,
            ),
        ],
      ),
    );
  }

  Widget _buildSeedJobCard(BuildContext context, ZoneSeedJob job) {
    final theme = Theme.of(context);
    final draft = job.draft;
    final statusColor = _statusColor(theme, job.status);
    final created = job.createdAt?.toLocal().toString();
    final updated = job.updatedAt?.toLocal().toString();
    final isAwaitingApproval = job.status == 'awaiting_approval';

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  'Job ${_shortId(job.id)}',
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: statusColor,
                  borderRadius: BorderRadius.circular(999),
                ),
                child: Text(
                  job.status.replaceAll('_', ' '),
                  style: theme.textTheme.labelSmall?.copyWith(
                    color: theme.colorScheme.onPrimary,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            'Counts: ${job.placeCount} places · ${job.characterCount} characters · ${job.questCount} quests · ${job.inputEncounterCount} input encounters · ${job.optionEncounterCount} option encounters',
          ),
          if (created != null) Text('Created: $created'),
          if (updated != null) Text('Updated: $updated'),
          if (job.errorMessage != null && job.errorMessage!.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Text(
                job.errorMessage!,
                style: TextStyle(color: theme.colorScheme.error),
              ),
            ),
          if (draft != null) ...[
            const SizedBox(height: 12),
            Text(
              draft.fantasyName ?? 'Draft',
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 6),
            if (draft.zoneDescription != null &&
                draft.zoneDescription!.isNotEmpty)
              Text(
                draft.zoneDescription!,
                maxLines: 4,
                overflow: TextOverflow.ellipsis,
              ),
            const SizedBox(height: 8),
            ExpansionTile(
              tilePadding: EdgeInsets.zero,
              title: const Text('Draft details'),
              children: [
                _buildDraftSection(
                  context,
                  title: 'Points of interest',
                  items: draft.pointsOfInterest
                      .map((poi) => '${poi.name} · ${poi.placeId}')
                      .toList(),
                ),
                _buildDraftSection(
                  context,
                  title: 'Characters',
                  items: draft.characters
                      .map((c) => '${c.name} · ${c.placeId}')
                      .toList(),
                ),
                _buildDraftSection(
                  context,
                  title: 'Quests',
                  items: draft.quests
                      .map(
                        (q) =>
                            '${q.name} · giver ${_shortId(q.questGiverDraftId)}',
                      )
                      .toList(),
                ),
              ],
            ),
          ],
          if (isAwaitingApproval)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: FilledButton(
                onPressed: _seedBusy.contains(job.id)
                    ? null
                    : () => _approveSeedJob(job),
                child: Text(
                  _seedBusy.contains(job.id)
                      ? 'Approving...'
                      : 'Approve & apply',
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildDraftSection(
    BuildContext context, {
    required String title,
    required List<String> items,
  }) {
    final theme = Theme.of(context);
    if (items.isEmpty) {
      return Padding(
        padding: const EdgeInsets.only(bottom: 12),
        child: Text('$title: none'),
      );
    }
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: theme.textTheme.labelLarge?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 6),
          ...items.take(6).map((item) => Text('- $item')),
          if (items.length > 6)
            Text(
              '... and ${items.length - 6} more',
              style: theme.textTheme.bodySmall,
            ),
        ],
      ),
    );
  }

  String _shortId(String id) {
    if (id.length <= 8) return id;
    return id.substring(0, 8);
  }

  Color _statusColor(ThemeData theme, String status) {
    switch (status) {
      case 'queued':
        return theme.colorScheme.secondary;
      case 'in_progress':
        return theme.colorScheme.tertiary;
      case 'awaiting_approval':
        return theme.colorScheme.primary;
      case 'approved':
        return theme.colorScheme.primary;
      case 'applying':
        return theme.colorScheme.tertiary;
      case 'applied':
        return Colors.green.shade700;
      case 'failed':
        return theme.colorScheme.error;
      default:
        return theme.colorScheme.outline;
    }
  }
}
