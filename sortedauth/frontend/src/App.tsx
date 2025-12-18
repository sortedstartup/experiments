import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Layout from "./Layout";
import Admin from "./admin/pages/Admin";


const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    children: [
      {
        path: "/",
        element: <Admin />,
      },
      {
        path: "/users/:page",
        element: <Admin />,
      }
    ],
  },
]);

function App() {

  return (
    <>
       <RouterProvider router={router} />;
    </>
  )
}

export default App
