import { atom, computed } from "nanostores";
import { 
    User, 
    UserServiceClient, 
    UsersListRequest,
    Tenant,
    TenantServiceClient,
    TenantsListRequest,
    CreateTenantRequest,
    AddUserRequest,
    RemoveUserRequest
} from "../../../proto/authservice";

const userClient = new UserServiceClient("http://localhost:8080", {});
const tenantClient = new TenantServiceClient("http://localhost:8080", {});

const PAGE_SIZE = 10;

// Data Stores
export const $Users = atom<User[]>([]);
export const $Tenants = atom<Tenant[]>([]);
export const $TenantUsers = atom<User[]>([]);
export const $SelectedTenant = atom<string>("");

// Pagination Stores
export const $UsersCurrentPage = atom<number>(1);
export const $TenantsCurrentPage = atom<number>(1);
export const $TenantUsersCurrentPage = atom<number>(1);

// Loading/Error Stores
export const $UsersLoading = atom<boolean>(false);
export const $TenantsLoading = atom<boolean>(false);
export const $TenantUsersLoading = atom<boolean>(false);

export const $UsersError = atom<string>("");
export const $TenantsError = atom<string>("");
export const $TenantUsersError = atom<string>("");

// Computed - Can go to next page if current page has full page size
export const $UsersCanGoNext = computed($Users, (users) => users.length === PAGE_SIZE);
export const $TenantsCanGoNext = computed($Tenants, (tenants) => tenants.length === PAGE_SIZE);
export const $TenantUsersCanGoNext = computed($TenantUsers, (users) => users.length === PAGE_SIZE);

export const $UsersCanGoPrev = computed($UsersCurrentPage, (page) => page > 1);
export const $TenantsCanGoPrev = computed($TenantsCurrentPage, (page) => page > 1);
export const $TenantUsersCanGoPrev = computed($TenantUsersCurrentPage, (page) => page > 1);

// User operations with pagination
export const fetchUsers = async (page?: number) => {
    const currentPage = page ?? $UsersCurrentPage.get();
    
    try {
        $UsersLoading.set(true);
        $UsersError.set("");
        
        const response = await userClient.UsersList(UsersListRequest.fromObject({
            page_request: {
                page: currentPage,
                page_size: PAGE_SIZE,
            }
        }), null);
        
        $Users.set(response.users);
        $UsersCurrentPage.set(currentPage);
    } catch (err: any) {
        $UsersError.set(err?.message || "Failed to fetch users");
        $Users.set([]);
    } finally {
        $UsersLoading.set(false);
    }
};

export const goToUsersPage = (page: number) => {
    if (page < 1) return;
    fetchUsers(page);
};

export const usersNextPage = () => {
    if ($UsersCanGoNext.get()) {
        fetchUsers($UsersCurrentPage.get() + 1);
    }
};

export const usersPrevPage = () => {
    if ($UsersCanGoPrev.get()) {
        fetchUsers($UsersCurrentPage.get() - 1);
    }
};

// Tenant operations with pagination
export const fetchTenants = async (page?: number) => {
    const currentPage = page ?? $TenantsCurrentPage.get();
    
    try {
        $TenantsLoading.set(true);
        $TenantsError.set("");
        
        const response = await tenantClient.TenantsList(TenantsListRequest.fromObject({
            page_request: {
                page: currentPage,
                page_size: PAGE_SIZE,
            }
        }), null);
        
        $Tenants.set(response.tenants);
        $TenantsCurrentPage.set(currentPage);
        
        // Set first tenant as selected if none selected
        if (response.tenants.length > 0 && !$SelectedTenant.get()) {
            $SelectedTenant.set(response.tenants[0].id);
        }
    } catch (err: any) {
        $TenantsError.set(err?.message || "Failed to fetch tenants");
        $Tenants.set([]);
    } finally {
        $TenantsLoading.set(false);
    }
};

export const goToTenantsPage = (page: number) => {
    if (page < 1) return;
    fetchTenants(page);
};

export const tenantsNextPage = () => {
    if ($TenantsCanGoNext.get()) {
        fetchTenants($TenantsCurrentPage.get() + 1);
    }
};

export const tenantsPrevPage = () => {
    if ($TenantsCanGoPrev.get()) {
        fetchTenants($TenantsCurrentPage.get() - 1);
    }
};

// Tenant Users operations with pagination
export const fetchTenantUsers = async (tenantId?: string, page?: number) => {
    const selectedTenant = tenantId ?? $SelectedTenant.get();
    const currentPage = page ?? $TenantUsersCurrentPage.get();
    
    if (!selectedTenant) return;
    
    try {
        $TenantUsersLoading.set(true);
        $TenantUsersError.set("");
        
        const response = await userClient.UsersList(UsersListRequest.fromObject({
            filters: {
                tenant_id: selectedTenant
            },
            page_request: {
                page: currentPage,
                page_size: PAGE_SIZE,
            }
        }), null);
        
        $TenantUsers.set(response.users);
        $TenantUsersCurrentPage.set(currentPage);
    } catch (err: any) {
        $TenantUsersError.set(err?.message || "Failed to fetch tenant users");
        $TenantUsers.set([]);
    } finally {
        $TenantUsersLoading.set(false);
    }
};

export const goToTenantUsersPage = (page: number) => {
    if (page < 1) return;
    fetchTenantUsers(undefined, page);
};

export const tenantUsersNextPage = () => {
    if ($TenantUsersCanGoNext.get()) {
        fetchTenantUsers(undefined, $TenantUsersCurrentPage.get() + 1);
    }
};

export const tenantUsersPrevPage = () => {
    if ($TenantUsersCanGoPrev.get()) {
        fetchTenantUsers(undefined, $TenantUsersCurrentPage.get() - 1);
    }
};

export const setSelectedTenant = (tenantId: string) => {
    $SelectedTenant.set(tenantId);
    $TenantUsersCurrentPage.set(1); // Reset to page 1 when changing tenant
    fetchTenantUsers(tenantId, 1);
};

// Tenant CRUD operations
export const createTenant = async (name: string, description: string, domain: string) => {
    const response = await tenantClient.CreateTenant(CreateTenantRequest.fromObject({
        name: name,
        description: description,
    }), null);
    
    // Refresh tenants list to show the new tenant
    await fetchTenants($TenantsCurrentPage.get());
    
    return response;
};

// Add/Remove user operations
export const addUserToTenant = async (tenantId: string, email: string, roleId: string = "user") => {
    const response = await tenantClient.AddUser(AddUserRequest.fromObject({
        tenant_id: tenantId,
        email: email,
        role_id: roleId,
    }), null);
    
    // Refresh tenant users list
    await fetchTenantUsers(tenantId, $TenantUsersCurrentPage.get());
    
    return response.tenant_user;
};

export const removeUserFromTenant = async (tenantId: string, userId: string) => {
    const response = await tenantClient.RemoveUser(RemoveUserRequest.fromObject({
        tenant_id: tenantId,
        user_id: userId,
    }), null);
    
    // Refresh tenant users list
    await fetchTenantUsers(tenantId, $TenantUsersCurrentPage.get());
    
    return response;
};