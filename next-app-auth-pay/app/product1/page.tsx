import { cookies } from "next/headers";
import { checkUserProductAccess } from "../store/payment";
import { redirect } from "next/navigation";

export default async function Page() {
  // ---- Get JWT from cookies (browser-based auth) ----
  const cookieStore = await cookies();
  const jwtToken = cookieStore.get("sortedchat.jwt")?.value;

  if (!jwtToken) {
    // Redirect to login or show unauthorized message
    return (
      <div>
        <h1>Unauthorized</h1>
        <p>No authorization token found. Please log in.</p>
      </div>
    );
  }

  // ---- Call backend REST API ----
  // const res = await fetch("http://localhost:8081/check-user-product", {
  //   method: "GET",
  //   headers: {
  //     Authorization: `Bearer ${jwtFromHeader}`,
  //   },
  //   cache: "no-store", // VERY IMPORTANT for SSR
  // });

  const hasAccess = await checkUserProductAccess("64c88ae5-9377-413d-b1e2-ba6bb0e63b1a");
  console.log("hasAccess", hasAccess);

  if (!hasAccess) {
    return (
      <div>
        <h1>Access Denied</h1>
        <p>You do not have access to this product.</p>
      </div>
    );
  }

  return (
    <div>
      <h1>Product 1</h1>
      <p>Welcome! You have access to this product.</p>
    </div>
  );
}
