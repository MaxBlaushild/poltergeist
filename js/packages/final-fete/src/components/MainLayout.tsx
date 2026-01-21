import { Link, useLocation, Outlet } from 'react-router-dom';

export const MainLayout = () => {
  const location = useLocation();
  const isLoggedIn = localStorage.getItem('token');

  if (!isLoggedIn) {
    return <Outlet />;
  }

  const isRoomsList = location.pathname === '/';
  const isCodeEntry = location.pathname === '/codes';

  return (
    <div className="flex flex-col min-h-screen">
      {/* Tab Navigation */}
      <div className="border-b-2 border-[#00ff00] bg-black/90 backdrop-blur-sm">
        <div className="container mx-auto px-4">
          <div className="flex gap-4">
            <Link
              to="/"
              className={`py-4 px-6 text-center font-medium transition-colors border-b-2 min-w-[120px] ${
                isRoomsList
                  ? 'border-[#00ff00] text-[#00ff00] font-bold'
                  : 'border-transparent text-[#00ff00] opacity-70 hover:opacity-100'
              }`}
            >
              Rooms
            </Link>
            <Link
              to="/codes"
              className={`py-4 px-6 text-center font-medium transition-colors border-b-2 min-w-[120px] ${
                isCodeEntry
                  ? 'border-[#00ff00] text-[#00ff00] font-bold'
                  : 'border-transparent text-[#00ff00] opacity-70 hover:opacity-100'
              }`}
            >
              ?????
            </Link>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1">
        <Outlet />
      </div>
    </div>
  );
};

