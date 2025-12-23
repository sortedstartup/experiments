'use client';

import { useEffect, useState } from "react";
import { listProducts, ProductList, createStripeCheckoutSession, checkUserProductAccess } from "./payment";
import { useStore } from "@nanostores/react";


export default function Home() {
  const products = useStore(ProductList);
  const [loadingProductId, setLoadingProductId] = useState<string | null>(null);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    listProducts();
  }, []);

  const handleBuyNow = async (product: any) => {
    try {
      setLoadingProductId(product.id);
      setError("");
      const sessionUrl = await createStripeCheckoutSession(product.id);
      // Redirect to Stripe checkout
      window.location.href = sessionUrl;
    } catch (err: any) {
      console.error("Failed to create checkout session:", err);
      setError(err?.message || "Failed to create checkout session");
    } finally {
      setLoadingProductId(null);
    }
  };

  const hasAccess = async (productId: string) => {
    try {
      const hasAccess = await checkUserProductAccess(productId);
      console.log("hasAccess", hasAccess);
      return hasAccess;
    } catch (err: any) {
      console.error("Failed to check user product access:", err);
      return false;
    }
  }

  return (
    <div className="container mx-auto p-6">
      <h1 className="text-3xl font-bold mb-8 text-center">Products</h1>
      
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-6">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {products.map((product: any) => (
          <div key={product.id} className="bg-white rounded-lg shadow-lg border border-gray-200 p-6 hover:shadow-xl transition-shadow">
            <div className="mb-4">
              <h2 className="text-xl font-semibold text-gray-800 mb-2">{product.name}</h2>
              <p className="text-gray-600 text-sm mb-4">{product.description}</p>
            </div>

            <div className="space-y-2 mb-6">
              <div className="flex justify-between text-sm">
                <span className="text-gray-500">Price:</span>
                <span className="font-medium">
                  {product.currency === 0 ? '$' : '₹'}{(product.amount_in_smallest_unit / 100).toFixed(2)}
                </span>
              </div>
              
              {product.is_recurring && (
                <div className="flex justify-between text-sm">
                  <span className="text-gray-500">Billing:</span>
                  <span className="font-medium">
                    Every {product.interval_count} {product.interval_period}(s)
                  </span>
                </div>
              )}

              <div className="flex justify-between text-sm">
                <span className="text-gray-500">Type:</span>
                <span className="font-medium">
                  {product.is_recurring ? 'Subscription' : 'One-time'}
                </span>
              </div>

              <div className="flex justify-between text-sm">
                <span className="text-gray-500">Access:</span>
                <span className={`font-medium ${product.has_access ? 'text-green-600' : 'text-gray-600'}`}>
                  {product.has_access ? 'You have access' : 'No access'}
                </span>
              </div>
            </div>

            <button
              onClick={() => handleBuyNow(product)}
              disabled={loadingProductId === product.id || product.has_access}
              className={`w-full py-3 px-4 rounded-lg font-medium transition-colors ${
                product.has_access
                  ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                  : loadingProductId === product.id
                  ? 'bg-blue-400 text-white cursor-not-allowed'
                  : 'bg-blue-600 hover:bg-blue-700 text-white'
              }`}
            >
              {loadingProductId === product.id ? (
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Processing...
                </div>
              ) : product.has_access ? (
                'Already Purchased'
              ) : (
                `Buy Now - ${product.currency === 0 ? '$' : '₹'}{(product.amount_in_smallest_unit / 100).toFixed(2)}`
              )}
            </button>

            <div className="mt-4 pt-4 border-t border-gray-100">
              <p className="text-xs text-gray-400">Product ID: {product.id}</p>
              <p className="text-xs text-gray-400">Stripe ID: {product.stripe_product_id}</p>
            </div>
          </div>
        ))}
      </div>

      <div className="flex justify-center">
        <button onClick={() => hasAccess(products[0].id)} className="bg-blue-500 text-white px-4 py-2 rounded-md">Check Access</button>
      </div>

      {products.length === 0 && (
        <div className="text-center py-12">
          <p className="text-gray-500 text-lg">No products available</p>
        </div>
      )}
    </div>
  )
}