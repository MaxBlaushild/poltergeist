import 'package:phone_numbers_parser/phone_numbers_parser.dart';

/// Country-aware parsing shared by phone inputs and authentication validation.
abstract final class InternationalPhoneNumber {
  static final _supportedCharacters = RegExp(r'^[0-9+\s().-]+$');
  static final _formatting = RegExp(r'[\s().-]');
  static final _nationalDigits = RegExp(r'^\d+$');
  static final _internationalDigits = RegExp(r'^\+[1-9]\d{1,14}$');

  /// Parses a valid number, interpreting national input in [country].
  ///
  /// A leading + always supplies its own country calling code. For shared
  /// calling codes, retain the selected country when the number is valid there.
  /// Letters, extensions, and unsupported punctuation are rejected before the
  /// parser can silently remove them.
  static PhoneNumber? parse(String input, {required IsoCode country}) {
    if (!_supportedCharacters.hasMatch(input)) return null;
    final digits = input.replaceAll(_formatting, '');
    final isInternational = digits.startsWith('+');
    if (isInternational) {
      if (!_internationalDigits.hasMatch(digits)) return null;
    } else if (!_nationalDigits.hasMatch(digits)) {
      return null;
    }

    try {
      var parsed = isInternational
          ? PhoneNumber.parse(digits)
          : PhoneNumber.parse(digits, destinationCountry: country);

      if (isInternational && parsed.countryCode == callingCode(country)) {
        final selectedCountry = PhoneNumber(isoCode: country, nsn: parsed.nsn);
        if (selectedCountry.isValid()) parsed = selectedCountry;
      }

      if (!parsed.isValid() ||
          !_internationalDigits.hasMatch(parsed.international)) {
        return null;
      }
      return parsed;
    } on PhoneNumberException {
      return null;
    }
  }

  /// Returns the canonical E.164 value only when the complete number is valid.
  static String? normalize(String input, {required IsoCode country}) =>
      parse(input, country: country)?.international;

  /// The actual E.164 calling code, without +, from the parser's metadata.
  /// Country picker display data can include a NANP area code and must not be
  /// used to construct the international number.
  static String callingCode(IsoCode country) =>
      PhoneNumber(isoCode: country, nsn: '').countryCode;
}
