import { headers, cookies } from "next/headers";
import { checkUserProductAccess } from "./store/payment";

export default async function Page() {
  
  return (
    <div>
      <h1>Page</h1>
    </div>
  )

}
