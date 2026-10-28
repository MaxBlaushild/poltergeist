import 'dart:math' as math;

import 'package:country_picker/country_picker.dart';
import 'package:flutter/material.dart';
import 'package:phone_numbers_parser/phone_numbers_parser.dart';
import 'package:travel_angels/utils/international_phone_number.dart';

class _PhoneCountry {
  _PhoneCountry(this.country, this.isoCode)
    : callingCode = PhoneNumber(isoCode: isoCode, nsn: '').countryCode;

  final Country country;
  final IsoCode isoCode;
  final String callingCode;
}

final _phoneCountries = () {
  final codes = {for (final iso in IsoCode.values) iso.name: iso};
  final countries = <String, _PhoneCountry>{};
  for (final country in CountryService().getAll()) {
    final iso = codes[country.countryCode];
    if (iso != null) {
      countries.putIfAbsent(
        country.countryCode,
        () => _PhoneCountry(country, iso),
      );
    }
  }
  return countries.values.toList()
    ..sort((a, b) => a.country.name.compareTo(b.country.name));
}();

/// A national phone field that emits only validated international destinations.
/// Its country remains visible and editable; locale is only the initial choice.
class InternationalPhoneField extends StatefulWidget {
  const InternationalPhoneField({
    super.key,
    required this.onChanged,
    this.enabled = true,
    this.focusNode,
    this.label = 'Phone number',
    this.initialCountryCode,
  });

  final ValueChanged<String?> onChanged;
  final bool enabled;
  final FocusNode? focusNode;
  final String label;
  final String? initialCountryCode;

  @override
  State<InternationalPhoneField> createState() =>
      _InternationalPhoneFieldState();
}

class _InternationalPhoneFieldState extends State<InternationalPhoneField> {
  final _controller = TextEditingController();
  late final FocusNode _focusNode = widget.focusNode ?? FocusNode();
  late _PhoneCountry _country;
  bool _showError = false;

  @override
  void initState() {
    super.initState();
    final preferred = [
      widget.initialCountryCode,
      ...WidgetsBinding.instance.platformDispatcher.locales.map(
        (l) => l.countryCode,
      ),
      'US',
    ];
    _country = preferred
        .whereType<String>()
        .map(
          (code) => _phoneCountries.where(
            (c) => c.isoCode.name == code.toUpperCase(),
          ),
        )
        .firstWhere((matches) => matches.isNotEmpty)
        .first;
    _focusNode.addListener(_onFocusChanged);
  }

  @override
  void dispose() {
    _controller.dispose();
    _focusNode.removeListener(_onFocusChanged);
    if (widget.focusNode == null) _focusNode.dispose();
    super.dispose();
  }

  void _onFocusChanged() {
    if (_focusNode.hasFocus || !widget.enabled) return;
    final parsed = InternationalPhoneNumber.parse(
      _controller.text,
      country: _country.isoCode,
    );
    setState(() => _showError = _controller.text.isNotEmpty && parsed == null);
    if (parsed != null) _replaceText(parsed.formatNsn());
  }

  void _replaceText(String text) {
    _controller.value = TextEditingValue(
      text: text,
      selection: TextSelection.collapsed(offset: text.length),
    );
  }

  void _changed(String input) {
    final parsed = InternationalPhoneNumber.parse(
      input,
      country: _country.isoCode,
    );
    setState(() {
      _showError = false;
      if (parsed != null &&
          input.replaceAll(RegExp(r'[\s().-]'), '').startsWith('+')) {
        final matches = _phoneCountries.where(
          (c) => c.isoCode == parsed.isoCode,
        );
        if (matches.isNotEmpty) {
          _country = matches.first;
          _replaceText(parsed.formatNsn());
        }
      }
    });
    widget.onChanged(parsed?.international);
  }

