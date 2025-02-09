import React from 'react';
import { usePointOfInterestGroups, useCityName } from '@poltergeist/hooks';
import { useAPI, useMediaContext } from '@poltergeist/contexts';
import { PointOfInterestGroup } from '@poltergeist/types';

const PointOfInterestGroupItem = ({ group }) => {
  const { lat, lng } = group.pointsOfInterest?.[0] || {};
  const { cityName, loading, error } = useCityName(lat, JSON.stringify(parseFloat(lng)));
  return (
    <li className="border p-4 mb-2 rounded shadow hover:shadow-md transition-shadow">
      <a href={`/arena/${group.id}`} className="block">
        <div className="flex gap-4">
          {group.imageUrl && (
            <img 
              src={group.imageUrl} 
              alt={group.name} 
              className="w-32 h-32 rounded-lg object-cover" 
            />
          )}
          <div className="flex flex-col items-end flex-1">
            <h3 className="text-xl font-bold mb-2 text-left w-full">{group.name}</h3>
            <p className="text-gray-600 text-left w-full">{group.description}</p>
            {cityName && (
              <p className="text-gray-600 text-left w-full">
                Location: {cityName}
              </p>
            )}
            <p className="text-lg font-bold text-gray-700 text-left w-full">
              Points of Interest: {group.pointsOfInterest?.length || 0}
            </p>
          </div>
        </div>
      </a>
    </li>
  );
}

export const Arenas = () => {
  const { pointOfInterestGroups, loading, error } = usePointOfInterestGroups();
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { apiClient } = useAPI();
  const [isModalOpen, setIsModalOpen] = React.useState(false);
  const[name, setName] = React.useState<string>('');
  const[description, setDescription] = React.useState<string>('');
  const [image, setImage] = React.useState<File | undefined>(
    undefined
  );
  const fileInputRef = React.useRef(null);

  const handleImageUpload = (e) => {
    const file = e.target.files[0];
    if (file) {
      setImage(file);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const imageKey = `arenas/${(image?.name || 'image.jpg').toLowerCase().replace(/\s+/g, '-')}`;
    let imageUrl = '';

    if (image) {
      const presignedUrl = await getPresignedUploadURL("crew-points-of-interest", imageKey);
      if (!presignedUrl) return;
      await uploadMedia(presignedUrl, image);
      imageUrl = presignedUrl.split("?")[0];
    }

    setIsModalOpen(false);
    setName('');
    setDescription('');
    setImage(undefined);

    await apiClient.post<PointOfInterestGroup>(`/sonar/pointsOfInterest/group`, {
      name: name,
      description: description,
      imageUrl: imageUrl,
    });
  };

  return (
    <div className="flex flex-col gap-4 p-4">
      <ul className="list-none">
        {pointOfInterestGroups?.map((group) => (
          <PointOfInterestGroupItem key={group.id} group={group} />
        ))}
      </ul>
      <button 
        className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded w-fit"
        onClick={() => setIsModalOpen(true)}
      >
        Create New POI Group
      </button>

      {isModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">Create New Arena</h2>
            <form onSubmit={handleSubmit} className="flex flex-col gap-4">
              <input
                type="text"
                placeholder="Arena Name"
                className="border p-2 rounded"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
              <textarea
                placeholder="Description"
                className="border p-2 rounded h-24"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                required
              />
              <input
                type="file"
                accept="image/*"
                ref={fileInputRef}
                onChange={handleImageUpload}
              />

              {image && (
                <img 
                  src={URL.createObjectURL(image)} 
                  alt="Preview" 
                  className="w-full h-32 object-cover rounded"
                />
              )}
              <div className="flex gap-2 justify-end">
                <button
                  type="button"
                  onClick={() => {
                    setIsModalOpen(false);
                    setName('');
                    setDescription('');
                    setImage(undefined);
                  }}
                  className="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
