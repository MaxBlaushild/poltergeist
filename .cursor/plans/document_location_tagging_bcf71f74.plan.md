---
name: Document Location Tagging
overview: Add the ability to tag documents with cities, countries, or continents using a search/autocomplete dropdown. Locations will be stored separately from regular tags in a new database table, and users can select multiple locations when uploading documents via any method (file upload, Google Drive import).
todos: []
---

# Document Location Tagging Implementation

## Overview

Add location tagging (cities, countries, continents) to documents. Locations are stored separately from regular tags, support multiple selections, and are available during all document upload methods.

## Backend Implementation

### 1. Database Migration

**File**: `go/migrate/internal/migrations/000XXX_add_document_locations.up.sql` (new file)

- Create `document_locations` table with:
- `id` (UUID, primary key)
- `created_at`, `updated_at` (timestamps)
- `document_id` (UUID, foreign key to documents)
- `place_id` (VARCHAR, Google Maps place ID)
- `name` (VARCHAR, location name)
- `formatted_address` (TEXT, full address)
- `latitude` (DOUBLE PRECISION)
- `longitude` (DOUBLE PRECISION)
- `location_type` (VARCHAR, enum: 'city', 'country', 'continent')
- Add indexes on `document_id` and `place_id`
- Add foreign key constraint to `documents` table with CASCADE delete

### 2. Go Model

**File**: `go/pkg/models/document_location.go` (new file)

- Create `DocumentLocation` struct with GORM tags
- Add `LocationType` enum (city, country, continent)
- Add relationship to `Document` model

**File**: `go/pkg/models/document.go`

- Add `DocumentLocations []DocumentLocation` field with GORM many-to-many or one-to-many relationship

### 3. Database Handler

**File**: `go/pkg/db/document_location.go` (new file)

- Create `DocumentLocationHandle` interface with methods:
- `Create(ctx, location)`
- `FindByDocumentID(ctx, documentID)`
- `DeleteByDocumentID(ctx, documentID)`
- Implement handler struct and methods

**File**: `go/pkg/db/interfaces.go`

- Add `DocumentLocationHandle` to the `DBClient` interface

**File**: `go/pkg/db/client.go`

- Add `DocumentLocation()` method to return handler

### 4. API Updates

**File**: `go/travel-angels/internal/server/document.go`

- Update `CreateDocumentRequest` struct to include:
- `Locations []DocumentLocationRequest` field
- Create `DocumentLocationRequest` struct with:
- `PlaceId`, `Name`, `FormattedAddress`, `Latitude`, `Longitude`, `Type`
- Update `CreateDocument` handler to:
- Parse location requests
- Create `DocumentLocation` records after document creation
- Return document with locations in response
- Update `UpdateDocumentRequest` struct similarly
- Update `UpdateDocument` handler to:
- Handle location updates (replace all locations for simplicity)
- Delete existing locations and create new ones

**File**: `go/travel-angels/internal/server/google_drive.go`

- Update `ImportGoogleDriveDocument` handler to accept and process location data from request

### 5. Location Type Detection

**File**: `go/travel-angels/internal/server/document.go` (helper function)

- Create helper function to determine location type from Google Maps response:
- Check address components for country, city, etc.
- Default to 'city' if unclear
- Or allow frontend to specify type explicitly

## Frontend Implementation

### 1. Location Model

**File**: `dart/travel_angels/lib/models/document_location.dart` (new file)

- Create `DocumentLocation` class with:
- `id`, `placeId`, `name`, `formattedAddress`, `latitude`, `longitude`, `type`
- JSON serialization methods

**File**: `dart/travel_angels/lib/models/document.dart`

- Add `documentLocations` field (List<DocumentLocation>?)

### 2. Location Selector Widget

**File**: `dart/travel_angels/lib/widgets/location_selector.dart` (new file)

- Create reusable widget for selecting multiple locations
- Features:
- Search/autocomplete using `LocationService`
- Display selected locations as chips
- Allow removing selected locations
- Show location type (city/country/continent) - either detect from address or let user specify
- Support multiple selections

### 3. Document Service Updates

**File**: `dart/travel_angels/lib/services/document_service.dart`

- Update `createDocument` method to accept:
- `List<DocumentLocation>? locations` parameter
- Include locations in request payload as `locations` array
- Update `updateDocument` method similarly

### 4. Upload Flow Updates

**File**: `dart/travel_angels/lib/widgets/file_picker_widget.dart`

- Add location selector before/after file selection
- Store selected locations in state
- Pass locations to `createDocument` call

**File**: `dart/travel_angels/lib/widgets/google_drive_file_picker.dart`

- Add location selector in the import flow
- Pass locations when calling `importDocument` API

**File**: `dart/travel_angels/lib/services/google_drive_service.dart`

- Update `importDocument` method to accept locations parameter
- Include locations in API request

### 5. API Constants

**File**: `dart/travel_angels/lib/constants/api_constants.dart`

- Verify `locationSearchEndpoint` exists (should already be there)

### 6. Display Locations

**File**: `dart/travel_angels/lib/screens/discover_screen.dart`

- Display document locations alongside tags in document cards

**File**: `dart/travel_angels/lib/screens/edit_document_screen.dart`

- Add location editing capability (similar to tag editing)

## Location Type Handling

Since Google Maps API doesn't directly provide city/country/continent classification, we have two options:

1. **Parse address components**: Extract city/country from `formattedAddress` or address components
2. **User selection**: Allow users to specify type when selecting location
3. **Hybrid**: Auto-detect when possible, allow manual override

**Recommendation**: Start with option 1 (auto-detect from address components), add option 2 if needed.

## Testing Considerations

- Test document creation with multiple locations
- Test location updates (replace all)
- Test location search/autocomplete
- Test with all upload methods (file, Google Drive)
- Verify locations are displayed correctly in document lists