  Future<void> _pickCountry() async {
    if (!widget.enabled) return;
    _focusNode.unfocus();
    final selected = await showModalBottomSheet<_PhoneCountry>(
      context: context,
      useRootNavigator: true,
      isScrollControlled: true,
      useSafeArea: true,
      showDragHandle: true,
      constraints: const BoxConstraints(maxWidth: 480),
      builder: (context) => Padding(
        padding: EdgeInsets.only(
          bottom: MediaQuery.viewInsetsOf(context).bottom,
        ),
        child: SizedBox(
          height: math.min(
            560,
            math.max(
              180,
              MediaQuery.sizeOf(context).height -
                  MediaQuery.viewInsetsOf(context).bottom -
                  MediaQuery.paddingOf(context).top -
                  100,
            ),
          ),
          child: _CountryPicker(selected: _country),
        ),
      ),
    );
    if (!mounted || selected == null || !widget.enabled) return;
    setState(() => _country = selected);
    // Re-interpret national input under the explicit new country; never keep a
    // previously valid destination after the country or number has changed.
    _changed(_controller.text);
    _focusNode.requestFocus();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final countryLabel =
        'Country code: ${_country.country.name}, +${_country.callingCode}';
    return TextField(
      controller: _controller,
      focusNode: _focusNode,
      enabled: widget.enabled,
      keyboardType: TextInputType.phone,
      autofillHints: const [AutofillHints.telephoneNumber],
      textInputAction: TextInputAction.done,
      onChanged: _changed,
      onSubmitted: (_) => _focusNode.unfocus(),
      decoration: InputDecoration(
        labelText: widget.label,
        hintText: 'Your phone number',
        errorText: _showError
            ? 'Check the number for ${_country.country.name}.'
            : null,
        errorMaxLines: 2,
        prefixIconConstraints: const BoxConstraints(minHeight: 48),
        prefixIcon: Padding(
          padding: const EdgeInsets.only(right: 10),
          child: DecoratedBox(
            decoration: BoxDecoration(
              border: Border(
                right: BorderSide(color: theme.colorScheme.outlineVariant),
              ),
            ),
            child: Semantics(
              label: countryLabel,
              button: true,
              enabled: widget.enabled,
              excludeSemantics: true,
              onTap: widget.enabled ? _pickCountry : null,
              child: Tooltip(
                message: countryLabel,
                child: TextButton(
                  onPressed: widget.enabled ? _pickCountry : null,
                  style: TextButton.styleFrom(
                    padding: const EdgeInsets.symmetric(horizontal: 10),
                    minimumSize: const Size(44, 48),
                    foregroundColor: theme.colorScheme.onSurface,
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text('${_country.isoCode.name} +${_country.callingCode}'),
                      const SizedBox(width: 4),
                      const Icon(Icons.expand_more, size: 18),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _CountryPicker extends StatefulWidget {
  const _CountryPicker({required this.selected});
  final _PhoneCountry selected;

  @override
  State<_CountryPicker> createState() => _CountryPickerState();
}

class _CountryPickerState extends State<_CountryPicker> {
  String _query = '';

  @override
  Widget build(BuildContext context) {
    final countries = _phoneCountries
        .where(
          (country) =>
              country.country.name.toLowerCase().contains(_query) ||
              country.isoCode.name.toLowerCase().startsWith(_query) ||
              country.callingCode.startsWith(
                _query.replaceFirst(RegExp(r'^\+'), ''),
              ),
        )
        .toList();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 0, 8, 8),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  'Country',
                  style: Theme.of(context).textTheme.titleLarge,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              IconButton(
                tooltip: 'Close country picker',
                onPressed: () => Navigator.pop(context),
                icon: const Icon(Icons.close),
              ),
            ],
          ),
        ),
        Expanded(
          child: CustomScrollView(
            keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
            slivers: [
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 20,
                    vertical: 8,
                  ),
                  child: TextField(
                    autofocus: false,
                    decoration: const InputDecoration(
                      hintText: 'Country or calling code',
                      labelText: 'Search countries',
                      prefixIcon: Icon(Icons.search),
                    ),
                    onChanged: (value) =>
                        setState(() => _query = value.trim().toLowerCase()),
                  ),
                ),
              ),
              if (countries.isEmpty)
                const SliverToBoxAdapter(
                  child: Padding(
                    padding: EdgeInsets.all(24),
                    child: Center(child: Text('No matching countries')),
                  ),
                )
              else
                SliverList.builder(
                  itemCount: countries.length,
                  itemBuilder: (context, index) {
                    final country = countries[index];
                    final selected = country.isoCode == widget.selected.isoCode;
                    return ListTile(
                      leading: SizedBox(
                        width: 40,
                        child: FittedBox(
                          fit: BoxFit.scaleDown,
                          alignment: Alignment.centerLeft,
                          child: Text(country.isoCode.name),
                        ),
                      ),
                      title: Text(country.country.name),
                      trailing: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text('+${country.callingCode}'),
                          if (selected) ...[
                            const SizedBox(width: 8),
                            const Icon(Icons.check, size: 18),
                          ],
                        ],
                      ),
                      selected: selected,
                      onTap: () => Navigator.pop(context, country),
                    );
                  },
                ),
            ],
          ),
        ),
      ],
    );
  }
}
