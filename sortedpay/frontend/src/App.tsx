import { createBrowserRouter, RouterProvider } from "react-router-dom";
import CreateProduct from "./payment/CreateProduct";
import ListProducts from "./payment/ListProducts";
import Success from "./payment/Success";
import Cancel from "./payment/Cancel";

// Dummy Home Page
function Home() {
  return <div>payment</div>;
}

const router = createBrowserRouter([
  {
    path: "/",
    element: <Home />,
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