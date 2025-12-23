import { Link, Outlet, useLocation } from "react-router-dom";

// Layout component with sidebar
function Layout() {
    const location = useLocation();
  
    const isActive = (path: string) => {
      return location.pathname === path;
    };
  
    return (
      <div className="flex h-screen bg-gray-50">
        <div className="w-64 bg-white border-r border-gray-200 flex flex-col">
          <div className="p-4 border-b border-gray-200">
            <h1 className="text-xl font-semibold text-gray-800">User Management</h1>
          </div>
          
          <nav className="flex-1 p-4">
            <div className="space-y-2">
              <Link
                to="/"
                className={`block w-full text-left px-4 py-3 rounded-lg transition-colors ${
                  isActive('/')
                    ? 'bg-blue-50 text-blue-700 border border-blue-200'
                    : 'text-gray-700 hover:bg-gray-50'
                }`}
              >
                Users
              </Link>
            </div>
          </nav>
        </div>
  
        {/* Main Content */}
        <div className="flex-1 overflow-auto">
          <Outlet />
        </div>
      </div>

    );
  }

  export default Layout;