import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/widgets/travel_calendar_dialogs.dart';

/// Authenticate and associate a verified guest invitation. This never RSVPs;
/// callers must return to the selected stops for an explicit submission.
Future<bool> claimTravelAccount(
  BuildContext context,
  TravelCalendarService service,
) async {
  final auth = context.read<AuthProvider>();
  await auth.verifyToken();
  if (!context.mounted) return false;
  return await showDialog<bool>(
        context: context,
        barrierDismissible: false,
        builder: (_) => _TravelAccountClaimDialog(service: service),
      ) ??
      false;
}

class _TravelAccountClaimDialog extends StatefulWidget {
  const _TravelAccountClaimDialog({required this.service});
  final TravelCalendarService service;

  @override
  State<_TravelAccountClaimDialog> createState() =>
      _TravelAccountClaimDialogState();
}

class _TravelAccountClaimDialogState extends State<_TravelAccountClaimDialog> {
  final _phone = TextEditingController();
  final _code = TextEditingController();
  bool _codeSent = false;
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _phone.dispose();
    _code.dispose();
    super.dispose();
  }

  Future<void> _continue() async {
    final auth = context.read<AuthProvider>();
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      if (!auth.isAuthenticated) {
        if (!_codeSent) {
          final phone = _phone.text.trim();
          if (!RegExp(r'^\+[1-9]\d{7,14}$').hasMatch(phone)) {
            throw const TravelCalendarException(
              'Enter your phone number with its country code, such as +14155552671.',
            );
          }
          await auth.getVerificationCode(phone);
          if (mounted) setState(() => _codeSent = true);
          return;
        }
        if (!RegExp(r'^\d{6}$').hasMatch(_code.text.trim())) {
          throw const TravelCalendarException(
            'Enter the six-digit verification code.',
          );
        }
        await auth.logister(_phone.text.trim(), _code.text.trim());
        if (!auth.isAuthenticated) {
          throw const TravelCalendarException(
            'Please verify your phone number again.',
          );
        }
      }
      final accountId = auth.user?.id ?? '';
      if (await widget.service.guestNeedsClaim(accountId)) {
        if (!await widget.service.hasGuestSession) {
          if (!mounted) return;
          final verified = await showDialog<bool>(
            context: context,
            builder: (_) =>
                TravelGuestVerificationDialog(service: widget.service),
          );
          if (verified != true) return;
        }
        try {
          await widget.service.post('/calendar-guests/claim');
        } on TravelCalendarException catch (error) {
          if (error.statusCode != 401 || !mounted) rethrow;
          final verified = await showDialog<bool>(
            context: context,
            builder: (_) =>
                TravelGuestVerificationDialog(service: widget.service),
          );
          if (verified != true) return;
          await widget.service.post('/calendar-guests/claim');
        }
        await widget.service.markGuestClaimed(accountId);
      }
      if (mounted) Navigator.of(context).pop(true);
    } catch (error) {
      if (mounted) setState(() => _error = error.toString());
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = context.watch<AuthProvider>();
    final signedIn = auth.isAuthenticated;
    return PopScope(
      canPop: !_busy,
      child: AlertDialog(
        title: Text(signedIn ? 'Confirm your account' : 'Claim your account'),
        content: SizedBox(
          width: 380,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Your selected stops are saved. Verify your account, then review '
                  'your dates before submitting. Existing accounts are reused.',
                ),
                const SizedBox(height: 20),
                if (signedIn) ...[
                  Text(
                    'Continuing as ${auth.user?.username ?? auth.user?.name ?? auth.user?.phoneNumber ?? 'your account'}',
                  ),
                  TextButton(
                    onPressed: _busy
                        ? null
                        : () async {
                            await auth.logout();
                            auth.cancelVerificationCode();
                            if (mounted) {
                              setState(() {
                                _codeSent = false;
                                _code.clear();
                                _phone.clear();
                                _error = null;
                              });
                            }
                          },
                    child: const Text('Use another account'),
                  ),
                ],
                if (!signedIn) ...[
                  TextField(
                    controller: _phone,
                    enabled: !_busy && !_codeSent,
                    keyboardType: TextInputType.phone,
                    autofillHints: const [AutofillHints.telephoneNumber],
                    decoration: const InputDecoration(
                      labelText: 'Phone number',
                      hintText: '+14155552671',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  if (_codeSent) ...[
                    const SizedBox(height: 16),
                    TextField(
                      controller: _code,
                      enabled: !_busy,
                      keyboardType: TextInputType.number,
                      autofillHints: const [AutofillHints.oneTimeCode],
                      inputFormatters: [
                        FilteringTextInputFormatter.digitsOnly,
                        LengthLimitingTextInputFormatter(6),
                      ],
                      decoration: const InputDecoration(
                        labelText: 'Verification code',
                        border: OutlineInputBorder(),
                      ),
                      onSubmitted: (_) => _busy ? null : _continue(),
                    ),
                    TextButton(
                      onPressed: _busy
                          ? null
                          : () {
                              context
                                  .read<AuthProvider>()
                                  .cancelVerificationCode();
                              setState(() {
                                _codeSent = false;
                                _code.clear();
                              });
                            },
                      child: const Text('Change number or resend code'),
                    ),
                  ],
                ],
                if (_error != null) ...[
                  const SizedBox(height: 12),
                  Text(
                    _error!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: _busy ? null : () => Navigator.of(context).pop(false),
            child: const Text('Keep browsing'),
          ),
          FilledButton(
            onPressed: _busy ? null : _continue,
            child: Text(
              _busy
                  ? 'Please wait…'
                  : signedIn
                  ? 'Continue to review'
                  : _codeSent
                  ? 'Verify account'
                  : 'Send code',
            ),
          ),
        ],
      ),
    );
  }
}
