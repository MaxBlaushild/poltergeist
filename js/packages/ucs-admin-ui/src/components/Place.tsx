import React from 'react';
import { usePlace } from '@poltergeist/hooks';
import { useParams } from 'react-router-dom';
export const Place = () => {
  const { id } = useParams();
  const { place, loading, error } = usePlace(id!);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-4">{place?.displayName?.text}</h1>
    
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-4">
          {place?.formattedAddress && (
            <div>
              <h2 className="text-lg font-semibold">Address</h2>
              <p className="text-gray-600">{place.formattedAddress}</p>
            </div>
          )}

          {place?.rating && (
            <div>
              <h2 className="text-lg font-semibold">Rating</h2>
              <p className="text-gray-600">{place.rating} / 5 ({place.userRatingCount} reviews)</p>
            </div>
          )}

          {place?.primaryType && (
            <div>
              <h2 className="text-lg font-semibold">Type</h2>
              <p className="text-gray-600">{place.primaryType}</p>
            </div>
          )}

          {(place?.internationalPhoneNumber || place?.nationalPhoneNumber) && (
            <div>
              <h2 className="text-lg font-semibold">Contact</h2>
              <p className="text-gray-600">{place.internationalPhoneNumber || place.nationalPhoneNumber}</p>
            </div>
          )}
        </div>

        <div className="space-y-4">
          {place?.editorialSummary?.text && (
            <div>
              <h2 className="text-lg font-semibold">About</h2>
              <p className="text-gray-600">{place.editorialSummary.text}</p>
            </div>
          )}

          <div>
            <h2 className="text-lg font-semibold">Features</h2>
            <div className="grid grid-cols-2 gap-2">
              {place?.delivery && <span className="text-gray-600">✓ Delivery</span>}
              {place?.takeout && <span className="text-gray-600">✓ Takeout</span>}
              {place?.dineIn && <span className="text-gray-600">✓ Dine-in</span>}
              {place?.outdoorSeating && <span className="text-gray-600">✓ Outdoor seating</span>}
              {place?.servesCocktails && <span className="text-gray-600">✓ Cocktails</span>}
              {place?.servesBeer && <span className="text-gray-600">✓ Beer</span>}
              {place?.servesWine && <span className="text-gray-600">✓ Wine</span>}
              {place?.goodForGroups && <span className="text-gray-600">✓ Good for groups</span>}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};