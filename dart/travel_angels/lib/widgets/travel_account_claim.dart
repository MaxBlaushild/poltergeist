import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/travel_calendar_service.dart';
import 'package:travel_angels/utils/auth_phone_number.dart';
import 'package:travel_angels/widgets/international_phone_field.dart';
import 'package:travel_angels/widgets/travel_calendar_dialogs.dart';
import 'package:travel_angels/widgets/travel_calendar_views.dart';

/// Authenticate and associate a verified guest invitation. This never RSVPs;
/// callers must return to the selected stops for an explicit submission.
Future<bool> claimTravelAccount(
  BuildContext context,
  TravelCalendarService service,
) async {
  final auth = context.read<AuthProvider>();
  await auth.verifyToken();
  if (!context.mounted) return false;
  auth.cancelVerificationCode();
  final claimed =
      await showDialog<bool>(
        context: context,
        barrierDismissible: false,
        builder: (_) => _TravelAccountClaimDialog(service: service),
      ) ??
      false;
  // System Back can dismiss this route too. Never leave a hidden destination
  // or test-account state for a subsequent sign-in screen to inherit.
  if (context.mounted) auth.cancelVerificationCode();
  return claimed;
}

class _TravelAccountClaimDialog extends StatefulWidget {
  const _TravelAccountClaimDialog({required this.service});
  final TravelCalendarService service;

  @override
  State<_TravelAccountClaimDialog> createState() =>
      _TravelAccountClaimDialogState();
}

class _TravelAccountClaimDialogState extends State<_TravelAccountClaimDialog> {
  String? _phoneNumber;
  int _phoneInputRevision = 0;
  final _code = TextEditingController();
  bool _codeSent = false;
  bool _busy = false;
  String? _error;

  @override
  void dispose() {
    _code.dispose();
    super.dispose();
  }

  Future<void> _continue() async {
    if (_busy) return;
    final auth = context.read<AuthProvider>();
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      if (!auth.isAuthenticated) {
        if (!_codeSent) {
          final phone = _phoneNumber;
          if (phone == null) {
            throw const TravelCalendarException(
              'Enter a valid phone number for the selected country.',
            );
          }
          await auth.getVerificationCode(phone);
          if (mounted) {
            setState(() => _codeSent = auth.isWaitingForVerificationCode);
          }
          return;
        }
        if (!RegExp(r'^\d{6}$').hasMatch(_code.text.trim())) {
          throw const TravelCalendarException(
            'Enter the six-digit verification code.',
          );
        }
        await auth.logister(
          auth.verificationPhoneNumber ?? _phoneNumber ?? '',
          _code.text.trim(),
        );
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
      if (mounted) {
        setState(
          () => _error = error is DioException
              ? auth.error ?? 'Unable to verify your account. Please try again.'
              : error.toString(),
        );
      }
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
        title: TravelCalendarDialogTitle(
          signedIn ? 'Confirm your account' : 'Claim your account',
          label: 'Your part of the plan',
        ),
        content: SizedBox(
          width: 380,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const TravelCalendarCallout(
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
                                _phoneNumber = null;
                                _phoneInputRevision++;
                                _error = null;
                              });
                            }
                          },
                    child: const Text('Use another account'),
                  ),
                ],
                if (!signedIn) ...[
                  InternationalPhoneField(
                    key: ValueKey(_phoneInputRevision),
                    enabled: !_busy && !_codeSent,
                    onChanged: (value) => setState(() => _phoneNumber = value),
                  ),
                  if (_codeSent) ...[
                    const SizedBox(height: 16),
                    if (auth.isTestVerification) ...[
                      const Text(
                        'Test account: enter your configured 6-digit test code. No text message will be sent.',
                      ),
                      const SizedBox(height: 12),
                    ],
                    TextField(
                      controller: _code,
                      enabled: !_busy,
                      keyboardType: TextInputType.number,
                      autofillHints: auth.isTestVerification
                          ? null
                          : const [AutofillHints.oneTimeCode],
                      inputFormatters: [
                        FilteringTextInputFormatter.digitsOnly,
                        LengthLimitingTextInputFormatter(6),
                      ],
                      decoration: InputDecoration(
                        labelText: auth.isTestVerification
                            ? 'Test code'
                            : 'Verification code',
                        border: const OutlineInputBorder(),
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
                                _error = null;
                              });
                            },
                      child: Text(
                        auth.isTestVerification
                            ? 'Change number'
                            : 'Change number or resend code',
                      ),
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
            onPressed: _busy
                ? null
                : () {
                    auth.cancelVerificationCode();
                    Navigator.of(context).pop(false);
                  },
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
                  : AuthPhoneNumber.isTestAccount(_phoneNumber ?? '')
                  ? 'Continue with test account'
                  : 'Send code',
            ),
          ),
        ],
      ),
    );
  }
}
