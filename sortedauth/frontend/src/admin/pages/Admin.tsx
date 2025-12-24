import { useEffect } from "react";
import { 
    $Users, 
    $UsersLoading, 
    $UsersError, 
    $UsersCurrentPage,
    $UsersCanGoNext,
    $UsersCanGoPrev,
    goToUsersPage,
    usersNextPage,
    usersPrevPage
} from "../store/admin";
import { useStore } from "@nanostores/react";
import { useParams, useNavigate } from "react-router-dom";

export default function Admin() {
    const users = useStore($Users);
    const loading = useStore($UsersLoading);
    const error = useStore($UsersError);
    const currentPage = useStore($UsersCurrentPage);
    const canGoNext = useStore($UsersCanGoNext);
    const canGoPrev = useStore($UsersCanGoPrev);
    
    const { page } = useParams<{ page: string }>();
    const navigate = useNavigate();
    const urlPage = parseInt(page || "1", 10);

    useEffect(() => {
        // Redirect to page 1 if no page or invalid page
        if (!page || isNaN(urlPage) || urlPage < 1) {
            navigate("/users/1", { replace: true });
            return;
        }
        
        // Always fetch on mount or page change
        goToUsersPage(urlPage);
    }, [page, urlPage, navigate]);

    const handlePreviousPage = () => {
        if (canGoPrev) {
            navigate(`/users/${currentPage - 1}`);
            usersPrevPage();
        }
    };

    const handleNextPage = () => {
        if (canGoNext) {
            navigate(`/users/${currentPage + 1}`);
            usersNextPage();
        }
    };

    if (loading) {
        return (
            <div className="flex-1 bg-gray-950 text-white">
                <div className="p-8">
                    <h1 className="text-3xl font-semibold mb-2">All Users</h1>
                    <p className="text-gray-400 mb-6">View all users in the system</p>
                    <div className="text-center py-8 text-gray-400">Loading users...</div>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex-1 bg-gray-950 text-white">
                <div className="p-8">
                    <h1 className="text-3xl font-semibold mb-2">All Users</h1>
                    <p className="text-gray-400 mb-6">View all users in the system</p>
                    <div className="bg-red-900/20 border border-red-800 text-red-400 px-4 py-3 rounded-lg">
                        Error: {error}
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="flex-1 bg-gray-950 text-white">
            <div className="p-8">
                <h1 className="text-3xl font-semibold mb-2">All Users</h1>
                <p className="text-gray-400 mb-6">View all users in the system</p>
                
                {users.length === 0 ? (
                    <div className="text-gray-400 text-center py-8">No users found</div>
                ) : (
                    <>
                        {/* User Table */}
                        <div className="bg-gray-900 rounded-lg border border-gray-800 overflow-hidden mb-6">
                            <table className="w-full">
                                <thead className="bg-gray-800/50 border-b border-gray-800">
                                    <tr>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Name</th>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Email</th>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">User ID</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-800">
                                    {users.map((user) => (
                                        <tr key={user.id} className="hover:bg-gray-800/50">
                                            <td className="px-6 py-4 text-white">{user.name}</td>
                                            <td className="px-6 py-4 text-gray-300">{user.email}</td>
                                            <td className="px-6 py-4 text-gray-400">{user.id}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                        
                        {/* Pagination */}
                        <div className="flex justify-center items-center space-x-2">
                            <button 
                                onClick={handlePreviousPage}
                                disabled={!canGoPrev}
                                className="px-4 py-2 bg-gray-900 border border-gray-800 rounded-lg hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed text-white"
                            >
                                Previous
                            </button>
                            <span className="px-4 py-2 bg-blue-600 text-white rounded-lg">
                                Page {currentPage}
                            </span>
                            <button 
                                onClick={handleNextPage}
                                disabled={!canGoNext}
                                className="px-4 py-2 bg-gray-900 border border-gray-800 rounded-lg hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed text-white"
                            >
                                Next
                            </button>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
}