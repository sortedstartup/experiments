import { useEffect, useState, useRef } from "react";
import { useStore } from "@nanostores/react";
import { useParams, useNavigate } from "react-router-dom";
import {
    $Tenants,
    $TenantUsers,
    $TenantUsersLoading,
    $TenantUsersError,
    $TenantUsersCurrentPage,
    $TenantUsersCanGoNext,
    $TenantUsersCanGoPrev,
    $SelectedTenant,
    fetchTenants,
    goToTenantUsersPage,
    tenantUsersNextPage,
    tenantUsersPrevPage,
    setSelectedTenant,
    removeUserFromTenant,
} from "../store/admin";
import AddUserModal from "../components/AddUserModal";
import UserActionsMenu from "../components/UserActionsMenu";

export default function TenantUsers() {
    const tenants = useStore($Tenants);
    const tenantUsers = useStore($TenantUsers);
    const loading = useStore($TenantUsersLoading);
    const error = useStore($TenantUsersError);
    const currentPage = useStore($TenantUsersCurrentPage);
    const canGoNext = useStore($TenantUsersCanGoNext);
    const canGoPrev = useStore($TenantUsersCanGoPrev);
    const selectedTenant = useStore($SelectedTenant);
    
    const [showAddUserModal, setShowAddUserModal] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [openMenuUserId, setOpenMenuUserId] = useState<string | null>(null);
    const buttonRefs = useRef<Map<string, HTMLButtonElement>>(new Map());
    
    const { page } = useParams<{ page: string }>();
    const navigate = useNavigate();
    const urlPage = parseInt(page || "1", 10);

    const setButtonRef = (userId: string, element: HTMLButtonElement | null) => {
        if (element) {
            buttonRefs.current.set(userId, element);
        } else {
            buttonRefs.current.delete(userId);
        }
    };

    const getButtonRef = (userId: string): React.RefObject<HTMLButtonElement> => {
        return { current: buttonRefs.current.get(userId) || null } as React.RefObject<HTMLButtonElement>;
    };

    // Fetch tenants on mount
    useEffect(() => {
        fetchTenants();
    }, []);

    // Redirect to page 1 if no page
    useEffect(() => {
        if (!page || isNaN(urlPage) || urlPage < 1) {
            navigate("/tenant-users/1", { replace: true });
        }
    }, [page, urlPage, navigate]);

    // Fetch tenant users when selected tenant or page changes
    useEffect(() => {
        if (selectedTenant && page && !isNaN(urlPage) && urlPage >= 1) {
            goToTenantUsersPage(urlPage);
        }
    }, [selectedTenant, page, urlPage]);

    const handleTenantChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        navigate("/tenant-users/1");
        setSelectedTenant(e.target.value);
    };

    const handlePreviousPage = () => {
        if (canGoPrev) {
            navigate(`/tenant-users/${currentPage - 1}`);
            tenantUsersPrevPage();
        }
    };

    const handleNextPage = () => {
        if (canGoNext) {
            navigate(`/tenant-users/${currentPage + 1}`);
            tenantUsersNextPage();
        }
    };

    const handleRemoveUser = async (userId: string) => {
        if (!selectedTenant) return;
        
        setOpenMenuUserId(null); // Close menu

        try {
            await removeUserFromTenant(selectedTenant, userId);
        } catch (err: any) {
            alert(err?.message || "Failed to remove user");
        }
    };

    const handleChangeRole = (_userId: string) => {
        setOpenMenuUserId(null); // Close menu
        // TODO: Implement role change modal
        console.log("Change role clicked for user:", _userId);
    };

    const toggleMenu = (userId: string, event: React.MouseEvent) => {
        event.stopPropagation();
        setOpenMenuUserId(openMenuUserId === userId ? null : userId);
    };

    const closeMenu = () => {
        setOpenMenuUserId(null);
    };

    const selectedTenantObj = tenants.find(t => t.id === selectedTenant);

    if (loading && tenants.length === 0) {
        return (
            <div className="flex-1 bg-gray-950 text-white">
                <div className="p-8">
                    <div className="flex items-center justify-between mb-6">
                        <div>
                            <h1 className="text-3xl font-semibold mb-2">Tenant Users</h1>
                            <p className="text-gray-400">Manage users within a specific tenant</p>
                        </div>
                    </div>
                    <div className="text-center py-8 text-gray-400">Loading...</div>
                </div>
            </div>
        );
    }

    return (
        <div className="flex-1 bg-gray-950 text-white">
            <div className="p-8">
                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                    <div>
                        <h1 className="text-3xl font-semibold mb-2">Tenant Users</h1>
                        <p className="text-gray-400">Manage users within a specific tenant</p>
                    </div>
                    <button
                        onClick={() => setShowAddUserModal(true)}
                        disabled={!selectedTenant}
                        className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                        </svg>
                        Add User
                    </button>
                </div>

                {/* Filters */}
                <div className="flex gap-4 mb-6">
                    {/* Tenant Dropdown */}
                    <div className="relative flex-1 max-w-xs">
                        <select
                            value={selectedTenant}
                            onChange={handleTenantChange}
                            className="w-full bg-gray-900 border border-gray-800 rounded-lg px-4 py-2.5 text-white appearance-none cursor-pointer focus:outline-none focus:border-blue-500"
                        >
                            {tenants.map((tenant) => (
                                <option key={tenant.id} value={tenant.id}>
                                    {tenant.name}
                                </option>
                            ))}
                        </select>
                        <svg 
                            className="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400 pointer-events-none"
                            fill="none" 
                            stroke="currentColor" 
                            viewBox="0 0 24 24"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                        </svg>
                    </div>

                    {/* Search */}
                    <div className="relative flex-1 max-w-md">
                        <svg 
                            className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400"
                            fill="none" 
                            stroke="currentColor" 
                            viewBox="0 0 24 24"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                        <input
                            type="text"
                            placeholder="Search users in tenant..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full bg-gray-900 border border-gray-800 rounded-lg pl-10 pr-4 py-2.5 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
                        />
                    </div>
                </div>

                {error && (
                    <div className="bg-red-900/20 border border-red-800 text-red-400 px-4 py-3 rounded-lg mb-6">
                        Error: {error}
                    </div>
                )}

                {/* User Table */}
                <div className="bg-gray-900 rounded-lg border border-gray-800">
                    <div className="overflow-x-auto">
                        <table className="w-full">
                        <thead className="bg-gray-800/50 border-b border-gray-800">
                            <tr>
                                <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Name</th>
                                <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Email</th>
                                <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Role</th>
                                <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Added</th>
                                <th className="px-6 py-4 text-right text-sm font-medium text-gray-300">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-800">
                            {loading && selectedTenant ? (
                                <tr>
                                    <td colSpan={5} className="px-6 py-8 text-center text-gray-400">
                                        Loading users...
                                    </td>
                                </tr>
                            ) : tenantUsers.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-6 py-8 text-center text-gray-400">
                                        No users found in this tenant
                                    </td>
                                </tr>
                            ) : (
                                tenantUsers.map((user) => (
                                    <tr key={user.id} className="hover:bg-gray-800/50">
                                        <td className="px-6 py-4 text-white">{user.name}</td>
                                        <td className="px-6 py-4 text-gray-300">{user.email}</td>
                                        <td className="px-6 py-4">
                                            {user.email.includes('john') ? (
                                                <span className="inline-flex px-3 py-1 rounded-full text-xs font-medium bg-blue-900/30 text-blue-400 border border-blue-800">
                                                    admin
                                                </span>
                                            ) : (
                                                <span className="inline-flex px-3 py-1 rounded-full text-xs font-medium bg-gray-800/50 text-gray-300 border border-gray-700">
                                                    user
                                                </span>
                                            )}
                                        </td>
                                        <td className="px-6 py-4 text-gray-400">
                                            {user.email.includes('john') ? '2024-01-15' : '2024-02-10'}
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="flex justify-end">
                                                <button 
                                                    ref={(el) => setButtonRef(user.id, el)}
                                                    onClick={(e) => toggleMenu(user.id, e)}
                                                    className="text-gray-400 hover:text-white p-1 rounded hover:bg-gray-800"
                                                >
                                                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                                                    </svg>
                                                </button>
                                            </div>
                                            
                                            <UserActionsMenu
                                                userId={user.id}
                                                isOpen={openMenuUserId === user.id}
                                                onClose={closeMenu}
                                                onChangeRole={() => handleChangeRole(user.id)}
                                                onRemove={() => handleRemoveUser(user.id)}
                                                buttonRef={getButtonRef(user.id)}
                                            />
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                    </div>
                </div>

                {/* Pagination */}
                {tenantUsers.length > 0 && (
                    <div className="flex justify-center items-center space-x-2 mt-6">
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
                )}
            </div>

            {/* Add User Modal */}
            {showAddUserModal && selectedTenant && selectedTenantObj && (
                <AddUserModal
                    tenantId={selectedTenant}
                    tenantName={selectedTenantObj.name}
                    onClose={() => setShowAddUserModal(false)}
                />
            )}
        </div>
    );
}

