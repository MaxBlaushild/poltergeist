import 'package:flutter/material.dart';
import 'package:travel_angels/models/document.dart';
import 'package:travel_angels/screens/edit_document_screen.dart';

Future<bool?> editGuide(BuildContext context, Document document) =>
    Navigator.push<bool>(
      context,
      MaterialPageRoute(builder: (_) => EditDocumentScreen(document: document)),
    );
