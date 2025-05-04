import React from 'react';
import { usePointOfInterestGroups, useCityName } from '@poltergeist/hooks';
import { useAPI, useArena, useMediaContext } from '@poltergeist/contexts';
import { PointOfInterestGroup, PointOfInterestGroupType } from '@poltergeist/types';

const PointOfInterestGroupItem = ({ group }) => {
  const { lat, lng } = group.pointsOfInterest?.[0] || {};
  // const { cityName, loading, error } = useCityName(lat, JSON.stringify(parseFloat(lng)));
  const { apiClient } = useAPI();

  const deletePointOfInterestGroup = async (id: string) => {
    const response = await apiClient.delete(`/sonar/pointsOfInterest/group/${id}`);
    return response;
  };
  return (
    <li className="border p-4 mb-2 rounded shadow hover:shadow-md transition-shadow">
      <div className="flex justify-between">
        <a href={`/arena/${group.id}`} className="block flex-1">
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
              {/* {cityName && (
                <p className="text-gray-600 text-left w-full">
                  Location: {cityName}
                </p>
              )} */}
              <p className="text-gray-600 text-left w-full">
                Type: {PointOfInterestGroupType[group.type]}
              </p>
              <p className="text-gray-600 text-left w-full">
                ID: {group.id}
              </p>
              <p className="text-lg font-bold text-gray-700 text-left w-full">
                Points of Interest: {group.groupMembers?.length || 0}
              </p>
            </div>
          </div>
        </a>
        <button 
          onClick={async (e) => {
            e.preventDefault();
            await deletePointOfInterestGroup(group.id);
          }}
          className="ml-4 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 h-fit"
        >
          Delete
        </button>
      </div>
    </li>
  );
}

export const Arenas = () => {
  const { pointOfInterestGroups, loading, error } = usePointOfInterestGroups();
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { apiClient } = useAPI();
  const [isModalOpen, setIsModalOpen] = React.useState(false);
  const[name, setName] = React.useState<string>('');
  const[type, setType] = React.useState<PointOfInterestGroupType>(PointOfInterestGroupType.Arena);
  const[description, setDescription] = React.useState<string>('');
  const [image, setImage] = React.useState<File | undefined>(undefined);
  const [searchTerm, setSearchTerm] = React.useState<string>('');
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
      type,
    });
  };

  const filteredGroups = pointOfInterestGroups?.filter(group => 
    group.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="w-full">
        <input
          type="text"
          placeholder="Search arenas by name..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full p-2 border rounded shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>
      
      <ul className="list-none">
        {filteredGroups?.sort((a, b) => {
          const dateA = typeof a.createdAt === 'string' ? new Date(a.createdAt) : a.createdAt;
          const dateB = typeof b.createdAt === 'string' ? new Date(b.createdAt) : b.createdAt;
          return dateB.getTime() - dateA.getTime();
        }).map((group) => (
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
              <select
                className="border p-2 rounded"
                value={type}
                onChange={(e) => setType(Number(e.target.value))}
                required
              >
                <option value={PointOfInterestGroupType.Arena}>Arena</option>
                <option value={PointOfInterestGroupType.Quest}>Quest</option>
              </select>
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
