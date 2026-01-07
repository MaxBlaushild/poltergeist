import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

export const AntiviralPage = () => {
  const navigate = useNavigate();

  useEffect(() => {
    // Set the antiviral-installed localStorage variable
    localStorage.setItem('antiviral-installed', 'true');
    
    // Redirect to press scanner after a brief delay
    const timer = setTimeout(() => {
      navigate('/press-scanner', { replace: true });
    }, 3000);

    return () => clearTimeout(timer);
  }, [navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="bg-black/90 backdrop-blur-sm p-4 md:p-6 lg:p-8 rounded-lg border-2 border-[#00ff00] shadow-[0_0_20px_rgba(0,255,0,0.5)] w-full max-w-md mx-4 matrix-card">
        <h1 className="text-xl md:text-2xl font-bold mb-6 text-center text-[#00ff00]">Antiviral Protection Activated</h1>
        <p className="text-[#00ff00] text-center">
          You can now inject antiviral into servers.
        </p>
      </div>
    </div>
  );
};

