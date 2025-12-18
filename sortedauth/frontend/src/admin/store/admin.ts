import { atom } from "nanostores";
import { User, UserServiceClient, UsersListRequest} from "../../../proto/authservice";



const client = new UserServiceClient("http://localhost:8080", {});

export const $Users = atom<User[]>([]);

export const getUsers = async (page: number, pageSize: number) => {
    const response = await client.UsersList(UsersListRequest.fromObject({
        page: page,
        page_size: pageSize,
    }), null);
    return response.users;
};