import 'package:flutter_test/flutter_test.dart';
import 'package:phone_numbers_parser/phone_numbers_parser.dart';
import 'package:travel_angels/utils/international_phone_number.dart';

void main() {
  group('national number parsing', () {
    test('uses the selected US calling code', () {
      final number = InternationalPhoneNumber.parse(
        '(202) 555-0128',
        country: IsoCode.US,
      );
      expect(number?.isoCode, IsoCode.US);
      expect(number?.international, '+12025550128');
    });

    test('removes the UK national trunk prefix', () {
      final number = InternationalPhoneNumber.parse(
        '020 7946 0018',
        country: IsoCode.GB,
      );
      expect(number?.isoCode, IsoCode.GB);
      expect(number?.nsn, '2079460018');
      expect(number?.international, '+442079460018');
    });

    test('retains the meaningful Italian leading zero', () {
      final number = InternationalPhoneNumber.parse(
        '02 3661 8300',
        country: IsoCode.IT,
      );
      expect(number?.isoCode, IsoCode.IT);
      expect(number?.nsn, '0236618300');
      expect(number?.international, '+390236618300');
    });

    test('removes the Japanese national trunk prefix', () {
      final number = InternationalPhoneNumber.parse(
        '090-1234-5678',
        country: IsoCode.JP,
      );
      expect(number?.isoCode, IsoCode.JP);
      expect(number?.nsn, '9012345678');
      expect(number?.international, '+819012345678');
    });
  });

  group('international paste', () {
    test('detects the pasted country instead of the selected country', () {
      final number = InternationalPhoneNumber.parse(
        '+44 (20) 7946-0018',
        country: IsoCode.US,
      );
      expect(number?.isoCode, IsoCode.GB);
      expect(number?.international, '+442079460018');
    });

    test('retains Italian zero when pasted into another country', () {
      final number = InternationalPhoneNumber.parse(
        '+39 02 3661 8300',
        country: IsoCode.JP,
      );
      expect(number?.isoCode, IsoCode.IT);
      expect(number?.international, '+390236618300');
    });

    test('detects an unambiguous Canadian area code under shared +1', () {
      final number = InternationalPhoneNumber.parse(
        '+1 613 555 0165',
        country: IsoCode.GB,
      );
      expect(number?.isoCode, IsoCode.CA);
      expect(number?.international, '+16135550165');
    });

    for (final country in [IsoCode.US, IsoCode.CA]) {
      test('preserves selected $country for an ambiguous +1 number', () {
        final number = InternationalPhoneNumber.parse(
          '+1 800 555 1234',
          country: country,
        );
        expect(number?.isoCode, country);
        expect(number?.international, '+18005551234');
      });
    }
  });

  group('validation', () {
    for (final input in [
      '',
      ' ',
      '123',
      '0000000000',
      '202555',
      '+',
      '+4',
      '+44 20',
      '+999 123456789',
      '+1 202 555 0128 123456',
      '++1 202 555 0128',
      '202+5550128',
      '2025550128x123',
      '2025550128 ext. 123',
      '2025550128#123',
      '2025550128,123',
      'tel:+12025550128',
      '1-800-FLOWERS',
      '+١٢٠٢٥٥٥٠١٢٨',
    ]) {
      test('does not produce a canonical value for "$input"', () {
        expect(
          InternationalPhoneNumber.normalize(input, country: IsoCode.US),
          isNull,
        );
      });
    }

    test('accepts ordinary formatting without requiring a mobile line', () {
      expect(
        InternationalPhoneNumber.normalize(
          ' +44 (20) 7946.0018 ',
          country: IsoCode.US,
        ),
        '+442079460018',
      );
    });

    test('keeps both existing demo accounts valid', () {
      for (final number in ['+14407858475', '+12025550101']) {
        expect(
          InternationalPhoneNumber.normalize(number, country: IsoCode.US),
          number,
        );
      }
    });
  });

  test('uses calling codes without duplicating NANP area codes', () {
    expect(InternationalPhoneNumber.callingCode(IsoCode.US), '1');
    expect(InternationalPhoneNumber.callingCode(IsoCode.CA), '1');
    expect(InternationalPhoneNumber.callingCode(IsoCode.AS), '1');
    expect(InternationalPhoneNumber.callingCode(IsoCode.AI), '1');
    expect(InternationalPhoneNumber.callingCode(IsoCode.GB), '44');
    expect(InternationalPhoneNumber.callingCode(IsoCode.IT), '39');
    expect(InternationalPhoneNumber.callingCode(IsoCode.JP), '81');
  });
}
