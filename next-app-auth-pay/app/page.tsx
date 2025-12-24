'use client';
import { useEffect, useState } from "react";
import { useStore } from "@nanostores/react";
import { useRouter } from "next/navigation";
import { $LoggedInUser, isLoggedIn } from "./store/auth";
import { ProductList, listProducts, checkUserProductAccess } from "./store/payment";

export default function Page() {
  const router = useRouter();
  const loggedInUser = useStore($LoggedInUser);
  const products = useStore(ProductList);
  const [productAccess, setProductAccess] = useState<Record<string, boolean>>({});

  useEffect(() => {
    if (!isLoggedIn()) {
      router.push('/login');
      return;
    }
    listProducts();
  }, []);

  useEffect(() => {
    if (products.length > 0) {
      products.forEach(async (product) => {
        const hasAccess = await checkUserProductAccess(product.id);
        setProductAccess(prev => ({ ...prev, [product.id]: hasAccess }));
      });
    }
  }, [products]);

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">Products</h1>
      <p className="mb-6">Logged in: {loggedInUser.user?.email || 'No user'}</p>
      
      <div className="space-y-4">
        {products.map((product) => (
          <div key={product.id} className="border p-4 rounded flex justify-between items-center">
            <div>
              <h3 className="font-semibold">{product.name}</h3>
              <p className="text-sm text-gray-600">{product.description}</p>
            </div>
            {productAccess[product.id] ? (
              <button className="bg-green-500 text-white px-4 py-2 rounded">
                Access Now
              </button>
            ) : (
              <button className="bg-blue-500 text-white px-4 py-2 rounded">
                Buy
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
