import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Category } from '@poltergeist/types';

export const useCategories = () => {
  const { apiClient } = useAPI();
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const response = await apiClient.get<Category[]>('/sonar/categories');
        setCategories(response);
        setLoading(false);
      } catch (err) {
        setError(err);
        setLoading(false);
      }
    };

    fetchCategories();
  }, []);

  return { categories, loading, error };
};

export default useCategories;
