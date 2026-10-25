import 'dart:io';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/document_location.dart';
import 'package:travel_angels/services/api_client.dart';

class DocumentService {
  final APIClient _apiClient;

  DocumentService(this._apiClient) {
    _apiClient.preserveSessionOnForbidden();
  }

  /// Creates a document
  ///
  /// [title] - Document title
  /// [provider] - Document provider (google_docs, google_sheets, internal, unknown)
  /// [link] - Optional document link
  /// [content] - Optional document content
  /// [tagIds] - Optional list of existing tag IDs
  /// [newTags] - Optional list of new tag texts
  /// [locations] - Optional list of document locations
  ///
  /// Returns the created document
  Future<Map<String, dynamic>> createDocument({
    required String title,
    required String provider,
    String? link,
    String? content,
    List<String>? tagIds,
    List<String>? newTags,
    List<DocumentLocation>? locations,
    String? format,
    String? personalNote,
    String? destination,
  }) async {
    try {
      final data = <String, dynamic>{'title': title, 'provider': provider};

      if (link != null) data['link'] = link;
      if (content != null) data['content'] = content;
      if (format != null) data['format'] = format;
      if (personalNote != null) data['personalNote'] = personalNote;
      if (destination != null) data['destination'] = destination;
      if (tagIds != null && tagIds.isNotEmpty) {
        data['existingTagIds'] = tagIds;
      }
      if (newTags != null && newTags.isNotEmpty) {
        data['newTagTexts'] = newTags;
      }
      if (locations != null && locations.isNotEmpty) {
        data['locations'] = locations
            .map(
              (loc) => {
                'placeId': loc.placeId,
                'name': loc.name,
                'formattedAddress': loc.formattedAddress,
                'latitude': loc.latitude,
                'longitude': loc.longitude,
                'type': loc.locationType.name,
              },
            )
            .toList();
      }

      final response = await _apiClient.post<Map<String, dynamic>>(
        ApiConstants.documentsEndpoint,
        data: data,
      );
      return response;
    } catch (e) {
      rethrow;
    }
  }

  /// All library data is scoped to the authenticated viewer on the server.
  Future<List<Map<String, dynamic>>> getLibraryDocuments(String scope) async {
    final response = await _apiClient.get<List<dynamic>>(
      '${ApiConstants.documentsEndpoint}/library',
      params: {'scope': scope},
    );
    return response.map((doc) => Map<String, dynamic>.from(doc)).toList();
  }

  Future<List<Map<String, dynamic>>> getTripDocuments(String tripId) async {
    final response = await _apiClient.get<List<dynamic>>(
      '${ApiConstants.documentsEndpoint}/library',
      params: {'scope': 'saved', 'tripId': tripId},
    );
    return response.map((doc) => Map<String, dynamic>.from(doc)).toList();
  }

  Future<Map<String, dynamic>> getDocument(String id) => _apiClient
      .get<Map<String, dynamic>>('${ApiConstants.documentsEndpoint}/$id');

  Future<Map<String, dynamic>> setDocumentSaved(String id, bool saved) =>
      _apiClient.put<Map<String, dynamic>>(
        '${ApiConstants.documentsEndpoint}/$id/saved',
        data: {'saved': saved},
      );

  Future<Map<String, dynamic>> updateDocumentSharing(
    String id, {
    required String visibility,
    List<String> userIds = const [],
  }) => _apiClient.put<Map<String, dynamic>>(
    '${ApiConstants.documentsEndpoint}/$id/sharing',
    data: {'visibility': visibility, 'userIds': userIds},
  );

  Future<List<Map<String, dynamic>>> getDocumentTrips(String id) async {
    final response = await _apiClient.get<List<dynamic>>(
      '${ApiConstants.documentsEndpoint}/$id/trips',
    );
    return response.map((trip) => Map<String, dynamic>.from(trip)).toList();
  }

  Future<Map<String, dynamic>> setDocumentTrip(
    String id,
    String tripId,
    bool saved,
  ) => _apiClient.put<Map<String, dynamic>>(
    '${ApiConstants.documentsEndpoint}/$id/trips/$tripId',
    data: {'saved': saved},
  );

  /// Gets documents for a user
  ///
  /// [userId] - The user ID
  ///
  /// Returns a list of documents
  Future<List<Map<String, dynamic>>> getDocumentsByUserId(String userId) async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        '${ApiConstants.documentsEndpoint}/user/$userId',
      );
      return response.map((doc) => Map<String, dynamic>.from(doc)).toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Gets documents from friends
  ///
  /// Returns a list of documents from all friends, sorted by createdAt descending
  Future<List<Map<String, dynamic>>> getFriendsDocuments() async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.friendsDocumentsEndpoint,
      );
      return response.map((doc) => Map<String, dynamic>.from(doc)).toList();
    } catch (e) {
      rethrow;
    }
  }

  /// Updates a document
  ///
  /// [documentId] - The document ID
  /// [title] - Optional new title
  /// [content] - Optional new content
  /// [existingTagIds] - Optional list of existing tag IDs to keep/add
  /// [newTagTexts] - Optional list of new tag texts to create and add
  /// [locations] - Optional list of document locations (replaces all existing)
  ///
  /// Returns the updated document
  Future<Map<String, dynamic>> updateDocument({
    required String documentId,
    String? title,
    String? content,
    List<String>? existingTagIds,
    List<String>? newTagTexts,
    List<DocumentLocation>? locations,
    String? format,
    String? personalNote,
    String? destination,
  }) async {
    try {
      final data = <String, dynamic>{};

      if (title != null) data['title'] = title;
      if (content != null) data['content'] = content;
      if (format != null) data['format'] = format;
      if (personalNote != null) data['personalNote'] = personalNote;
      if (destination != null) data['destination'] = destination;
      if (existingTagIds != null) data['existingTagIds'] = existingTagIds;
      if (newTagTexts != null) data['newTagTexts'] = newTagTexts;
      if (locations != null) {
        data['locations'] = locations
            .map(
              (loc) => {
                'placeId': loc.placeId,
                'name': loc.name,
                'formattedAddress': loc.formattedAddress,
                'latitude': loc.latitude,
                'longitude': loc.longitude,
                'type': loc.locationType.name,
              },
            )
            .toList();
      }

      final response = await _apiClient.put<Map<String, dynamic>>(
        ApiConstants.updateDocumentEndpoint(documentId),
        data: data,
      );
      return response;
    } catch (e) {
      rethrow;
    }
  }

  /// Deletes a document
  ///
  /// [documentId] - The document ID
  ///
  /// Throws an exception if deletion fails
  Future<void> deleteDocument(String documentId) async {
    try {
      await _apiClient.delete<void>(
        ApiConstants.updateDocumentEndpoint(documentId),
      );
    } catch (e) {
      rethrow;
    }
  }

  /// Parses a document file (PDF or Word)
  ///
  /// [file] - The file to parse
  ///
  /// Returns parsed document data (content, fileType, wordCount, etc.)
  Future<Map<String, dynamic>> parseDocument(File file) async {
    try {
      final response = await _apiClient.postMultipart<Map<String, dynamic>>(
        ApiConstants.parseDocumentEndpoint,
        filePath: file.path,
      );
      return response;
    } catch (e) {
      rethrow;
    }
  }

  Future<Map<String, dynamic>> parseDocumentBytes(
    List<int> bytes,
    String filename,
  ) => _apiClient.postMultipartBytes<Map<String, dynamic>>(
    ApiConstants.parseDocumentEndpoint,
    bytes: bytes,
    filename: filename,
  );
}
