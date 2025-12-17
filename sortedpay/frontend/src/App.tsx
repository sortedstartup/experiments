import { createBrowserRouter, RouterProvider } from "react-router-dom";
import CreateProduct from "./payment/CreateProduct";
import ListProducts from "./payment/ListProducts";
import Transactions from "./payment/Transactions";
import Success from "./payment/Success";
import Cancel from "./payment/Cancel";
import Layout from "./Layout";
import Dashboard from "./payment/Dashboard";


const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    children: [
      {
        path: "/",
        element: <Dashboard />,
      },
      {
        path: "/create-product",
        element: <CreateProduct />,
      },
      {
        path: "/list-products",
        element: <ListProducts />,
      },
      {
        path: "/transactions/:page",
        element: <Transactions />,
      },
    ],
  },
  {
    path: "/success",
    element: <Success />,
  },
  {
    path: "/cancel",
    element: <Cancel />,
  },
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;