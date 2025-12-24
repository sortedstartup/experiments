import { useEffect, useState } from "react";
import { 
    $Tenants, 
    $TenantsLoading, 
    $TenantsError, 
    $TenantsCurrentPage,
    $TenantsCanGoNext,
    $TenantsCanGoPrev,
    goToTenantsPage,
    tenantsNextPage,
    tenantsPrevPage,
    createTenant
} from "../store/admin";
import { useStore } from "@nanostores/react";
import { useParams, useNavigate } from "react-router-dom";

export default function Tenants() {
    const tenants = useStore($Tenants);
    const loading = useStore($TenantsLoading);
    const error = useStore($TenantsError);
    const currentPage = useStore($TenantsCurrentPage);
    const canGoNext = useStore($TenantsCanGoNext);
    const canGoPrev = useStore($TenantsCanGoPrev);
    
    const { page } = useParams<{ page: string }>();
    const navigate = useNavigate();
    const urlPage = parseInt(page || "1", 10);

    // Dialog state
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [tenantName, setTenantName] = useState("");
    const [tenantDomain, setTenantDomain] = useState("");
    const [tenantStatus, setTenantStatus] = useState("Active");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [submitError, setSubmitError] = useState("");

    useEffect(() => {
        // Redirect to page 1 if no page or invalid page
        if (!page || isNaN(urlPage) || urlPage < 1) {
            navigate("/tenants/1", { replace: true });
            return;
        }
        
        // Always fetch on mount or page change
        goToTenantsPage(urlPage);
    }, [page, urlPage, navigate]);

    const handlePreviousPage = () => {
        if (canGoPrev) {
            navigate(`/tenants/${currentPage - 1}`);
            tenantsPrevPage();
        }
    };

    const handleNextPage = () => {
        if (canGoNext) {
            navigate(`/tenants/${currentPage + 1}`);
            tenantsNextPage();
        }
    };

    const handleCreateTenant = async (e: React.FormEvent) => {
        e.preventDefault();
        setSubmitError("");
        setIsSubmitting(true);

        try {
            await createTenant(tenantName, "", tenantDomain);
            // Reset form and close dialog
            setTenantName("");
            setTenantDomain("");
            setTenantStatus("Active");
            setIsDialogOpen(false);
        } catch (err: any) {
            setSubmitError(err?.message || "Failed to create tenant");
        } finally {
            setIsSubmitting(false);
        }
    };

    if (loading) {
        return (
            <div className="flex-1 bg-gray-950 text-white">
                <div className="p-8">
                    <h1 className="text-3xl font-semibold mb-2">All Tenants</h1>
                    <p className="text-gray-400 mb-6">View all tenants in the system</p>
                    <div className="text-center py-8 text-gray-400">Loading tenants...</div>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex-1 bg-gray-950 text-white">
                <div className="p-8">
                    <h1 className="text-3xl font-semibold mb-2">All Tenants</h1>
                    <p className="text-gray-400 mb-6">View all tenants in the system</p>
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
                <div className="flex justify-between items-center mb-2">
                    <h1 className="text-3xl font-semibold">All Tenants</h1>
                    <button
                        onClick={() => setIsDialogOpen(true)}
                        className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors"
                    >
                        Add Tenant
                    </button>
                </div>
                <p className="text-gray-400 mb-6">View all tenants in the system</p>
                
                {tenants.length === 0 ? (
                    <div className="text-gray-400 text-center py-8">No tenants found</div>
                ) : (
                    <>
                        {/* Tenant Table */}
                        <div className="bg-gray-900 rounded-lg border border-gray-800 overflow-hidden mb-6">
                            <table className="w-full">
                                <thead className="bg-gray-800/50 border-b border-gray-800">
                                    <tr>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Tenant ID</th>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Name</th>
                                        <th className="px-6 py-4 text-left text-sm font-medium text-gray-300">Description</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-800">
                                    {tenants.map((tenant) => (
                                        <tr key={tenant.id} className="hover:bg-gray-800/50">
                                            <td className="px-6 py-4 text-gray-400">{tenant.id}</td>
                                            <td className="px-6 py-4 text-white">{tenant.name}</td>
                                            <td className="px-6 py-4 text-gray-300">{tenant.description}</td>
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

            {/* Add Tenant Dialog */}
            {isDialogOpen && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-gray-900 rounded-lg border border-gray-800 w-full max-w-md mx-4">
                        <div className="p-6">
                            <div className="flex justify-between items-center mb-4">
                                <div>
                                    <h2 className="text-xl font-semibold text-white">Add New Tenant</h2>
                                    <p className="text-sm text-gray-400 mt-1">Create a new tenant organization in the system</p>
                                </div>
                                <button
                                    onClick={() => {
                                        setIsDialogOpen(false);
                                        setSubmitError("");
                                    }}
                                    className="text-gray-400 hover:text-white"
                                >
                                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </button>
                            </div>

                            <form onSubmit={handleCreateTenant} className="space-y-4">
                                {submitError && (
                                    <div className="bg-red-900/20 border border-red-800 text-red-400 px-4 py-3 rounded-lg text-sm">
                                        {submitError}
                                    </div>
                                )}

                                <div>
                                    <label htmlFor="tenantName" className="block text-sm font-medium text-gray-300 mb-2">
                                        Tenant Name
                                    </label>
                                    <input
                                        id="tenantName"
                                        type="text"
                                        value={tenantName}
                                        onChange={(e) => setTenantName(e.target.value)}
                                        className="w-full px-3 py-2 bg-gray-950 border border-gray-800 rounded-lg text-white focus:outline-none focus:border-blue-600"
                                        placeholder="s"
                                        required
                                    />
                                </div>

                                <div>
                                    <label htmlFor="tenantDomain" className="block text-sm font-medium text-gray-300 mb-2">
                                        Domain
                                    </label>
                                    <input
                                        id="tenantDomain"
                                        type="text"
                                        value={tenantDomain}
                                        onChange={(e) => setTenantDomain(e.target.value)}
                                        className="w-full px-3 py-2 bg-gray-950 border border-gray-800 rounded-lg text-white focus:outline-none focus:border-blue-600"
                                        placeholder="e.g., acme.com"
                                    />
                                </div>

                                <div>
                                    <label htmlFor="tenantStatus" className="block text-sm font-medium text-gray-300 mb-2">
                                        Status
                                    </label>
                                    <select
                                        id="tenantStatus"
                                        value={tenantStatus}
                                        onChange={(e) => setTenantStatus(e.target.value)}
                                        className="w-full px-3 py-2 bg-gray-950 border border-gray-800 rounded-lg text-white focus:outline-none focus:border-blue-600"
                                    >
                                        <option value="Active">Active</option>
                                        <option value="Inactive">Inactive</option>
                                    </select>
                                </div>

                                <div className="flex justify-end space-x-3 pt-4">
                                    <button
                                        type="button"
                                        onClick={() => {
                                            setIsDialogOpen(false);
                                            setSubmitError("");
                                        }}
                                        className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-lg font-medium transition-colors"
                                        disabled={isSubmitting}
                                    >
                                        Cancel
                                    </button>
                                    <button
                                        type="submit"
                                        className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                        disabled={isSubmitting}
                                    >
                                        {isSubmitting ? "Creating..." : "Create Tenant"}
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

