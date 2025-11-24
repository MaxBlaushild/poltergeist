import 'dart:io';
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

  /// Gets documents from friends
  /// 
  /// Returns a list of documents from all friends, sorted by createdAt descending
  Future<List<Map<String, dynamic>>> getFriendsDocuments() async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        ApiConstants.friendsDocumentsEndpoint,
      );
      return response
          .map((doc) => Map<String, dynamic>.from(doc))
          .toList();
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
  /// 
  /// Returns the updated document
  Future<Map<String, dynamic>> updateDocument({
    required String documentId,
    String? title,
    String? content,
    List<String>? existingTagIds,
    List<String>? newTagTexts,
  }) async {
    try {
      final data = <String, dynamic>{};
      
      if (title != null) data['title'] = title;
      if (content != null) data['content'] = content;
      if (existingTagIds != null) data['existingTagIds'] = existingTagIds;
      if (newTagTexts != null) data['newTagTexts'] = newTagTexts;

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
}

