import { $LoggedInUser, Logout } from "../store/auth";
import {useStore} from "@nanostores/react";


export function Signin() {
   const { isLoggedIn, user } = useStore($LoggedInUser);
   if (isLoggedIn) {
    return (
        <>
            <div>Welcome {user?.name} <br />
            <button onClick={Logout}>Logout</button></div>
        </>
    )
   }
    return (
        <>
            <a href="/login">Login</a>
        </>
    )
}

