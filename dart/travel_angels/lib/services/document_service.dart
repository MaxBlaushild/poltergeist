import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/services/api_client.dart';

class DocumentService {
  final APIClient _apiClient;

  DocumentService(this._apiClient);

  /// Creates a document
  /// 
  /// [title] - Document title
  /// [provider] - Document provider (google_docs, google_sheets, internal, unknown)
  /// [link] - Optional document link
  /// [content] - Optional document content
  /// [tagIds] - Optional list of existing tag IDs
  /// [newTags] - Optional list of new tag texts
  /// 
  /// Returns the created document
  Future<Map<String, dynamic>> createDocument({
    required String title,
    required String provider,
    String? link,
    String? content,
    List<String>? tagIds,
    List<String>? newTags,
  }) async {
    try {
      final data = <String, dynamic>{
        'title': title,
        'provider': provider,
      };

      if (link != null) data['link'] = link;
      if (content != null) data['content'] = content;
      if (tagIds != null && tagIds.isNotEmpty) {
        data['existingTagIds'] = tagIds;
      }
      if (newTags != null && newTags.isNotEmpty) {
        data['newTagTexts'] = newTags;
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
      return response
          .map((doc) => Map<String, dynamic>.from(doc))
          .toList();
    } catch (e) {
      rethrow;
    }
  }
}

