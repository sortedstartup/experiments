import { useEffect, useState } from "react";
import { $Users, getUsers } from "../store/admin";
import { useStore } from "@nanostores/react";
import { useParams, useNavigate } from "react-router-dom";

export default function Admin() {
    //TODO:pagination logic should be there in nanostore logic
    const users = useStore($Users);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const { page } = useParams<{ page: string }>();
    const navigate = useNavigate();
    const currentPage = parseInt(page || "1", 10);
    const pageSize = 10;

    const fetchUsers = async (page: number) => {
        try {
            setLoading(true);
            setError("");
            const fetchedUsers = await getUsers(page, pageSize);
            $Users.set(fetchedUsers);
        } catch (err: any) {
            setError(err?.message || "Failed to fetch users");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        // Redirect to page 1 if no page or invalid page
        if (!page || isNaN(currentPage) || currentPage < 1) {
            navigate("/users/1", { replace: true });
            return;
        }
        fetchUsers(currentPage);
    }, [page, currentPage, navigate]);

    const handlePreviousPage = () => {
        if (currentPage > 1) {
            navigate(`/users/${currentPage - 1}`);
        }
    };

    const handleNextPage = () => {
        if (users.length === pageSize) {
            navigate(`/users/${currentPage + 1}`);
        }
    };

    if (loading) {
        return (
            <div className="p-6">
                <h1 className="text-2xl font-semibold mb-6">Users</h1>
                <div className="text-center py-8">Loading users...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-6">
                <h1 className="text-2xl font-semibold mb-6">Users</h1>
                <div className="text-red-600 text-center py-8">Error: {error}</div>
            </div>
        );
    }

    return (
        <div className="p-6">
            <h1 className="text-2xl font-semibold mb-6">Users</h1>
            
            {users.length === 0 ? (
                <div className="text-gray-500 text-center py-8">No users found</div>
            ) : (
                <>
                    {/* User List */}
                    <div className="space-y-4">
                        {users.map((user) => (
                            <div key={user.id} className="border rounded-lg p-4 bg-white shadow-sm">
                                <div className="flex justify-between items-start">
                                    <div className="flex-1">
                                        <h3 className="font-medium text-lg">
                                            {user.name}
                                        </h3>
                                        <p className="text-sm text-gray-500 mt-1">
                                            Email: {user.email}
                                        </p>
                                        <p className="text-sm text-gray-500">
                                            User ID: {user.id}
                                        </p>
                                        <p className="text-sm text-gray-500">
                                            Roles: {user.roles || 'No roles assigned'}
                                        </p>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                    
                    {/* Pagination */}
                    <div className="flex justify-center items-center mt-8 space-x-2">
                        <button 
                            onClick={handlePreviousPage}
                            disabled={currentPage === 1}
                            className="px-3 py-2 border rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            &lt;
                        </button>
                        <span className="px-3 py-2 bg-blue-500 text-white rounded">
                            {currentPage}
                        </span>
                        <button 
                            onClick={handleNextPage}
                            disabled={users.length < pageSize}
                            className="px-3 py-2 border rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            &gt;
                        </button>
                    </div>
                </>
            )}
        </div>
    );
}