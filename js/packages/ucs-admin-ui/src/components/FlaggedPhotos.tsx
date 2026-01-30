import { useAPI } from '@poltergeist/contexts';
import React, { useState, useEffect } from 'react';

interface FlaggedPostUser {
  id?: string;
  username?: string;
  phoneNumber?: string;
  profilePictureUrl?: string;
}

interface FlaggedPost {
  id?: string;
  imageUrl?: string;
  caption?: string;
  mediaType?: string;
  createdAt?: string;
  userId?: string;
  tags?: string[];
}

interface FlaggedPostItem {
  post: FlaggedPost;
  user: FlaggedPostUser;
  reactions?: { emoji: string; count: number }[];
  flagCount: number;
}

export const FlaggedPhotos = () => {
  const { apiClient } = useAPI();
  const [items, setItems] = useState<FlaggedPostItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actioning, setActioning] = useState<string | null>(null);

  const fetchFlagged = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<FlaggedPostItem[]>(
        '/verifiable-sn/admin/flagged-posts'
      );
      setItems(Array.isArray(response) ? response : []);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(`Failed to load flagged posts: ${msg}`);
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFlagged();
  }, []);

  const handleDismiss = async (postId: string) => {
    if (!postId) return;
    setActioning(postId);
    try {
      await apiClient.post(
        `/verifiable-sn/admin/flagged-posts/${postId}/dismiss`
      );
      setItems((prev) => prev.filter((i) => i.post.id !== postId));
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      alert(`Failed to dismiss: ${msg}`);
    } finally {
      setActioning(null);
    }
  };

  const handleDelete = async (postId: string) => {
    if (!postId || !confirm('Delete this post permanently?')) return;
    setActioning(postId);
    try {
      await apiClient.delete(
        `/verifiable-sn/admin/flagged-posts/${postId}`
      );
      setItems((prev) => prev.filter((i) => i.post.id !== postId));
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      alert(`Failed to delete: ${msg}`);
    } finally {
      setActioning(null);
    }
  };

  if (loading) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-4">Flagged Photos</h1>
        <p>Loading...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <h1 className="text-2xl font-bold mb-4">Flagged Photos</h1>
        <p className="text-red-600">{error}</p>
        <button
          onClick={fetchFlagged}
          className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Flagged Photos</h1>
      <p className="text-gray-600 mb-4">
        Photos that have been reported by users. Use OK to dismiss the flag, or
        Delete to remove the post.
      </p>

      {items.length === 0 ? (
        <p className="text-gray-500">No flagged photos.</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((item) => {
            const postId = item.post.id;
            const isActioning = actioning === postId;

            return (
              <div
                key={postId ?? Math.random()}
                className="border rounded-lg overflow-hidden bg-white shadow"
              >
                <div className="aspect-square bg-gray-100 flex items-center justify-center">
                  {item.post.imageUrl ? (
                    item.post.mediaType === 'video' ? (
                      <video
                        src={item.post.imageUrl}
                        className="w-full h-full object-cover"
                        muted
                        playsInline
                        preload="metadata"
                      />
                    ) : (
                      <img
                        src={item.post.imageUrl}
                        alt=""
                        className="w-full h-full object-cover"
                      />
                    )
                  ) : (
                    <span className="text-gray-400">No media</span>
                  )}
                </div>
                <div className="p-3">
                  <div className="flex items-center gap-2 text-sm text-gray-600 mb-1">
                    <span>
                      {item.user?.username ?? item.user?.phoneNumber ?? 'Unknown'}
                    </span>
                    <span className="text-red-600 font-medium">
                      {item.flagCount} flag{item.flagCount !== 1 ? 's' : ''}
                    </span>
                  </div>
                  {item.post.caption && (
                    <p className="text-sm text-gray-700 truncate mb-2">
                      {item.post.caption}
                    </p>
                  )}
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleDismiss(postId!)}
                      disabled={isActioning}
                      className="flex-1 px-3 py-1.5 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
                    >
                      {isActioning ? '...' : 'OK'}
                    </button>
                    <button
                      onClick={() => handleDelete(postId!)}
                      disabled={isActioning}
                      className="flex-1 px-3 py-1.5 text-sm bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
